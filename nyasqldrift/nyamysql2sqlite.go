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
sqliteSQL, err := mysql2sqlite.ConvertSQL(mysqlSQL)
if err != nil {
	log.Fatal(err)
}
*/
