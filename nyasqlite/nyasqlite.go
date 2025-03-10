// SQLite 資料庫操作
package nyasqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v3"
)

type SQLiteConfig struct {
	SQLiteVer  string `json:"sqlite_ver" yaml:"sqlite_ver"`
	SQLiteFile string `json:"sqlite_file" yaml:"sqlite_file"`
}

type NyaSQLite NyaSQLiteT
type NyaSQLiteT struct {
	db  *sql.DB
	err error
}

// New 函式用於根據配置字串建立一個新的 NyaSQLite 例項。
// 該函式首先嚐試將配置字串解析為 JSON 格式，如果失敗則嘗試解析為 YAML 格式。
// 如果兩種格式均無法解析，則返回一個包含錯誤的 NyaSQLite 例項。
//
// 引數:
//   - configString: 包含配置資訊的字串，可以是 JSON 或 YAML 格式。
//   - Debug: 用於除錯的日誌記錄器，可以為 nil。
//
// 返回值:
//   - *NyaSQLite: 返回一個 NyaSQLite 例項，如果解析失敗則包含錯誤資訊。
func New(configString string, Debug *log.Logger) *NyaSQLite {
	var sqliteConfig SQLiteConfig
	var err error = nil

	// 嘗試將配置字串解析為 JSON 格式
	if err := json.Unmarshal([]byte(configString), &sqliteConfig); err == nil {
		return NewC(sqliteConfig, Debug)
	}

	// 如果 JSON 解析失敗，嘗試將配置字串解析為 YAML 格式
	if err := yaml.Unmarshal([]byte(configString), &sqliteConfig); err == nil {
		return NewC(sqliteConfig, Debug)
	}

	// 如果 JSON 和 YAML 解析均失敗，返回一個包含錯誤的 NyaSQLite 例項
	return &NyaSQLite{err: err}
}

// NewC 初始化並返回一個新的 NyaSQLite 例項。
// 該函式負責建立與 SQLite 資料庫的連線，並返回一個包含資料庫連線和錯誤資訊的 NyaSQLite 結構體。
//
// 引數:
//   - sqliteConfig: SQLiteConfig 結構體，包含 SQLite 資料庫的版本資訊和檔案路徑。
//   - Debug: *log.Logger 型別的日誌記錄器，用於除錯日誌輸出（當前程式碼中未使用）。
//
// 返回值:
//   - *NyaSQLite: 返回一個 NyaSQLite 結構體指標，包含資料庫連線和錯誤資訊。
//     如果連線失敗，返回的 NyaSQLite 結構體中將包含錯誤資訊。
func NewC(sqliteConfig SQLiteConfig, Debug *log.Logger) *NyaSQLite {
	var err error = nil

	// 嘗試開啟 SQLite 資料庫連線
	sqlLiteDB, err := sql.Open(sqliteConfig.SQLiteVer, sqliteConfig.SQLiteFile)
	if err != nil {
		// 如果連線失敗，返回包含錯誤資訊的 NyaSQLite 例項
		return &NyaSQLite{err: err}
	}

	// 返回包含成功連線的 NyaSQLite 例項
	return &NyaSQLite{
		db:  sqlLiteDB,
		err: nil,
	}
}

// SqlExec 執行給定的SQL命令，並返回最後插入行的ID。
// 如果執行過程中發生錯誤，則返回-1。
//
// 引數:
//   - sqlCmd: 要執行的SQL命令字串。
//
// 返回值:
//   - int64: 最後插入行的ID，如果發生錯誤則返回-1。
func (p *NyaSQLite) SqlExec(sqlCmd string) int64 {
	var result sql.Result = nil
	result, p.err = p.db.Exec(sqlCmd)
	if p.err != nil {
		return -1
	}

	// 獲取最後插入行的ID
	var id int64 = -1
	id, p.err = result.LastInsertId()
	if p.err != nil {
		return -1
	}

	return id
}

// Error 返回 NyaSQLite 例項中儲存的上一次操作產生的錯誤。
// 該函式通常用於檢查在執行資料庫操作時是否發生了錯誤。
//
// 返回值:
//   - error: 返回上一次操作中儲存的錯誤物件。如果沒有錯誤發生，則返回 nil。
func (p *NyaSQLite) Error() error {
	return p.err
}

// ErrorString 返回與 NyaMySQL 例項關聯的錯誤資訊字串。
// 如果沒有錯誤，則返回空字串。
//
// 返回值:
//   - string: 如果存在錯誤，返回錯誤描述字串；否則返回空字串。
func (p *NyaSQLite) ErrorString() string {
	if p.err == nil {
		return ""
	}
	return p.err.Error()
}

// SqliteAddRecord 向指定的SQLite表中插入一條記錄。
// 該函式根據提供的表名、鍵、值以及可選的多個值字串，構建並執行SQL插入語句。
//
// 引數:
//   - table: 目標表的名稱，插入操作將在此表中進行。
//   - key: 插入記錄的列名，用於指定插入的欄位。
//   - val: 插入記錄的值，與key對應的欄位值。
//   - values: 可選的多個值字串，如果提供，則直接使用該字串作為插入的值部分。
//
// 返回值:
//   - int64: 插入記錄的自增ID，如果插入失敗則返回0。
//   - error: 如果插入過程中發生錯誤，則返回相應的錯誤資訊。
func (p *NyaSQLite) SqliteAddRecord(table string, key string, val string, values string) (int64, error) {
	// 構建SQL插入語句
	var dbq string = "insert into `" + table + "` (" + key + ")" + "VALUES "
	if values != "" {
		dbq += values
	} else {
		dbq += "(" + val + ")"
	}
	fmt.Print(dbq, "\r\n")

	// 執行SQL語句並獲取結果
	result, err := p.db.Exec(dbq)
	if err != nil {
		return 0, err
	}

	// 獲取插入記錄的自增ID
	id, _ := result.LastInsertId()
	return id, nil
}

// Close 關閉與 NyaSQLite 例項關聯的資料庫連線。
// 如果資料庫連線已經關閉或未初始化，則此函式不會執行任何操作。
// 該方法確保在關閉連線後，將內部資料庫連線指標設定為 nil，以避免重複關閉或空指標引用。
func (p *NyaSQLite) Close() {
	// 檢查資料庫連線是否已初始化
	if p.db != nil {
		// 關閉資料庫連線
		p.db.Close()
		// 將資料庫連線指標置為 nil，防止重複關閉
		p.db = nil
	}
}
