// MySQL 表結構
package nyamysql

import "database/sql"

// TableColumn 結構體表示資料庫表的列結構資訊
// 每個欄位對應 MySQL INFORMATION_SCHEMA.COLUMNS 表中的列
type TableColumn struct {
	ColumnName    string         // 列名
	ColumnType    string         // 列的資料型別（如 INT、VARCHAR(255) 等）
	IsNullable    string         // 是否允許為空（YES/NO）
	ColumnKey     string         // 是否是主鍵（PRI）、唯一鍵（UNI）等
	ColumnDefault sql.NullString // 列的預設值（可能為空）
	Extra         string         // 額外資訊（如 auto_increment）
}

// GetTableStructure 查詢資料庫，獲取指定表的結構資訊
//
// 引數：
//   - db: *sql.DB，資料庫連線物件
//   - tableName: string，需要查詢的表名
//
// 返回值：
//   - []TableColumn: 返回表的結構資訊，包含所有列的詳細資訊
//   - error: 可能的錯誤資訊
func (p *NyaMySQL) GetTableStructure(tableName string) ([]TableColumn, error) {
	// SQL 查詢語句：從 INFORMATION_SCHEMA.COLUMNS 獲取表的列資訊
	query := `
		SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_KEY, COLUMN_DEFAULT, EXTRA
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?`
	// 執行查詢
	rows, err := p.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// 用於儲存表結構的切片
	var columns []TableColumn
	// 遍歷查詢結果
	for rows.Next() {
		var column TableColumn
		// 讀取查詢結果，並存入 TableColumn 結構體
		err := rows.Scan(&column.ColumnName, &column.ColumnType, &column.IsNullable, &column.ColumnKey, &column.ColumnDefault, &column.Extra)
		if err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}
	return columns, nil
}
