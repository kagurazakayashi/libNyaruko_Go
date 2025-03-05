// MySQL 表結構
package nyamysql

import (
	"database/sql"
	"fmt"
	"strings"
)

type TableColumn struct {
	ColumnName    string
	ColumnType    string
	IsNullable    string
	ColumnKey     string
	ColumnDefault sql.NullString
	Extra         string
}

// GetTableStructure 獲取指定表的表結構
// tableName 是要查詢的表名
// 返回值為 TableColumn 型別的切片和一個 error 物件
// 如果查詢失敗，則返回 nil 和相應的 error 物件
func (p *NyaMySQL) GetTableStructure(tableName string) ([]TableColumn, error) {
	// 構造查詢表結構的SQL語句
	query := `
		SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_KEY, COLUMN_DEFAULT, EXTRA
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?`

	// 執行查詢
	rows, err := p.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}

	// 確保在函式返回前關閉rows
	defer rows.Close()

	var columns []TableColumn
	// 遍歷查詢結果
	for rows.Next() {
		var column TableColumn
		// 掃描當前行到column變數中
		err := rows.Scan(&column.ColumnName, &column.ColumnType, &column.IsNullable, &column.ColumnKey, &column.ColumnDefault, &column.Extra)
		if err != nil {
			return nil, err
		}
		// 將當前列新增到columns切片中
		columns = append(columns, column)
	}

	return columns, nil
}

// CreateTableFromColumns 根據給定的列資訊建立表
//
// 引數：
// - tableName: 要建立的表名
// - columns: 表列的資訊，包括列名、資料型別、是否可為空、預設值、額外屬性以及主鍵標識
//
// 返回值：
// - error: 如果建立表過程中發生錯誤，則返回錯誤物件；否則返回nil
func (p *NyaMySQL) CreateTableFromColumns(tableName string, columns []TableColumn) error {
	// 定義列定義和主鍵的切片
	var columnDefinitions []string
	var primaryKeys []string

	// 遍歷列資訊
	for _, col := range columns {
		// 構建列定義字串
		columnDef := fmt.Sprintf("`%s` %s", col.ColumnName, col.ColumnType)

		// 如果列不可為空，則新增 NOT NULL
		if col.IsNullable == "NO" {
			columnDef += " NOT NULL"
		}

		// 如果列有預設值，則新增預設值
		if col.ColumnDefault.Valid {
			columnDef += fmt.Sprintf(" DEFAULT '%s'", col.ColumnDefault.String)
		}

		// 如果列有額外資訊，則新增額外資訊
		if col.Extra != "" {
			columnDef += " " + col.Extra
		}

		// 如果列是主鍵，則新增到主鍵切片中
		if col.ColumnKey == "PRI" {
			primaryKeys = append(primaryKeys, fmt.Sprintf("`%s`", col.ColumnName))
		}

		// 將列定義新增到列定義切片中
		columnDefinitions = append(columnDefinitions, columnDef)
	}

	// 如果存在主鍵，則構建主鍵定義字串並新增到列定義切片中
	if len(primaryKeys) > 0 {
		columnDefinitions = append(columnDefinitions, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
	}

	// 構建建立表的 SQL 語句
	createTableSQL := fmt.Sprintf("CREATE TABLE `%s` (\n  %s\n);", tableName, strings.Join(columnDefinitions, ",\n  "))

	// 執行建立表的 SQL 語句
	rows, err := p.db.Query(createTableSQL)
	if err != nil {
		return err
	}

	// 延遲關閉結果集
	defer rows.Close()

	return nil
}
