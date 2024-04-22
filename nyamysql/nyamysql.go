// MySQL 輔助查詢器
package nyamysql

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// QueryDataCMD: 從SQL資料庫中操作并查詢
//
//	所有關鍵字除*以外需要用``包裹
//	`sql`	string			mysql语句
//	`value`	...[]interface{}	语句中的值
//	return cmap.ConcurrentMap 和 error 物件，結構為：
//	{
//	    "0":{"id":1,"name":"1"},
//	    "1":{"id":2,"name":"2"}
//	}
func (p *NyaMySQL) QueryDataCMD(sql string, value ...[]interface{}) (map[string]map[string]string, error) {
	sqls := strings.Split(sql, ";")
	for i, v := range sqls {
		if p.debug != nil {
			p.debug.Println("\n" + dbPrintStr(v, value[i]))
		} else {
			fmt.Println("[QueryData]", dbPrintStr(v, value[i]))
		}
		if value == nil {
			if i+1 == len(sqls) {
				query, err := p.db.Query(v)
				if err != nil {
					if p.debug != nil {
						p.debug.Printf("query faied, error:[%v]", err.Error())
					}
					return map[string]map[string]string{}, err
				}
				return handleQD(query, p.debug)
			} else {
				_, err := p.db.Exec(v)
				if p.debug != nil {
					p.debug.Printf("query faied, error:[%v]", err.Error())
				}
				return map[string]map[string]string{}, err
			}
		} else {
			stmt, err := p.db.Prepare(v)
			if err != nil {
				if p.debug != nil {
					p.debug.Printf("query faied, error:[%v]", err.Error())
				}
				return map[string]map[string]string{}, err
			}
			val := value[i]
			if i+1 == len(sqls) {
				query, err := stmt.Query(val...)
				stmt.Close()
				if err != nil {
					if p.debug != nil {
						p.debug.Printf("query faied, error:[%v]", err.Error())
					}
					return map[string]map[string]string{}, err
				}
				return handleQD(query, p.debug)
			} else {
				_, err = stmt.Exec(val...)
				stmt.Close()
				if err != nil {
					if p.debug != nil {
						p.debug.Printf("query faied, error:[%v]", err.Error())
					}
					return map[string]map[string]string{}, err
				}
			}
		}
	}
	return map[string]map[string]string{}, fmt.Errorf("query is null")
}

// QueryData: 從SQL資料庫中查詢
//
//	所有關鍵字除*以外需要用``包裹
//	`recn`		string		查詢語句的返回。全部：*，指定：`id`
//	`table`		string		從哪個表中查詢，此處可以使用關聯語句
//	`where`		string		where語句部分，最前方不需要填寫where，例：`id`=1
//	`orderby`	string		排序，例：`id` ASC/DESC
//	`limit`		string		分頁，例：1,10
//	`value`		interface{}	查詢條件的值
//	return cmap.ConcurrentMap 和 error 物件，結構為：
//	{
//	    "0":{"id":1,"name":"1"},
//	    "1":{"id":2,"name":"2"}
//	}
func (p *NyaMySQL) QueryData(recn string, table string, where string, orderby string, limit string, value ...interface{}) (map[string]map[string]string, error) {
	var dbq string = "select " + recn + " from `" + table + "`"
	if where != "" {
		dbq += " where " + where
	}
	if orderby != "" {
		dbq += " ORDER BY " + orderby
	}
	if limit != "" {
		dbq += " limit " + limit
	} else {
		dbq += " limit " + p.limit
	}
	if p.debug != nil {
		p.debug.Println("\n" + dbPrintStr(dbq, value))
	} else {
		fmt.Println("[QueryData]", dbPrintStr(dbq, value))
	}
	var (
		query *sql.Rows
		err   error
	)

	if value == nil {
		query, err = p.db.Query(dbq)
		if err != nil {
			if p.debug != nil {
				p.debug.Printf("query faied, error:[%v]", err.Error())
			}
			return map[string]map[string]string{}, err
		}
	} else {
		stmt, err := p.db.Prepare(dbq)
		if err != nil {
			if p.debug != nil {
				p.debug.Printf("query faied, error:[%v]", err.Error())
			}
			return map[string]map[string]string{}, err
		}

		query, err = stmt.Query(value...)
		stmt.Close()
		if err != nil {
			if p.debug != nil {
				p.debug.Printf("query faied, error:[%v]", err.Error())
			}
			return map[string]map[string]string{}, err
		}
	}
	return handleQD(query, p.debug)
}

