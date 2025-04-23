// MySQL 資料庫操作
package nyamysql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

// MySQLDBConfig 結構體用於配置MySQL資料庫的連線引數。
// 該結構體包含以下欄位：
// - User: MySQL使用者名稱。
// - Password: MySQL密碼。
// - Address: MySQL伺服器地址。
// - Port: MySQL伺服器埠。
// - DbName: 資料庫名稱。
// - MaxLimit: 資料的最大限制。
type MySQLDBConfig struct {
	User     string `json:"mysql_user" yaml:"mysql_user"`
	Password string `json:"mysql_pwd" yaml:"mysql_pwd"`
	Address  string `json:"mysql_addr" yaml:"mysql_addr"`
	Port     string `json:"mysql_port" yaml:"mysql_port"`
	DbName   string `json:"mysql_db" yaml:"mysql_db"`
	MaxLimit string `json:"mysql_limit" yaml:"mysql_limit"`
}

// NyaMySQL 是一個類型別名，指向 NyaMySQLT 結構體。
// 該類型別名用於簡化程式碼中的型別引用。
type NyaMySQL NyaMySQLT

// NyaMySQLT 結構體用於封裝MySQL資料庫連線及其相關配置。
// 該結構體包含以下欄位：
// - db: 指向MySQL資料庫連線的指標。
// - limit: 資料庫連線的最大限制。
// - err: 用於儲存資料庫操作中的錯誤資訊。
// - debug: 用於除錯的日誌記錄器。
type NyaMySQLT struct {
	db    *sql.DB
	limit string
	err   error
	debug *log.Logger
}

// New 函式用於根據傳入的配置字串建立一個新的 NyaMySQL 例項。
// 該函式嘗試解析配置字串為 JSON 或 YAML 格式，並根據解析結果初始化 MySQL 配置。
// 如果解析成功，則呼叫 NewC 函式建立並返回 NyaMySQL 例項；如果解析失敗，則返回一個包含錯誤的 NyaMySQL 例項。
//
// 引數:
//   - configString: 包含 MySQL 配置資訊的字串，可以是 JSON 或 YAML 格式。
//   - Debug: 用於記錄除錯資訊的日誌記錄器。
//
// 返回值:
//   - *NyaMySQL: 返回一個 NyaMySQL 例項指標，如果配置解析失敗，例項中將包含錯誤資訊。
func New(configString string, Debug *log.Logger) *NyaMySQL {
	var mySQLConfig MySQLDBConfig
	var err error = nil

	// 嘗試將配置字串解析為 JSON 格式
	if err := json.Unmarshal([]byte(configString), &mySQLConfig); err == nil {
		return NewC(mySQLConfig, Debug)
	}

	// 如果 JSON 解析失敗，嘗試將配置字串解析為 YAML 格式
	if err := yaml.Unmarshal([]byte(configString), &mySQLConfig); err == nil {
		return NewC(mySQLConfig, Debug)
	}

	// 如果 JSON 和 YAML 解析均失敗，返回一個包含錯誤的 NyaMySQL 例項
	return &NyaMySQL{err: err}
}

// 设置数据库连接池 最多允许打开的连接总数（包括使用中和空闲）。
func (p *NyaMySQL) SetMaxOpenConns(n int) {
	p.db.SetMaxOpenConns(n)
}

// 最多允许多少个空闲（未被使用）的连接 保留在连接池中，避免频繁建立/释放连接。
func (p *NyaMySQL) SetMaxIdleConns(n int) {
	p.db.SetMaxIdleConns(n)
}

// 每个连接最多可以使用多久（生命周期）。一旦超过这个时间，无论连接是否仍然有效，下次使用时都会被关闭并重新建立。
func (p *NyaMySQL) SetConnMaxIdleTime(d time.Duration) {
	p.db.SetConnMaxIdleTime(d)
}

// 每个连接最多可以使用多久（生命周期）。一旦超过这个时间，无论连接是否仍然有效，下次使用时都会被关闭并重新建立。
func (p *NyaMySQL) SetConnMaxLifetime(d time.Duration) {
	p.db.SetConnMaxLifetime(d)
}

func (p *NyaMySQL) Stats() sql.DBStats {
	return p.db.Stats()
}

