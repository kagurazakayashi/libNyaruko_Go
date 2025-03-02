package mysql2sqlite

import (
	"strings"
)

// ConvertSQL 是外部调用的主要函数
func ConvertSQL(mysqlSQL string) (string, error) {
	mysqlSQL = strings.TrimSpace(mysqlSQL)
	if strings.HasPrefix(strings.ToUpper(mysqlSQL), "CREATE TABLE") {
		return convertCreateTable(mysqlSQL)
	}
	// 可以支持更多类型，如 INSERT、ALTER 等
	return "", ErrUnsupportedStatement
}

/*
sqliteSQL, err := mysql2sqlite.ConvertSQL(mysqlSQL)
if err != nil {
	log.Fatal(err)
}
*/
