package nyamysql

import (
	"database/sql"
	"fmt"

	"github.com/tidwall/gjson"
)

type NyaMySQL NyaMySQLT
type NyaMySQLT struct {
	db  *sql.DB
	err error
}

//New: 建立新的 NyaMySQL 例項
//	`configJsonString` string 配置 JSON 字串
//	從配置 JSON 檔案中取出的本模組所需的配置段落 JSON 字串
//  示例配置數值參考 config.template.json
//	本模組所需配置項: mysql_addr, mysql_port, mysql_db, mysql_user, mysql_pwd
//  return *NyaMySQL 新的 NyaMySQL 例項
//	下一步使用 `Error()` 或 `ErrorString()` 檢查是否有錯誤
func New(configJsonString string) *NyaMySQL {
	var configNG string = "NO CONFIG KEY : "
	var configKey string = "mysql_user"
	var sqlName gjson.Result = gjson.Get(configJsonString, configKey)
	if !sqlName.Exists() {
		return &NyaMySQL{err: fmt.Errorf(configNG + configKey)}
	}
	var name string = sqlName.String()
	configKey = "mysql_pwd"
	var sqlPassword gjson.Result = gjson.Get(configJsonString, configKey)
	if !sqlPassword.Exists() {
		return &NyaMySQL{err: fmt.Errorf(configNG + configKey)}
	}
	var pwd string = sqlPassword.String()
	configKey = "mysql_addr"
	var sqlAddress gjson.Result = gjson.Get(configJsonString, configKey)
	if !sqlAddress.Exists() {
		return &NyaMySQL{err: fmt.Errorf(configNG + configKey)}
	}
	var addr string = sqlAddress.String()
	configKey = "mysql_db"
	var sqlDBName gjson.Result = gjson.Get(configJsonString, configKey)
	if !sqlDBName.Exists() {
		return &NyaMySQL{err: fmt.Errorf(configNG + configKey)}
	}
	var dbname string = sqlDBName.String()
	configKey = "mysql_port"
	var sqlPort gjson.Result = gjson.Get(configJsonString, configKey)
	if !sqlPort.Exists() {
		return &NyaMySQL{err: fmt.Errorf(configNG + configKey)}
	}
	var port string = sqlPort.String()
	var sqlsetting string = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", name, pwd, addr, port, dbname)
	sqldb, err := sql.Open("mysql", sqlsetting)
	if err != nil {
		return &NyaMySQL{err: err}
	}
	if err := sqldb.Ping(); err != nil {
		return &NyaMySQL{err: err}
	}
	return &NyaMySQL{db: sqldb}
}

//SqlExec: 執行 SQL 語句
//	請先對要執行的語句進行安全檢查。建議使用 nyasql 生成 SQL 語句
//	`sqlCmd` string 要執行的 SQL 語句
//	return   int64  如果是插入操作返回 -1 表示失敗
func (p *NyaMySQL) SqlExec(sqlCmd string) int64 {
	result, err := p.db.Exec(sqlCmd)
	p.err = err
	if err != nil {
		return -1
	}
	id, _ := result.LastInsertId()
	return id
}

//Error: 獲取上一次操作時可能產生的錯誤
//	return error 如果有錯誤，返回錯誤物件，如果沒有錯誤返回 nil
func (p *NyaMySQL) Error() error {
	return p.err
}

//ErrorString: 獲取上一次操作時可能產生的錯誤資訊字串
//	return string 如果有錯誤，返回錯誤描述字串，如果沒有錯誤返回空字串
func (p *NyaMySQL) ErrorString() string {
	if p.err == nil {
		return ""
	}
	return p.err.Error()
}

//Close: 斷開與資料庫的連線
func (p *NyaMySQL) Close() {
	p.db.Close()
	p.db = nil
}
