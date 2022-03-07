package nyasqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	cmap "github.com/orcaman/concurrent-map"
)

type NyaSQLite NyaSQLiteT
type NyaSQLiteT struct {
	db  *sql.DB
	err error
}

func Init(confCMap cmap.ConcurrentMap) *NyaSQLite {
	sqlLiteVersion, err := loadConfig(confCMap, "sqlite_ver")
	if err != nil {
		return &NyaSQLite{err: err}
	}
	sqlLiteFile, err := loadConfig(confCMap, "sqlite_file")
	if err != nil {
		return &NyaSQLite{err: err}
	}
	sqlLiteDB, err := sql.Open(sqlLiteVersion, sqlLiteFile)
	if err != nil {
		return &NyaSQLite{err: err}
	}
	return &NyaSQLite{db: sqlLiteDB}
}

func (p *NyaSQLite) Close() {
	p.db.Close()
	p.db = nil
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

func (p *NyaSQLite) SqlINSERT(sqlCmd string) int64 {
	var result sql.Result = nil
	fmt.Println(sqlCmd)
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

func loadConfig(confCMap cmap.ConcurrentMap, key string) (string, error) {
	val, isExist := confCMap.Get(key)
	if !isExist {
		return "", fmt.Errorf("no config : " + key)
	}
	return val.(string), nil
}

//向SQL数据库中添加
//	所有关键字除*以外需要用``包裹
//	`table`		string		从哪个表中查询不需要``包裹
//	`key`		string		需要添加的列，需要以,分割
//	`val`		string		与key对应的值，以,分割
//	`values`	string		(此项不为"",val无效)添加多行数据与key对应的值，以,分割,例(1,2),(2,3)
//	return		int64 和 error 对象，返回添加行的id
func (p *NyaSQLite) sqliteAddRecord(table string, key string, val string, values string) (int64, error) {
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
