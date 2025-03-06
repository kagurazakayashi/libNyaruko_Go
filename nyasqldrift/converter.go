package nyasqldrift

import (
	"regexp"
	"strings"
)

// convertCreateTableMySQLToSQLite 將 MySQL 風格的 SQL 語句轉換為 SQLite 風格的 SQL 語句
//
// 引數:
//
//	mysqlSQL: MySQL 風格的 SQL 語句
//
// 返回值:
//
//	string: 轉換後的 SQLite 風格的 SQL 語句
//	error: 如果發生錯誤，則返回錯誤資訊
func convertCreateTableMySQLToSQLite(mysqlSQL string) (string, error) {
	// 編譯正則表示式，用於匹配以 ") ENGINE=" 開頭的字串，並忽略大小寫
	reEngine := regexp.MustCompile(`(?i)\)\s*ENGINE=.*?;`)
	// 使用正則表示式替換掉原始 SQL 語句中的 ENGINE 部分
	mysqlSQL = reEngine.ReplaceAllString(mysqlSQL, ");")

	// 按行分割 SQL 語句
	lines := strings.Split(mysqlSQL, "\n")
	// 遍歷每一行
	for i, line := range lines {
		// 去除行首尾的空白字元
		line = strings.TrimSpace(line)
		// 如果行是以 "CREATE TABLE" 開頭（忽略大小寫），或者行為空，或者行為 ");"，則跳過該行
		if strings.HasPrefix(strings.ToUpper(line), "CREATE TABLE") || line == "" || line == ");" {
			continue
		}
		// 替換資料型別
		line = replaceDataTypes(line)
		// 如果行中包含 "AUTO_INCREMENT"（忽略大小寫）
		if strings.Contains(strings.ToUpper(line), "AUTO_INCREMENT") {
			// 替換掉 "AUTO_INCREMENT"
			line = strings.ReplaceAll(line, "AUTO_INCREMENT", "")
			line = strings.ReplaceAll(line, "auto_increment", "")
			// 如果行中不包含 "PRIMARY KEY"（忽略大小寫）
			if !strings.Contains(strings.ToUpper(line), "PRIMARY KEY") {
				// 在行末新增 " PRIMARY KEY AUTOINCREMENT"
				line += " PRIMARY KEY AUTOINCREMENT"
			} else {
				// 替換 "PRIMARY KEY" 為 "PRIMARY KEY AUTOINCREMENT"
				line = strings.ReplaceAll(line, "PRIMARY KEY", "PRIMARY KEY AUTOINCREMENT")
			}
		}

		// 更新行內容
		lines[i] = line
	}
	// 將修改後的行重新拼接成完整的 SQL 語句
	newSQL := strings.Join(lines, "\n")
	return newSQL, nil
}

// convertCreateTableSQLiteToMySQL 將 SQLite 的建立表語句轉換為 MySQL 相容的建立表語句
// 引數：
// sqliteSQL：要轉換的 SQLite 建立表語句
// 返回值：
// string：轉換後的 MySQL 建立表語句
// error：可能出現的錯誤
func convertCreateTableSQLiteToMySQL(sqliteSQL string) (string, error) {
	// 將輸入的SQLite SQL語句按行分割
	lines := strings.Split(sqliteSQL, "\n")
	for i, line := range lines {
		// 去除每行首尾的空白字元
		line = strings.TrimSpace(line)
		// 如果當前行以"CREATE TABLE"開頭，或者為空行，或者為以");"結尾的行，則跳過處理
		if strings.HasPrefix(strings.ToUpper(line), "CREATE TABLE") || line == "" || line == ");" {
			continue
		}
		// 恢復資料型別
		line = restoreDataTypes(line)
		// 如果當前行包含"AUTOINCREMENT"
		if strings.Contains(strings.ToUpper(line), "AUTOINCREMENT") {
			// 將"AUTOINCREMENT"替換為空字串
			line = strings.ReplaceAll(line, "AUTOINCREMENT", "")
			// 如果當前行包含"PRIMARY KEY"
			if strings.Contains(strings.ToUpper(line), "PRIMARY KEY") {
				// 將"PRIMARY KEY"替換為"AUTO_INCREMENT PRIMARY KEY"
				line = strings.ReplaceAll(line, "PRIMARY KEY", "AUTO_INCREMENT PRIMARY KEY")
			} else {
				// 在當前行末尾新增" AUTO_INCREMENT PRIMARY KEY"
				line += " AUTO_INCREMENT PRIMARY KEY"
			}
		}

		// 更新lines陣列中的當前行
		lines[i] = line
	}
	// 將處理後的行重新拼接成完整的SQL語句
	newSQL := strings.Join(lines, "\n")
	// 去除newSQL末尾的分號
	newSQL = strings.TrimSuffix(newSQL, ";")
	// 在newSQL末尾新增" ENGINE=InnoDB DEFAULT CHARSET=utf8;"
	newSQL += " ENGINE=InnoDB DEFAULT CHARSET=utf8;"
	return newSQL, nil
}

