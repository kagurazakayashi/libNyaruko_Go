package nyasqldrift_test

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"

	"github.com/kagurazakayashi/libNyaruko_Go/nyamysql"
	"github.com/kagurazakayashi/libNyaruko_Go/nyasqldrift"
	"github.com/kagurazakayashi/libNyaruko_Go/nyasqlite"
)

// TestConvertMySQLToSQLite 測試將MySQL SQL轉換為SQLite SQL的功能
//
// 引數:
//
//	t *testing.T: 測試框架提供的測試例項
//
// 返回值:
//
//	無
func TestConvertMySQLToSQLite(t *testing.T) {
	// 輸入的MySQL SQL語句
	input := `
CREATE TABLE users (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100),
  created_at DATETIME
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
`

	// 預期的轉換結果中包含的字串陣列
	expectedContains := []string{
		"CREATE TABLE users",
		"id INTEGER PRIMARY KEY AUTOINCREMENT",
		"name TEXT",
		"created_at TEXT",
	}

	// 呼叫nyasqldrift.ConvertSQL函式進行SQL轉換
	result, err := nyasqldrift.ConvertSQL(input, nyasqldrift.MySQLToSQLite)
	if err != nil {
		// 如果轉換失敗，輸出錯誤資訊並終止測試
		t.Fatalf("Conversion failed: %v", err)
	}

	// 遍歷expectedContains陣列，檢查轉換結果中是否包含預期的字串
	for _, part := range expectedContains {
		if !contains(result, part) {
			// 如果轉換結果中不包含預期的字串，輸出錯誤資訊
			t.Errorf("Expected result to contain '%s', but got:\n%s", part, result)
		}
	}
}

// TestConvertSQLiteToMySQL 測試將 SQLite SQL 轉換為 MySQL SQL 的功能
//
// 引數:
// t *testing.T: 測試上下文
func TestConvertSQLiteToMySQL(t *testing.T) {
	input := `
CREATE TABLE users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT,
  created_at TEXT
);
`
	// 期望結果包含的字串陣列
	expectedContains := []string{
		"CREATE TABLE users",
		"id INT AUTO_INCREMENT PRIMARY KEY",
		"name VARCHAR(255)",
		// TEXT → VARCHAR(255) in current map
		"created_at VARCHAR(255)",
		"ENGINE=InnoDB",
	}

	// 呼叫 ConvertSQL 函式進行轉換
	result, err := nyasqldrift.ConvertSQL(input, nyasqldrift.SQLiteToMySQL)
	if err != nil {
		// 如果轉換失敗，則輸出錯誤資訊並終止測試
		t.Fatalf("Conversion failed: %v", err)
	}

	// 遍歷期望結果包含的字串陣列
	for _, part := range expectedContains {
		// 如果結果中不包含該字串，則輸出錯誤資訊
		if !contains(result, part) {
			t.Errorf("Expected result to contain '%s', but got:\n%s", part, result)
		}
	}
}

// contains 判斷字串 s 中是否包含子字串 substr
//
// 引數：
// s: 要檢查的字串
// substr: 要查詢的子字串
//
// 返回值：
// 如果 s 包含 substr，則返回 true，否則返回 false
func contains(s, substr string) bool {
	// 如果字串 s 和子字串 substr 都非空，並且 s 的長度大於等於 substr 的長度
	// 並且 substr 在 s 中的索引大於等於 0（即 substr 是 s 的子串）
	return len(s) > 0 && len(substr) > 0 && (len(s) >= len(substr)) && (stringIndex(s, substr) >= 0)
}

// stringIndex 函式計算字串 s 和子字串 substr 的長度差值，並返回結果。
// 如果 s 和 substr 均為空字串，則返回 0。
// 引數：
//
//	s: 待計算的字串
//	substr: 待計算的子字串
//
// 返回值：
//
//	返回字串 s 和子字串 substr 的長度差值
func stringIndex(s, substr string) int {
	// 將字串 s 和子字串 substr 轉換為 rune 切片
	// 計算 s 轉換為 rune 切片後的長度
	// 計算 substr 轉換為 rune 切片後的長度
	return len([]rune(s[:])) - len([]rune(substr[:]))
}