// NewC 建立一個新的 NyaMySQL 例項，用於管理與 MySQL 資料庫的連線。
// 該函式接受 MySQL 資料庫的配置資訊和一個除錯日誌記錄器，並返回一個初始化後的 NyaMySQL 物件。
// 如果連線或 ping 資料庫時發生錯誤，返回的 NyaMySQL 物件將包含該錯誤資訊。
//
// 引數:
//   - mySQLConfig: MySQLDBConfig 結構體，包含 MySQL 資料庫的連線配置資訊，如使用者名稱、密碼、地址、埠和資料庫名稱。
//   - Debug: *log.Logger 型別，用於記錄除錯資訊的日誌記錄器。可以為 nil，表示不記錄除錯資訊。
//
// 返回值:
//   - *NyaMySQL: 返回一個指向 NyaMySQL 結構體的指標，該結構體包含資料庫連線、最大連線限制和除錯日誌記錄器。
func NewC(mySQLConfig MySQLDBConfig, Debug *log.Logger) *NyaMySQL {
	// 根據 MySQL 配置資訊生成連線字串
	var sqlsetting string = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", mySQLConfig.User, mySQLConfig.Password, mySQLConfig.Address, mySQLConfig.Port, mySQLConfig.DbName)

	// 使用生成的連線字串開啟資料庫連線
	sqldb, err := sql.Open("mysql", sqlsetting)
	if err != nil {
		// 如果連線失敗，返回包含錯誤資訊的 NyaMySQL 物件
		return &NyaMySQL{err: err}
	}

	// 嘗試 ping 資料庫以驗證連線是否成功
	if err := sqldb.Ping(); err != nil {
		// 如果 ping 失敗，返回包含錯誤資訊的 NyaMySQL 物件
		return &NyaMySQL{err: err}
	}

	// 返回成功初始化的 NyaMySQL 物件，包含資料庫連線、最大連線限制和除錯日誌記錄器
	return &NyaMySQL{
		db:    sqldb,
		limit: mySQLConfig.MaxLimit,
		debug: Debug,
	}
}

// SqlExec 執行給定的SQL命令，並返回受影響行的最後插入ID。
// 如果執行過程中發生錯誤，返回-1，並將錯誤資訊儲存在p.err中。
//
// 引數:
//   - sqlCmd: 要執行的SQL命令字串。
//
// 返回值:
//   - int64: 成功執行時返回最後插入的ID，失敗時返回-1。
func (p *NyaMySQL) SqlExec(sqlCmd string) int64 {
	// 執行SQL命令
	result, err := p.db.Exec(sqlCmd)
	p.err = err

	// 如果執行過程中發生錯誤，返回-1
	if err != nil {
		return -1
	}

	// 獲取最後插入的ID並返回
	id, _ := result.LastInsertId()
	return id
}

// Error 返回 NyaMySQL 例項中儲存的上一次操作產生的錯誤。
// 該函式通常用於檢查在執行資料庫操作時是否發生了錯誤。
//
// 返回值:
//   - error: 返回上一次操作中儲存的錯誤物件。如果沒有錯誤發生，則返回 nil。
func (p *NyaMySQL) Error() error {
	return p.err
}

// ErrorString 返回與 NyaMySQL 例項關聯的錯誤資訊字串。
// 如果沒有錯誤，則返回空字串。
//
// 返回值:
//   - string: 如果存在錯誤，返回錯誤描述字串；否則返回空字串。
func (p *NyaMySQL) ErrorString() string {
	if p.err == nil {
		return ""
	}
	// 返回底層錯誤物件的錯誤資訊。
	return p.err.Error()
}

// Close 關閉與 NyaMySQL 例項關聯的資料庫連線。
// 如果資料庫連線已經關閉或未初始化，則此函式不會執行任何操作。
// 該方法確保在關閉連線後，將內部資料庫連線指標設定為 nil，以避免重複關閉或空指標引用。
func (p *NyaMySQL) Close() {
	// 檢查資料庫連線是否已初始化
	if p.db != nil {
		// 關閉資料庫連線
		p.db.Close()
		// 將資料庫連線指標置為 nil，防止重複關閉
		p.db = nil
	}
}

// SQLScurity 函式用於檢測輸入字串中是否包含常見的SQL注入攻擊模式。
// 該函式透過正則表示式匹配來識別潛在的SQL注入關鍵字和模式。
//
// 引數:
//   - text: 需要檢測的字串。
//
// 返回值:
//   - 如果檢測到SQL注入模式，返回字串 "21001"；
//   - 如果未檢測到SQL注入模式，返回空字串；
//   - 如果正則表示式編譯失敗，返回錯誤資訊。
func SQLScurity(text string) string {
	// 定義正則表示式模式，用於匹配常見的SQL注入關鍵字和模式
	// str := `(?:')|(?:--)|(/\\*(?:.|[\\n\\r])*?\\*/)|(\b(select|update|and|or|delete|insert|trancate|char|chr|into|substr|ascii|declare|exec|count|master|into|drop|execute)\b)`
	str := `(?:')|(?:--)|(?:#)|(/\\*(?:.|[\\n\\r])*?\\*/)|(\b(select|update|and|or|delete|insert|truncate|char|chr|into|substr|ascii|declare|exec|count|master|into|drop|execute|union|join|having|group by|order by|case|when|then|else|cast|convert|nullif|coalesce|xp_cmdshell|sp_executesql)\b)|(\+|\|\|)|(0x[0-9a-fA-F]+)|(%[0-9a-fA-F]{2})|(1=1|1=0)|(' OR '1'='1|' AND '1'='1)|(; DROP TABLE|; SHUTDOWN)`
	reint := ""

	// 編譯正則表示式
	re, err := regexp.Compile(str)
	if err != nil {
		// 如果編譯失敗，返回錯誤資訊
		return err.Error()
	}

	// 檢測輸入字串是否匹配正則表示式
	if re.MatchString(text) {
		// 如果匹配，返回 "21001" 表示檢測到SQL注入模式
		reint = "21001"
	}

	// 返回檢測結果
	return reint
}
