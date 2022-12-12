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

// QueryDataCMD: 從SQL資料庫中查詢
//
//	所有關鍵字除*以外需要用``包裹
//	`sql`	string		mysql语句
//	`Debug`	*log.Logger	指定log物件，沒有填寫nil
//	return cmap.ConcurrentMap 和 error 物件，結構為：
//	{
//	    "0":{"id":1,"name":"1"},
//	    "1":{"id":2,"name":"2"}
//	}
func (p *NyaMySQL) QueryDataCMD(sql string, Debug *log.Logger) (map[string]map[string]string, error) {
	sqls := strings.Split(sql, ";")
	for i, v := range sqls {
		if Debug != nil {
			Debug.Println("\n" + v)
		} else {
			fmt.Println("[QueryData]", v)
		}
		if i+1 == len(sqls) {
			query, err := p.db.Query(v)
			if err != nil {
				if Debug != nil {
					Debug.Printf("query faied, error:[%v]", err.Error())
				}
				return map[string]map[string]string{}, err
			}
			return handleQD(query, Debug)
		} else {
			_, err := p.db.Exec(v)
			if err != nil {
				if Debug != nil {
					Debug.Printf("query faied, error:[%v]", err.Error())
				}
				return map[string]map[string]string{}, err
			}
		}
	}
	return map[string]map[string]string{}, fmt.Errorf("query is null")
}

// QueryData: 從SQL資料庫中查詢
//
//	所有關鍵字除*以外需要用``包裹
//	`sqldb`   string      mysql資料庫名，使用預設值只需填寫""
//	`recn`    string      查詢語句的返回。全部：*，指定：`id`
//	`table`   string      從哪個表中查詢，此處可以使用關聯語句
//	`where`   string      where語句部分，最前方不需要填寫where，例：`id`=1
//	`orderby` string      排序，例：`id` ASC/DESC
//	`limit`   string      分頁，例：1,10
//	`Debug`   *log.Logger 指定log物件，沒有填寫nil
//	return cmap.ConcurrentMap 和 error 物件，結構為：
//	{
//	    "0":{"id":1,"name":"1"},
//	    "1":{"id":2,"name":"2"}
//	}
func (p *NyaMySQL) QueryData(recn string, table string, where string, orderby string, limit string, Debug *log.Logger) (map[string]map[string]string, error) {
	var dbq string = "select " + recn + " from `" + table + "`"
	if where != "" {
		dbq += " where " + where
	}
	if orderby != "" {
		dbq += " ORDER BY " + orderby
	}
	if limit != "" {
		dbq += " limit " + limit
	}
	if Debug != nil {
		Debug.Println("\n" + dbq)
	} else {
		fmt.Println("[QueryData]", dbq)
	}
	query, err := p.db.Query(dbq)
	if err != nil {
		if Debug != nil {
			Debug.Printf("query faied, error:[%v]", err.Error())
		}
		return map[string]map[string]string{}, err
	}
	return handleQD(query, Debug)
}

