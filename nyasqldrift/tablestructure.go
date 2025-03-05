package nyasqldrift

import (
	"fmt"
	"strings"

	"github.com/kagurazakayashi/libNyaruko_Go/nyamysql"
	"github.com/kagurazakayashi/libNyaruko_Go/nyasqlite"
)

// mapMySQLTypeToSQLite 將 MySQL 型別轉換為 SQLite 型別
//
// 引數：
//   - mysqlType: 需要轉換的 MySQL 型別字串
//
// 返回值：
//   - 返回轉換後的 SQLite 型別字串
func mapMySQLTypeToSQLite(mysqlType string) string {
	// 將mysqlType轉換為小寫
	mysqlType = strings.ToLower(mysqlType)

	// 根據mysqlType的字首，選擇對應的SQLite型別
	switch {
	case strings.HasPrefix(mysqlType, "int"):
		// 如果mysqlType以"int"開頭，則返回"INTEGER"
		return "INTEGER"
	case strings.HasPrefix(mysqlType, "bigint"):
		// 如果mysqlType以"bigint"開頭，則返回"INTEGER"
		return "INTEGER"
	case strings.HasPrefix(mysqlType, "varchar"), strings.HasPrefix(mysqlType, "text"):
		// 如果mysqlType以"varchar"或"text"開頭，則返回"TEXT"
		return "TEXT"
	case strings.HasPrefix(mysqlType, "datetime"), strings.HasPrefix(mysqlType, "timestamp"):
		// 如果mysqlType以"datetime"或"timestamp"開頭，則返回"TEXT"
		return "TEXT"
	case strings.HasPrefix(mysqlType, "double"), strings.HasPrefix(mysqlType, "float"), strings.HasPrefix(mysqlType, "decimal"):
		// 如果mysqlType以"double"、"float"或"decimal"開頭，則返回"REAL"
		return "REAL"
	case strings.HasPrefix(mysqlType, "blob"):
		// 如果mysqlType以"blob"開頭，則返回"BLOB"
		return "BLOB"
	default:
		// 如果mysqlType的字首不符合以上任何條件，則返回"TEXT"
		return "TEXT"
	}
}

// MigrateMySQLTableToSQLite 將 MySQL 資料庫中的表遷移到 SQLite 資料庫中
//
// 引數:
//
//	mysqlClient: *nyamysql.NyaMySQL，MySQL 客戶端指標
//	sqliteClient: *nyasqlite.NyaSQLite，SQLite 客戶端指標
//	tableName: string，要遷移的表名
//
// 返回值:
//
//	error，遷移過程中可能發生的錯誤
//
// 說明:
//
//	本函式首先會從 MySQL 資料庫中獲取指定表的表結構，然後將這些表結構轉換為 SQLite 可識別的格式，
//	最後在 SQLite 資料庫中建立相應的表。如果過程中出現任何錯誤，將返回相應的錯誤資訊。
func MigrateMySQLTableToSQLite(
	mysqlClient *nyamysql.NyaMySQL,
	sqliteClient *nyasqlite.NyaSQLite,
	tableName string,
) error {
	// 從MySQL獲取表結構
	mysqlColumns, err := mysqlClient.GetTableStructure(tableName)
	if err != nil {
		return fmt.Errorf("failed to get MySQL table structure: %w", err)
	}

	// 初始化SQLite表列陣列
	var sqliteColumns []nyasqlite.TableColumn
	for _, col := range mysqlColumns {
		// 將MySQL列轉換為SQLite列
		sqliteColumns = append(sqliteColumns, nyasqlite.TableColumn{
			ColumnName:   col.ColumnName,
			ColumnType:   mapMySQLTypeToSQLite(col.ColumnType),
			NotNull:      col.IsNullable == "NO",
			DefaultValue: col.ColumnDefault,
			PrimaryKey:   col.ColumnKey == "PRI",
		})
	}

	// 在SQLite中建立表
	err = sqliteClient.CreateTableFromColumns(tableName, sqliteColumns)
	if err != nil {
		// 建立表失敗時返回錯誤
		return fmt.Errorf("failed to create SQLite table: %w", err)
	}

	return nil
}