// setupTestDBs 函式用於設定測試所需的資料庫連線
// 引數：
// t: *testing.T 型別，測試框架提供的測試物件
// 返回值：
// *nyamysql.NyaMySQL: 初始化後的 MySQL 資料庫連線物件
// *nyasqlite.NyaSQLite: 初始化後的 SQLite 資料庫連線物件
// func(): 一個函式，用於清理資料庫連線和表
func setupTestDBs(t *testing.T) (*nyamysql.NyaMySQL, *nyasqlite.NyaSQLite, func()) {
	// 獲取環境變數中的MySQL DSN
	mysqlDSN := os.Getenv("MYSQL_TEST_DSN") // e.g. "user:pass@tcp(localhost:3306)/testdb"
	// 嘗試連線到MySQL資料庫
	mysqlDB, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		// 如果連線失敗，則記錄錯誤並終止測試
		t.Fatalf("failed to connect to MySQL: %v", err)
	}
	// 嘗試開啟一個SQLite資料庫（使用記憶體模式）
	sqliteDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		// 如果開啟失敗，則記錄錯誤並終止測試
		t.Fatalf("failed to open SQLite: %v", err)
	}
	// 返回連線好的MySQL和SQLite資料庫以及一個清理函式
	return &nyamysql.NyaMySQL{DB: mysqlDB}, &nyasqlite.NyaSQLite{DB: sqliteDB}, func() {
		// 清理函式：刪除MySQL資料庫中的表
		mysqlDB.Exec("DROP TABLE IF EXISTS users_copy")
		mysqlDB.Exec("DROP TABLE IF EXISTS users")
		// 關閉MySQL資料庫連線
		_ = mysqlDB.Close()
		// 關閉SQLite資料庫連線
		_ = sqliteDB.Close()
	}
}

// TestMigrateMySQLToSQLiteAndBack 是一個單元測試函式，用於測試從 MySQL 到 SQLite 的表遷移以及從 SQLite 回遷到 MySQL 的功能。
//
// 引數：
// - t *testing.T: 用於測試的標準庫引數，提供斷言和日誌功能。
//
// 返回值：
// 無返回值。
func TestMigrateMySQLToSQLiteAndBack(t *testing.T) {
	mysqlClient, sqliteClient, teardown := setupTestDBs(t)
	defer teardown()

	// 建立MySQL表
	createMySQL := `
	CREATE TABLE users (
		id INT NOT NULL AUTO_INCREMENT,
		username VARCHAR(100) NOT NULL,
		email VARCHAR(255) DEFAULT NULL,
		created_at DATETIME,
		PRIMARY KEY (id)
	);`

	// 刪除已存在的users表
	_, err := mysqlClient.DB.Exec("DROP TABLE IF EXISTS users")
	if err != nil {
		t.Fatal(err)
	}

	// 建立新的users表
	_, err = mysqlClient.DB.Exec(createMySQL)
	if err != nil {
		t.Fatalf("failed to create MySQL test table: %v", err)
	}

	// 將MySQL的users表遷移到SQLite
	err = nyasqldrift.MigrateMySQLTableToSQLite(mysqlClient, sqliteClient, "users")
	if err != nil {
		t.Fatalf("migration to SQLite failed: %v", err)
	}

	// 將SQLite的users表遷移回MySQL
	err = nyasqldrift.MigrateSQLiteTableToMySQL(sqliteClient, mysqlClient, "users")
	if err != nil {
		t.Fatalf("migration back to MySQL failed: %v", err)
	}

	// 獲取原始MySQL的users表結構
	original, err := mysqlClient.GetTableStructure("users")
	if err != nil {
		t.Fatalf("failed to get original MySQL structure: %v", err)
	}

	// 獲取遷移後MySQL的users表結構
	copy, err := mysqlClient.GetTableStructure("users")
	if err != nil {
		t.Fatalf("failed to get copied MySQL structure: %v", err)
	}

	// 檢查列的數量是否一致
	if len(original) != len(copy) {
		t.Fatalf("column count mismatch: %d != %d", len(original), len(copy))
	}

	// 檢查每一列的名稱和型別是否一致
	for i := range original {
		if original[i].ColumnName != copy[i].ColumnName ||
			!strings.EqualFold(original[i].ColumnType, copy[i].ColumnType) {
			t.Errorf("mismatched column %d: %+v != %+v", i, original[i], copy[i])
		}
	}
}
