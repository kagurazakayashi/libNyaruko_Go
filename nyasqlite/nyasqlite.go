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

func Init(confCMap cmap.ConcurrentMap) NyaSQLite {
	sqlLiteVersion, err := loadConfig(confCMap, "sqlite_ver")
	if err != nil {
		return NyaSQLite{err: err}
	}
	sqlLiteFile, err := loadConfig(confCMap, "sqlite_file")
	if err != nil {
		return NyaSQLite{err: err}
	}
	sqlLiteDB, err := sql.Open(sqlLiteVersion, sqlLiteFile)
	if err != nil {
		return NyaSQLite{err: err}
	}
	return NyaSQLite{db: sqlLiteDB}
}

func (p NyaSQLite) Close() {
	p.db.Close()
	p.db = nil
}

func (p NyaSQLite) Error() error {
	return p.err
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
func (p NyaSQLite) sqliteAddRecord(table string, key string, val string, values string) (int64, error) {
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
