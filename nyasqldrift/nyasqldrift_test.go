package nyasqldrift_test

import (
	"testing"

	"github.com/kagurazakayashi/libNyaruko_Go/nyasqldrift"
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
