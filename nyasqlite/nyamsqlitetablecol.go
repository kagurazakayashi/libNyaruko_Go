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

func (p *NyaSQLite) GetTableStructure(tableName string) ([]TableColumn, error) {
	query := fmt.Sprintf(`PRAGMA table_info('%s');`, tableName)
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []TableColumn
	for rows.Next() {
		var (
			cid       int
			name      string
			colType   string
			notnull   int
			dfltValue sql.NullString
			pk        int
		)
		if err := rows.Scan(&cid, &name, &colType, &notnull, &dfltValue, &pk); err != nil {
			return nil, err
		}
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

func (p *NyaSQLite) CreateTableFromColumns(tableName string, columns []TableColumn) error {
	var columnDefinitions []string
	var primaryKeys []string

	for _, col := range columns {
		columnDef := fmt.Sprintf("`%s` %s", col.ColumnName, col.ColumnType)

		if col.NotNull {
			columnDef += " NOT NULL"
		}
		if col.DefaultValue.Valid {
			columnDef += fmt.Sprintf(" DEFAULT %s", col.DefaultValue.String)
		}
		if col.PrimaryKey {
			primaryKeys = append(primaryKeys, fmt.Sprintf("`%s`", col.ColumnName))
		}
		columnDefinitions = append(columnDefinitions, columnDef)
	}

	if len(primaryKeys) > 0 {
		columnDefinitions = append(columnDefinitions, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
	}

	createTableSQL := fmt.Sprintf("CREATE TABLE `%s` (\n  %s\n);", tableName, strings.Join(columnDefinitions, ",\n  "))

	_, err := p.db.Exec(createTableSQL)
	if err != nil {
		return err
	}
	return nil
}
