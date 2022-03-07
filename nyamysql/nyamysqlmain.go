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

func Init(confCMap cmap.ConcurrentMap) *NyaMySQL {
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

func (p *NyaMySQL) Close() {
	p.db.Close()
	p.db = nil
}

func (p *NyaMySQL) Error() error {
	return p.err
}

func loadConfig(confCMap cmap.ConcurrentMap, key string) (string, error) {
	val, isExist := confCMap.Get(key)
	if !isExist {
		return "", fmt.Errorf("no config : " + key)
	}
	return val.(string), nil
}
