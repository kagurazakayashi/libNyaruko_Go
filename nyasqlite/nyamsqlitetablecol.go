// SQLite 表結構
package nyasqlite

import (
	"database/sql"
	"fmt"
	"strings"
)

type TableColumn struct {
	ColumnName   string
	ColumnType   string
	NotNull      bool
	DefaultValue sql.NullString
	PrimaryKey   bool
}

// GetTableStructure 獲取指定表的結構資訊
//
// 引數:
// tableName - 表名
//
// 返回值:
// []TableColumn - 表結構資訊切片
// error - 錯誤資訊
func (p *NyaSQLite) GetTableStructure(tableName string) ([]TableColumn, error) {
	// 構建查詢語句
	query := fmt.Sprintf(`PRAGMA table_info('%s');`, tableName)
	// 執行查詢
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	// 確保在函式結束時關閉行集
	defer rows.Close()

	// 初始化列切片
	var columns []TableColumn
	for rows.Next() {
		// 宣告變數以儲存查詢結果
		var (
			cid       int
			name      string
			colType   string
			notnull   int
			dfltValue sql.NullString
			pk        int
		)
		// 將查詢結果掃描到變數中
		if err := rows.Scan(&cid, &name, &colType, &notnull, &dfltValue, &pk); err != nil {
			return nil, err
		}
		// 將變數轉換為 TableColumn 結構並新增到切片中
		columns = append(columns, TableColumn{
			ColumnName:   name,
			ColumnType:   colType,
			NotNull:      notnull == 1,
			DefaultValue: dfltValue,
			PrimaryKey:   pk > 0,
		})
	}
	return columns, nil
}

// CreateTableFromColumns 根據列定義建立表
//
// tableName 是要建立的表的名稱
// columns 是包含列定義的切片，每個列定義包含列名、資料型別、是否允許為空、預設值以及是否是主鍵
//
// 返回值是一個 error 物件，如果建立表成功，則返回 nil，否則返回具體的錯誤資訊
func (p *NyaSQLite) CreateTableFromColumns(tableName string, columns []TableColumn) error {
	var columnDefinitions []string
	var primaryKeys []string

	// 遍歷columns陣列，為每一個column生成定義
	for _, col := range columns {
		columnDef := fmt.Sprintf("`%s` %s", col.ColumnName, col.ColumnType)

		// 如果該列不能為空，則在定義後新增NOT NULL
		if col.NotNull {
			columnDef += " NOT NULL"
		}
		// 如果該列有預設值，則在定義後新增DEFAULT
		if col.DefaultValue.Valid {
			columnDef += fmt.Sprintf(" DEFAULT %s", col.DefaultValue.String)
		}
		// 如果該列是主鍵，則將其名稱新增到primaryKeys陣列中
		if col.PrimaryKey {
			primaryKeys = append(primaryKeys, fmt.Sprintf("`%s`", col.ColumnName))
		}
		columnDefinitions = append(columnDefinitions, columnDef)
	}

	// 如果存在主鍵，則在columnDefinitions陣列後新增PRIMARY KEY定義
	if len(primaryKeys) > 0 {
		columnDefinitions = append(columnDefinitions, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
	}

	// 生成建立表的SQL語句
	createTableSQL := fmt.Sprintf("CREATE TABLE `%s` (\n  %s\n);", tableName, strings.Join(columnDefinitions, ",\n  "))

	// 執行SQL語句建立表
	_, err := p.db.Exec(createTableSQL)
	if err != nil {
		return err
	}
	return nil
}
