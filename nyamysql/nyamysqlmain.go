package nyamysql

import (
	"database/sql"
	"fmt"

	cmap "github.com/orcaman/concurrent-map"
)

type NyaMySQL NyaMySQLT
type NyaMySQLT struct {
	db  *sql.DB
	err error
}

//New: 建立新的 NyaMySQL 例項
//	`confCMap` cmap.ConcurrentMap 載入的配置檔案字典
//  return *NyaMySQL 新的 NyaMySQL 例項
//	下一步使用 `Error()` 或 `ErrorString()` 檢查是否有錯誤
func New(confCMap cmap.ConcurrentMap) *NyaMySQL {
	sqlname, err := loadConfig(confCMap, "mysql_user")
	if err != nil {
		return &NyaMySQL{err: err}
	}
	sqlpassword, err := loadConfig(confCMap, "mysql_pwd")
	if err != nil {
		return &NyaMySQL{err: err}
	}
	sqlpath, err := loadConfig(confCMap, "mysql_addr")
	if err != nil {
		return &NyaMySQL{err: err}
	}
	sqlport, err := loadConfig(confCMap, "mysql_port")
	if err != nil {
		return &NyaMySQL{err: err}
	}
	sqllibrary, err := loadConfig(confCMap, "mysql_db")
	if err != nil {
		return &NyaMySQL{err: err}
	}
	sqlsetting := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", sqlname, sqlpassword, sqlpath, sqlport, sqllibrary)
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

//loadConfig: 從載入的配置檔案中載入配置
//	`confCMap` cmap.ConcurrentMap 載入的配置檔案字典
//	`key`      string             配置名稱
//	return     string             配置內容
//	return     error              可能遇到的錯誤
func loadConfig(confCMap cmap.ConcurrentMap, key string) (string, error) {
	val, isExist := confCMap.Get(key)
	if !isExist {
		return "", fmt.Errorf("no config : " + key)
	}
	return val.(string), nil
}

//Close: 斷開與資料庫的連線
func (p *NyaMySQL) Close() {
	p.db.Close()
	p.db = nil
}