// AddRecord: 向SQL資料庫中新增
//
//	所有關鍵字除*以外需要用``包裹
//	`table`		string		從哪個表中查詢不需要``包裹
//	`key`		string		需要新增的列，需要以,分割
//	`values`	...interface{}	額外的新增值，會在values後面新增
//	return int64 和 error 物件，返回新增行的id
func (p *NyaMySQL) AddRecord(table string, key []string, values ...interface{}) (int64, error) {
	if len(values)%len(key) != 0 {
		return 0, fmt.Errorf("'values'内容数量与'key'不符")
	}
	var dbq string = "insert into `" + table + "` ("
	keyStr := ""
	for _, v := range key {
		if keyStr != "" {
			keyStr += ","
		}
		keyStr += "`" + v + "`"
	}
	length := len(key)
	dbq += keyStr + ")" + " VALUES "
	valuesStr := "("
	for i := range values {
		if i != 0 {
			if i%length == 0 {
				valuesStr += "),("
			} else {
				valuesStr += ","
			}
		}
		valuesStr += "?"
	}
	dbq += valuesStr + ")"
	if p.debug != nil {
		p.debug.Println("\n" + dbPrintStr(dbq, values))
	} else {
		fmt.Println("[AddRecord]", dbPrintStr(dbq, values))
	}

	stmt, err := p.db.Prepare(dbq)
	if err != nil {
		if p.debug != nil {
			p.debug.Printf("query faied, error:[%v]", err.Error())
		}
		return 0, err
	}

	result, err := stmt.Exec(values...)
	stmt.Close()
	if err != nil {
		if p.debug != nil {
			p.debug.Printf("data insert faied, error:[%v]", err.Error())
		}
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		if p.debug != nil {
			p.debug.Printf("insert faied, error:[%v]", err.Error())
		}
		return 0, err
	}
	if p.debug != nil {
		p.debug.Printf("insert success, last id:[%d]\n", id)
	}
	return id, nil
}

// UpdataRecord: 從SQL資料庫中修改指定的值
//
//	所有關鍵字除*以外需要用``包裹
//	`table`		string		從哪個表中查詢不需要``包裹
//	`updata`	string		需要修改的值，需要以,分割，例:`name`="aa",`age`=10
//	`where`		string		需要修改行的條件，例:`id`=10
//	`values`	...interface{}	額外的修改值
//	return int64 和 error，返回更新的行数
func (p *NyaMySQL) UpdateRecord(table string, updata string, where string, values ...interface{}) (int64, error) {
	var dbq string = "update `" + table + "` set " + updata
	if where != "" {
		dbq += " where " + where
	}
	if p.debug != nil {
		p.debug.Println("\n" + dbPrintStr(dbq, values))
	} else {
		fmt.Println("[UpdataRecord]", dbPrintStr(dbq, values))
	}

	stmt, err := p.db.Prepare(dbq)
	if err != nil {
		if p.debug != nil {
			p.debug.Printf("query faied, error:[%v]", err.Error())
		}
		return 0, err
	}
	result, err := stmt.Exec(values...)
	stmt.Close()
	if err != nil {
		if p.debug != nil {
			p.debug.Printf("update faied, error:[%v]", err.Error())
		} else {
			log.Printf("update faied, error:[%v]", err.Error())
		}
		return 0, err
	}
	num, _ := result.RowsAffected()
	if p.debug != nil {
		p.debug.Printf("update success, affected rows:[%d]\n", num)
	}
	return num, nil
}

