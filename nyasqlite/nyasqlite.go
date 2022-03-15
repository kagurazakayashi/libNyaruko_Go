// SQLite 資料庫操作
package nyasqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tidwall/gjson"
)

type NyaSQLite NyaSQLiteT
type NyaSQLiteT struct {
	db  *sql.DB
	err error
}

//New: 建立新的 NyaSQLite 例項
//	`configJsonString` string 配置 JSON 字串
//	從配置 JSON 檔案中取出的本模組所需的配置段落 JSON 字串
//	示例配置數值參考 config.template.json
//	本模組所需配置項: sqlite_ver, sqlite_file
//	return *NyaSQLite 新的 NyaSQLite 例項
//	下一步使用 `Error()` 或 `ErrorString()` 檢查是否有錯誤
func New(configJsonString string) *NyaSQLite {
	var configNG string = "NO CONFIG KEY : "
	var configKey string = "sqlite_ver"
	sqlLiteVersion := gjson.Get(configJsonString, configKey)
	if !sqlLiteVersion.Exists() {
		return &NyaSQLite{err: fmt.Errorf(configNG + configKey)}
	}
	var dbver string = sqlLiteVersion.String()
	configKey = "sqlite_file"
	sqlLiteFile := gjson.Get(configJsonString, configKey)
	if !sqlLiteFile.Exists() {
		return &NyaSQLite{err: fmt.Errorf(configNG + configKey)}
	}
	var dbfile string = sqlLiteFile.String()

	sqlLiteDB, err := sql.Open(dbver, dbfile)
	if err != nil {
		return &NyaSQLite{err: err}
	}
	return &NyaSQLite{db: sqlLiteDB}
}

//SqlExec: 執行 SQL 語句
//	請先對要執行的語句進行安全檢查。建議使用 nyasql 生成 SQL 語句
//	`sqlCmd` string 要執行的 SQL 語句
//	return   int64  資料被新增到了哪行，如果是插入操作返回 -1 表示失敗
func (p *NyaSQLite) SqlExec(sqlCmd string) int64 {
	var result sql.Result = nil
	result, p.err = p.db.Exec(sqlCmd)
	if p.err != nil {
		return -1
	}
	var id int64 = -1
	id, p.err = result.LastInsertId()
	if p.err != nil {
		return -1
	}
	return id
}

//Error: 獲取上一次操作時可能產生的錯誤
//	return error 如果有錯誤，返回錯誤物件，如果沒有錯誤返回 nil
func (p *NyaSQLite) Error() error {
	return p.err
}

//ErrorString: 獲取上一次操作時可能產生的錯誤資訊字串
//	return string 如果有錯誤，返回錯誤描述字串，如果沒有錯誤返回空字串
func (p *NyaSQLite) ErrorString() string {
	if p.err == nil {
		return ""
	}
	return p.err.Error()
}

//SqliteAddRecord: 向SQL数据库中添加
//	所有关键字除*以外需要用``包裹
//	`table`  string 从哪个表中查询不需要``包裹
//	`key`    string 需要添加的列，需要以,分割
//	`val`    string 与key对应的值，以,分割
//	`values` string (此项不为"",val无效)添加多行数据与key对应的值，以,分割,例(1,2),(2,3)
//	return   int64  添加行的id
//	return   error  可能发生的错误
func (p *NyaSQLite) SqliteAddRecord(table string, key string, val string, values string) (int64, error) {
	var dbq string = "insert into `" + table + "` (" + key + ")" + "VALUES "
	if values != "" {
		dbq += values
	} else {
		dbq += "(" + val + ")"
	}
	fmt.Print(dbq, "\r\n")
	result, err := p.db.Exec(dbq)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

//Close: 斷開與資料庫的連線
func (p *NyaSQLite) Close() {
	p.db.Close()
	p.db = nil
}
