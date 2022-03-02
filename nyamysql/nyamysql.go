package nyamysql

import (
	"database/sql"
	"fmt"

	cmap "github.com/orcaman/concurrent-map"
)

type NyaMySQL NyaMySQLT
type NyaMySQLT struct {
	db *sql.DB
}

func Init(confCMap cmap.ConcurrentMap) (NyaMySQL, error) {
	sqlname, _ := confCMap.Get("mysql_user")
	sqlpassword, _ := confCMap.Get("mysql_pwd")
	sqlpath, _ := confCMap.Get("mysql_addr")
	sqlport, _ := confCMap.Get("mysql_port")
	sqllibrary, _ := confCMap.Get("mysql_db")

	sqlsetting := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", sqlname.(string), sqlpassword.(string), sqlpath.(string), sqlport.(string), sqllibrary.(string))
	sqldb, _ := sql.Open("mysql", sqlsetting)

	if err := sqldb.Ping(); err != nil {
		return NyaMySQL{}, err
	}
	return NyaMySQL{db: sqldb}, nil
}

func (p NyaMySQL) Close() {
	p.db.Close()
	p.db = nil
}