// DeleteRecord: 從SQL資料庫中刪除行
//
//	所有關鍵字除*以外需要用``包裹
//	`table`		string		從哪個表中查詢不需要``包裹
//	`key`		string		根據哪個關鍵字刪除
//	`and`		string		刪除條件的附加條件會在語句末尾新增。可以寫入 and 或 or 或其他邏輯關鍵字以新增多個判斷條件
//	`values`	...interface{}	刪除條件的值
//	return		int64		刪除的行数
//	return		error		錯誤
func (p *NyaMySQL) DeleteRecord(table string, key string, and string, values ...interface{}) (int64, error) {
	var dbq string = fmt.Sprintf("delete from `%s` where `%s`", table, key)
	if len(values) <= 1 {
		dbq += fmt.Sprintf("=? %s", and)
	} else {
		wherein := ""
		length := len(values)
		andLen := 0
		for i := range and {
			if and[i:i+1] == "?" {
				andLen++
			}
		}
		length = length - andLen
		for i := 0; i < length; i++ {
			if wherein != "" {
				wherein += ","
			}
			wherein += "?"
		}
		dbq += fmt.Sprintf(" in (%s) %s", wherein, and)
	}
	//删除uid=2的数据
	if p.debug != nil {
		p.debug.Println("\n" + dbPrintStr(dbq, values))
	} else {
		fmt.Println("[DeleteRecord]", dbPrintStr(dbq, values))
	}
	stmt, err := p.db.Prepare(dbq)
	if err != nil {
		if p.debug != nil {
			p.debug.Printf("query faied, error:[%v]", err.Error())
		}
		return 0, err
	}
	result, err := stmt.Exec(values...)
	stmt.Close()
	if err != nil {
		if p.debug != nil {
			p.debug.Printf("delete faied, error:[%v]", err.Error())
		}
		return 0, err
	}
	num, _ := result.RowsAffected()
	if p.debug != nil {
		p.debug.Printf("delete success, affected rows:[%d]\n", num)
	}
	return num, nil
}

// 從 SQL 資料庫無主鍵表中刪除行
//
//	所有關鍵字除*以外需要用``包裹
//	`table`		string		從哪個表中查詢不需要``包裹
//	`keys`		[]string	根據哪個關鍵字刪除
//	`values`	...interface{}	刪除條件的值
func (p *NyaMySQL) DeleteRecordNoPK(table string, keys []string, values ...interface{}) error {
	if len(values)%len(keys) != 0 {
		return fmt.Errorf("'values'内容数量与'keys'不符")
	}
	var dbq string = "delete from `" + table + "` where ("
	where := ""
	len := len(values) / len(keys)
	for i := 0; i < len; i++ {
		if where != "" {
			where += ") OR ("
		}
		for ii, vv := range keys {
			if where != "" && ii != 0 {
				where += " AND "
			}
			where += "`" + vv + "`=?"
		}
	}
	dbq += where + ")"

	//删除uid=2的数据
	if p.debug != nil {
		p.debug.Println("\n" + dbPrintStr(dbq, values))
	} else {
		fmt.Println("[DeleteRecordNoPK]", dbPrintStr(dbq, values))
	}

	stmt, err := p.db.Prepare(dbq)
	if err != nil {
		if p.debug != nil {
			p.debug.Printf("query faied, error:[%v]", err.Error())
		}
		return err
	}
	result, err := stmt.Exec(values...)
	stmt.Close()
	if err != nil {
		if p.debug != nil {
			p.debug.Printf("delete faied, error:[%v]", err.Error())
		}
		return err
	}
	num, _ := result.RowsAffected()
	if p.debug != nil {
		p.debug.Printf("delete success, affected rows:[%d]\n", num)
	}
	return nil
}