func handleQD(query *sql.Rows, Debug *log.Logger) (map[string]map[string]string, error) {
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
			if Debug != nil {
				Debug.Println(err)
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

// AddRecord: 向SQL資料庫中新增
//
//	所有關鍵字除*以外需要用``包裹
//	`sqldb`  string      mysql資料庫名，使用預設值只需填寫""
//	`table`  string      從哪個表中查詢不需要``包裹
//	`key`    string      需要新增的列，需要以,分割
//	`val`    string      與key對應的值，以,分割
//	`values` string      (此項不為"",val無效)新增多行資料與key對應的值，以,分割,例(1,2),(2,3)
//	`Debug`  *log.Logger 指定log物件，沒有填寫nil
//	return int64 和 error 物件，返回新增行的id
func (p *NyaMySQL) AddRecord(table string, key string, val string, values string, Debug *log.Logger) (int64, error) {
	var dbq string = "insert into `" + table + "` (" + key + ")" + "VALUES "
	if values != "" {
		dbq += values
	} else {
		dbq += "(" + val + ")"
	}
	if Debug != nil {
		Debug.Println("\n" + dbq)
	} else {
		fmt.Println("[AddRecord]", dbq)
	}
	result, err := p.db.Exec(dbq)
	if err != nil {
		if Debug != nil {
			Debug.Printf("data insert faied, error:[%v]", err.Error())
		}
		return 0, err
	}
	id, _ := result.LastInsertId()
	if Debug != nil {
		Debug.Printf("insert success, last id:[%d]\n", id)
	}
	return id, nil
}

// UpdataRecord: 從SQL資料庫中修改指定的值
//
//	所有關鍵字除*以外需要用``包裹
//	`sqldb`  string mysql資料庫名，使用預設值只需填寫""
//	`table`  從哪個表中查詢不需要``包裹
//	`updata` 需要修改的值，需要以,分割，例:`name`="aa",`age`=10
//	`where`  需要修改行的條件，例:`id`=10
//	`Debug`  *log.Logger 指定log物件，沒有填寫nil
//	return int64 和 error，返回更新的行数
func (p *NyaMySQL) UpdataRecord(table string, updata string, where string, Debug *log.Logger) (int64, error) {
	var dbq string = "update `" + table + "` set " + updata
	if where != "" {
		dbq += " where " + where
	}
	if Debug != nil {
		Debug.Println("\n" + dbq)
	} else {
		fmt.Println("[UpdataRecord]", dbq)
	}
	//更新uid=1的username
	result, err := p.db.Exec(dbq)
	if err != nil {
		if Debug != nil {
			Debug.Printf("update faied, error:[%v]", err.Error())
		} else {
			log.Printf("update faied, error:[%v]", err.Error())
		}
		return 0, err
	}
	num, _ := result.RowsAffected()
	if Debug != nil {
		Debug.Printf("update success, affected rows:[%d]\n", num)
	}
	return num, nil
}

// DeleteRecord: 從SQL資料庫中刪除行
//
//	所有關鍵字除*以外需要用``包裹
//	`sqldb`   string mysql資料庫名，使用預設值只需填寫""
//	`table`   string 從哪個表中查詢不需要``包裹
//	`key`     string 根據哪個關鍵字刪除
//	`value`   string 關鍵字對應的值
//	`and`     string 刪除條件的附加條件會在語句末尾新增。可以寫入 and 或 or 或其他邏輯關鍵字以新增多個判斷條件
//	`wherein` string 以where xx in (wherein)的方式刪除。wherein不為""時value無效，and 仍然有效
//	`Debug`   *log.Logger 指定log物件，沒有填寫nil
//	return    error
func (p *NyaMySQL) DeleteRecord(table string, key string, value string, and string, wherein string, Debug *log.Logger) error {
	if value == "" && wherein == "" {
		return fmt.Errorf("删除语句条件不能为空")
	}
	var dbq string = "delete from `" + table + "` where `"
	if wherein == "" {
		dbq += key + "`='" + value + "'" + and
	} else {
		dbq += key + "` in (" + wherein + ")" + and
	}
	//删除uid=2的数据
	if Debug != nil {
		Debug.Println("\n" + dbq)
	} else {
		fmt.Println("[DeleteRecord]", dbq)
	}
	result, err := p.db.Exec(dbq)
	if err != nil {
		if Debug != nil {
			Debug.Printf("delete faied, error:[%v]", err.Error())
		}
		return err
	}
	num, _ := result.RowsAffected()
	if Debug != nil {
		Debug.Printf("delete success, affected rows:[%d]\n", num)
	}
	return nil
}

// 從 SQL 資料庫無主鍵表中刪除行
//
//	所有關鍵字除*以外需要用``包裹
//	`sqldb`  string     MySQL 資料庫名，使用預設值只需填寫 ""
//	`table`  string     從哪個表中查詢不需要``包裹
//	`keys`   []string   根據哪個關鍵字刪除
//	`values` [][]string 關鍵字對應的值
func (p *NyaMySQL) DeleteRecordNoPK(table string, keys []string, values [][]string, Debug *log.Logger) error {
	for _, v := range values {
		if len(v) != len(keys) {
			return fmt.Errorf("'values'内容数量与'keys'不符")
		}
	}
	var dbq string = "delete from `" + table + "` where ("
	where := ""
	for _, v := range values {
		if where != "" {
			where += ") OR ("
		}
		for ii, vv := range keys {
			if where != "" && ii != 0 {
				where += " AND "
			}
			where += "`" + vv + "`='" + v[ii] + "'"
		}
	}
	dbq += where + ")"
	//删除uid=2的数据
	if Debug != nil {
		Debug.Println("\n" + dbq)
	} else {
		fmt.Println("[DeleteRecord]", dbq)
	}
	result, err := p.db.Exec(dbq)
	if err != nil {
		if Debug != nil {
			Debug.Printf("delete faied, error:[%v]", err.Error())
		}
		return err
	}
	num, _ := result.RowsAffected()
	if Debug != nil {
		Debug.Printf("delete success, affected rows:[%d]\n", num)
	}
	return nil
}

// FreequeryData: 從SQL資料庫中尋找
//
//	所有關鍵字除*以外需要用``包裹
//	`sqldb`  string      MySQL 資料庫名，使用預設值只需填寫 ""
//	`sqlstr` string      需要執行的SQL語句
//	`Debug`  *log.Logger 指定log物件，沒有填寫nil
//	return   cmap.ConcurrentMap 和 error 物件，結構為：
//	{
//	    "0":{"id":1,"name":"1"},
//	    "1":{"id":2,"name":"2"}
//	}
func (p *NyaMySQL) FreequeryData(sqlstr string, Debug *log.Logger) (map[string]map[string]string, error) {
	if Debug != nil {
		Debug.Println("\n" + sqlstr)
	} else {
		fmt.Println("[FreequeryData]", sqlstr)
	}
	query, err := p.db.Query(sqlstr)
	if err != nil {
		if Debug != nil {
			Debug.Printf("query faied, error:[%v]", err.Error())
		}
		return map[string]map[string]string{}, err
	}

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
			if Debug != nil {
				Debug.Println(err)
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
