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

func (p *NyaMySQL) GetTableStructure(tableName string) ([]TableColumn, error) {
	query := `
		SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_KEY, COLUMN_DEFAULT, EXTRA
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?`
	rows, err := p.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var columns []TableColumn
	for rows.Next() {
		var column TableColumn
		err := rows.Scan(&column.ColumnName, &column.ColumnType, &column.IsNullable, &column.ColumnKey, &column.ColumnDefault, &column.Extra)
		if err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}
	return columns, nil
}
func (p *NyaMySQL) CreateTableFromColumns(tableName string, columns []TableColumn) error {
	var columnDefinitions []string
	var primaryKeys []string
	for _, col := range columns {
		columnDef := fmt.Sprintf("`%s` %s", col.ColumnName, col.ColumnType)
		if col.IsNullable == "NO" {
			columnDef += " NOT NULL"
		}
		if col.ColumnDefault.Valid {
			columnDef += fmt.Sprintf(" DEFAULT '%s'", col.ColumnDefault.String)
		}
		if col.Extra != "" {
			columnDef += " " + col.Extra
		}
		if col.ColumnKey == "PRI" {
			primaryKeys = append(primaryKeys, fmt.Sprintf("`%s`", col.ColumnName))
		}
		columnDefinitions = append(columnDefinitions, columnDef)
	}
	if len(primaryKeys) > 0 {
		columnDefinitions = append(columnDefinitions, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
	}
	createTableSQL := fmt.Sprintf("CREATE TABLE `%s` (\n  %s\n);", tableName, strings.Join(columnDefinitions, ",\n  "))
	rows, err := p.db.Query(createTableSQL)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}