// FreequeryData: 從SQL資料庫中尋找
//
//	所有關鍵字除*以外需要用``包裹
//	`sqldb`		string		MySQL 資料庫名，使用預設值只需填寫 ""
//	`sqlstr`	string		需要執行的SQL語句
//	`values`	...interface{}	指定log物件，沒有填寫nil
//	return   cmap.ConcurrentMap 和 error 物件，結構為：
//	{
//	    "0":{"id":1,"name":"1"},
//	    "1":{"id":2,"name":"2"}
//	}
func (p *NyaMySQL) FreequeryData(sqlstr string, values ...interface{}) (map[string]map[string]string, error) {
	if p.debug != nil {
		p.debug.Println("\n" + dbPrintStr(sqlstr, values))
	} else {
		fmt.Println("[FreequeryData]", dbPrintStr(sqlstr, values))
	}
	stmt, err := p.db.Prepare(sqlstr)
	if err != nil {
		if p.debug != nil {
			p.debug.Printf("query faied, error:[%v]", err.Error())
		}
		return map[string]map[string]string{}, err
	}
	query, err := stmt.Query(values...)
	stmt.Close()
	if err != nil {
		if p.debug != nil {
			p.debug.Printf("query faied, error:[%v]", err.Error())
		}
		return map[string]map[string]string{}, err
	}

	//读出查询出的列字段名
	cols, _ := query.Columns()
	//values是每个列的值，这里获取到byte里
	vals := make([][]byte, len(cols))
	//query.Scan的参数，因为每次查询出来的列是不定长的，用len(cols)定住当次查询的长度
	scans := make([]interface{}, len(cols))
	//让每一行数据都填充到[][]byte里面
	for i := range vals {
		scans[i] = &vals[i]
	}

	//最后得到的map
	results := map[string]map[string]string{}
	i := 0
	for query.Next() { //循环，让游标往下推
		if err := query.Scan(scans...); err != nil { //query.Scan查询出来的不定长值放到scans[i] = &values[i],也就是每行都放在values里
			if p.debug != nil {
				p.debug.Println(err)
			}
			return map[string]map[string]string{}, err
		}
		row := map[string]string{} //每行数据
		for k, v := range vals {   //每行数据是放在values里面，现在把它挪到row里
			key := cols[k]
			row[key] = string(v)
		}
		results[strconv.Itoa(i)] = row //装入结果集中
		i++
	}

	//关闭结果集（释放连接）
	query.Close()
	// fmt.Println("-----====-----")

	return results, nil
}

// handleQD: 處理查詢結果
func handleQD(query *sql.Rows, Deubg *log.Logger) (map[string]map[string]string, error) {
	//读出查询出的列字段名
	cols, _ := query.Columns()
	//values是每个列的值，这里获取到byte里
	values := make([][]byte, len(cols))
	//query.Scan的参数，因为每次查询出来的列是不定长的，用len(cols)定住当次查询的长度
	scans := make([]interface{}, len(cols))
	//让每一行数据都填充到[][]byte里面
	for i := range values {
		scans[i] = &values[i]
	}

	//最后得到的map
	results := map[string]map[string]string{}
	i := 0
	for query.Next() { //循环，让游标往下推
		if err := query.Scan(scans...); err != nil { //query.Scan查询出来的不定长值放到scans[i] = &values[i],也就是每行都放在values里
			if Deubg != nil {
				Deubg.Println(err)
			}
			return map[string]map[string]string{}, err
		}
		row := map[string]string{} //每行数据
		for k, v := range values { //每行数据是放在values里面，现在把它挪到row里
			key := cols[k]
			row[key] = string(v)
		}
		results[strconv.Itoa(i)] = row //装入结果集中
		i++
	}

	//关闭结果集（释放连接）
	query.Close()
	// fmt.Println("-----====-----")

	return results, nil
}

// dbPrintStr: 將SQL語句中的 ? 替換成實際值
func dbPrintStr(dbStr string, values []interface{}) string {
	for _, v := range values {
		dbStr = strings.Replace(dbStr, "?", fmt.Sprintf("'%v'", v), 1)
	}
	return dbStr
}
