// 轉換資料庫語句（MySQL|SQLite）的函式庫
package nyasqldrift

import (
	"strings"
)

// Direction 表示转换方向
type Direction int

const (
	MySQLToSQLite Direction = iota
	SQLiteToMySQL
)

// ConvertSQL 函式將給定的 SQL 語句從一種資料庫格式轉換為另一種資料庫格式
//
// 引數：
// sql string - 需要轉換的 SQL 語句
// dir Direction - 轉換方向，可以是 MySQLToSQLite 或 SQLiteToMySQL
//
// 返回值：
// string - 轉換後的 SQL 語句
// error - 如果轉換失敗或 SQL 語句不支援轉換，則返回錯誤
func ConvertSQL(sql string, dir Direction) (string, error) {
	// 去除SQL語句前後的空白字元
	sql = strings.TrimSpace(sql)

	// 根據轉換方向進行不同的處理
	switch dir {
	case MySQLToSQLite:
		// 如果SQL語句以"CREATE TABLE"開頭
		if strings.HasPrefix(strings.ToUpper(sql), "CREATE TABLE") {
			// 呼叫MySQL到SQLite的轉換函式
			return convertCreateTableMySQLToSQLite(sql)
		}
	case SQLiteToMySQL:
		// 如果SQL語句以"CREATE TABLE"開頭
		if strings.HasPrefix(strings.ToUpper(sql), "CREATE TABLE") {
			// 呼叫SQLite到MySQL的轉換函式
			return convertCreateTableSQLiteToMySQL(sql)
		}
	}

	// 如果SQL語句不符合轉換條件，返回錯誤
	return "", ErrUnsupportedStatement
}

/*
// test_main 函式演示瞭如何使用 nyasqldrift 庫將 MySQL SQL 語句轉換為 SQLite SQL 語句，
// 以及如何將轉換後的 SQLite SQL 語句轉換回 MySQL SQL 語句。
//
// 函式首先定義了一個 MySQL SQL 語句，用於建立一個名為 users 的表。
// 然後，函式打印出一條訊息，表示將 MySQL SQL 語句轉換為 SQLite SQL 語句。
//
// 接下來，使用 nyasqldrift.ConvertSQL 函式將 MySQL SQL 語句轉換為 SQLite SQL 語句。
// 如果轉換失敗，則記錄錯誤資訊並退出程式。
//
// 轉換成功後，打印出轉換後的 SQLite SQL 語句。
//
// 然後，函式打印出一條訊息，表示將 SQLite SQL 語句轉換回 MySQL SQL 語句。
// 使用 nyasqldrift.ConvertSQL 函式將 SQLite SQL 語句轉換回 MySQL SQL 語句。
// 如果反向轉換失敗，則記錄錯誤資訊並退出程式。
//
// 最後，打印出轉換後的 MySQL SQL 語句。
func test_main() {
	mysqlSQL := `
CREATE TABLE users (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  email TEXT,
  created_at DATETIME
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
`

	fmt.Println("=== MySQL ➜ SQLite ===")
	// 將MySQL SQL語句轉換為SQLite SQL語句
	sqliteSQL, err := nyasqldrift.ConvertSQL(mysqlSQL, nyasqldrift.MySQLToSQLite)
	if err != nil {
		// 如果轉換失敗，則記錄錯誤資訊並退出程式
		log.Fatalf("Conversion failed: %v", err)
	}
	fmt.Println(sqliteSQL)

	fmt.Println("\n=== SQLite ➜ MySQL ===")
	// 將SQLite SQL語句轉換回MySQL SQL語句
	mysqlBack, err := nyasqldrift.ConvertSQL(sqliteSQL, nyasqldrift.SQLiteToMySQL)
	if err != nil {
		// 如果反向轉換失敗，則記錄錯誤資訊並退出程式
		log.Fatalf("Reverse conversion failed: %v", err)
	}
	fmt.Println(mysqlBack)
}
*/