// mapSQLiteTypeToMySQL 將 SQLite 型別對映到 MySQL 型別
//
// 引數：
//
//	sqliteType: SQLite 型別的字串表示
//
// 返回值：
//
//	返回 MySQL 型別的字串表示
func mapSQLiteTypeToMySQL(sqliteType string) string {
	// 將sqliteType轉換為小寫
	t := strings.ToLower(sqliteType)
	switch {
	case strings.Contains(t, "int"):
		// 如果sqliteType包含"int"，則返回"INT"
		return "INT"
	case strings.Contains(t, "text"), strings.Contains(t, "char"):
		// 如果sqliteType包含"text"或"char"，則返回"VARCHAR(255)"
		return "VARCHAR(255)"
	case strings.Contains(t, "real"), strings.Contains(t, "double"), strings.Contains(t, "float"):
		// 如果sqliteType包含"real"、"double"或"float"，則返回"DOUBLE"
		return "DOUBLE"
	case strings.Contains(t, "blob"):
		// 如果sqliteType包含"blob"，則返回"BLOB"
		return "BLOB"
	case strings.Contains(t, "date"), strings.Contains(t, "time"):
		// 如果sqliteType包含"date"或"time"，則返回"DATETIME"
		return "DATETIME"
	default:
		// 預設情況，返回"VARCHAR(255)"作為回退選項
		return "VARCHAR(255)" // fallback
	}
}

// MigrateSQLiteTableToMySQL 將 SQLite 表遷移到 MySQL 中。
// 引數：
// - sqliteClient：指向 SQLite 客戶端的指標。
// - mysqlClient：指向 MySQL 客戶端的指標。
// - tableName：要遷移的表名。
//
// 返回值：
// - error：如果遷移過程中發生錯誤，則返回錯誤資訊；否則返回 nil。
func MigrateSQLiteTableToMySQL(
	sqliteClient *nyasqlite.NyaSQLite,
	mysqlClient *nyamysql.NyaMySQL,
	tableName string,
) error {
	// 從SQLite客戶端獲取表結構
	sqliteColumns, err := sqliteClient.GetTableStructure(tableName)
	if err != nil {
		// 如果獲取表結構失敗，返回錯誤
		return fmt.Errorf("failed to get SQLite table structure: %w", err)
	}

	// 初始化MySQL表列資訊切片
	var mysqlColumns []nyamysql.TableColumn
	// 遍歷SQLite表列資訊
	for _, col := range sqliteColumns {
		// 將SQLite表列資訊轉換為MySQL表列資訊
		mysqlColumns = append(mysqlColumns, nyamysql.TableColumn{
			ColumnName:    col.ColumnName,
			ColumnType:    mapSQLiteTypeToMySQL(col.ColumnType),
			IsNullable:    ifThenElse(col.NotNull, "NO", "YES"),
			ColumnKey:     ifThenElse(col.PrimaryKey, "PRI", ""),
			ColumnDefault: col.DefaultValue,
			Extra:         "",
		})
	}

	// 使用MySQL客戶端根據列資訊建立表
	err = mysqlClient.CreateTableFromColumns(tableName, mysqlColumns)
	if err != nil {
		// 如果建立表失敗，返回錯誤
		return fmt.Errorf("failed to create MySQL table: %w", err)
	}

	return nil
}

// ifThenElse 根據條件返回不同的字串
//
// 引數:
//   - condition: 條件值，布林型別
//   - a: 條件為真時返回的字串
//   - b: 條件為假時返回的字串
//
// 返回值:
//   - 返回字串 a 或 b，根據條件值決定
func ifThenElse(condition bool, a, b string) string {
	// 如果條件為真
	if condition {
		// 返回 a
		return a
	}
	// 如果條件為假，返回 b
	return b
}