/*
// test_main 是一個測試函式，用於演示如何從 MySQL 資料庫遷移資料到 SQLite 資料庫，以及從 SQLite 資料庫遷移資料回 MySQL 資料庫。
// 該函式首先獲取環境變數 MYSQL_TEST_DSN 的值，該值包含 MySQL 資料庫的連線字串。
// 如果環境變數未設定，則程式將終止並輸出錯誤資訊。
// 接著，函式嘗試開啟 MySQL 和 SQLite 資料庫連線。
// 如果連線失敗，程式將終止並輸出錯誤資訊。
// 然後，函式建立了一個名為 "users" 的表（如果不存在），並插入了一些示例資料。
// 隨後，函式執行了兩次資料遷移操作：首先將 MySQL 資料庫中的 "users" 表遷移到 SQLite 資料庫中，
// 然後將 SQLite 資料庫中的 "users" 表遷移回 MySQL 資料庫中。
// 如果遷移失敗，程式將終止並輸出錯誤資訊。
// 最後，程式輸出成功完成所有操作的訊息。
func test_main() {
	mysqlDSN := os.Getenv("MYSQL_TEST_DSN") // 如 "user:pass@tcp(localhost:3306)/testdb"
	// 檢查環境變數MYSQL_TEST_DSN是否設定
	if mysqlDSN == "" {
		log.Fatal("MYSQL_TEST_DSN environment variable not set")
	}
	// 連線到MySQL資料庫
	mysqlDB, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		log.Fatalf("MySQL connect failed: %v", err)
	}
	// 關閉MySQL資料庫連線
	defer mysqlDB.Close()
	// 連線到SQLite資料庫
	sqliteDB, err := sql.Open("sqlite3", "./test.sqlite")
	if err != nil {
		log.Fatalf("SQLite open failed: %v", err)
	}
	// 關閉SQLite資料庫連線
	defer sqliteDB.Close()
	// 建立MySQL客戶端例項
	mysqlClient := &nyamysql.NyaMySQL{DB: mysqlDB}
	// 建立SQLite客戶端例項
	sqliteClient := &nyasqlite.NyaSQLite{DB: sqliteDB}
	// 定義表名
	tableName := "users"
	// 在MySQL中建立表
	_, err = mysqlDB.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(100) NOT NULL,
		email VARCHAR(255),
		created_at DATETIME
	);`, tableName))
	if err != nil {
		log.Fatalf("Failed to create source table in MySQL: %v", err)
	}
	// 列印日誌，表示MySQL表已建立
	log.Printf("[MySQL] Source table `%s` ready", tableName)
	// 列印日誌，表示開始從MySQL遷移到SQLite
	log.Println("[Stage 1] Migrating MySQL ➜ SQLite...")
	// 執行遷移操作
	if err := migrator.MigrateMySQLTableToSQLite(mysqlClient, sqliteClient, tableName); err != nil {
		log.Fatalf("Migration to SQLite failed: %v", err)
	}
	// 列印日誌，表示遷移完成
	log.Println("Migration to SQLite complete.")
	// 定義複製表名
	copyTableName := "users_copy"
	// 列印日誌，表示開始從SQLite遷移到MySQL
	log.Println("[Stage 2] Migrating SQLite ➜ MySQL...")
	// 執行遷移操作
	if err := migrator.MigrateSQLiteTableToMySQL(sqliteClient, mysqlClient, tableName); err != nil {
		log.Fatalf("Migration back to MySQL failed: %v", err)
	}
	// 列印日誌，表示遷移完成
	log.Println("Migration back to MySQL complete.")

	// 列印日誌，表示所有操作完成
	log.Println("All done! Inspect your databases for results.")
}
*/
