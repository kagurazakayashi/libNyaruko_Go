package nyamysql

import (
	"fmt"
	"log"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	cmap "github.com/orcaman/concurrent-map"
)

//从SQL数据库中查找
//	所有关键字除*以外需要用``包裹
//	`sqldb`		string		mysql数据库名，使用默认值只需填写""
//	`recn`		string		查询语句的返回。全部：*，指定：`id`
//	`table`		string		从哪个表中查询，此处可以使用关联语句
//	`where`		string		where语句部分，最前方不需要填写where，例：`id`=1
//	`orderby`	string		排序，例：`id` ASC/DESC
//	`limit`		string		分页，例：1,10
//	`Debug`		*log.Logger	指定log对象，没有填写nil
//	return		cmap.ConcurrentMap 和 error 对象，结构为：
//	{
//		"0":{"id":1,"name":"1"},
//		"1":{"id":2,"name":"2"}
//	}
// func QueryData(sqldb string, recn string, table string, where string, orderby string, limit string, Debug *log.Logger) (cmap.ConcurrentMap, error) {
// 	dbitem, err := getRWSign(true, true)
// 	if err != nil {
// 		return nil, err
// 	}
// 	db, err := Linkmysql(true, dbitem, LSQLwithSQLDB(sqldb))
// 	if err != nil {
// 		return nil, err
// 	}
// 	redata, err := queryData(db, recn, table, where, orderby, limit, Debug)
// 	db.Close()
// 	return redata, err
// }
func (p *NyaMySQL) QueryData(recn string, table string, where string, orderby string, limit string, Debug *log.Logger) (cmap.ConcurrentMap, error) {
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
	}
	query, err := p.db.Query(dbq)
	if err != nil {
		Debug.Printf("query faied, error:[%v]", err.Error())
		return cmap.New(), err
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
	results := cmap.New()
	i := 0
	for query.Next() { //循环，让游标往下推
		if err := query.Scan(scans...); err != nil { //query.Scan查询出来的不定长值放到scans[i] = &values[i],也就是每行都放在values里
			Debug.Println(err)
			return cmap.New(), err
		}
		row := cmap.New()          //每行数据
		for k, v := range values { //每行数据是放在values里面，现在把它挪到row里
			key := cols[k]
			row.Set(key, string(v))
		}
		results.Set(strconv.Itoa(i), row) //装入结果集中
		i++
	}

	//关闭结果集（释放连接）
	query.Close()
	// fmt.Println("-----====-----")

	return results, nil
}

//向SQL数据库中添加
//	所有关键字除*以外需要用``包裹
//	`sqldb`		string		mysql数据库名，使用默认值只需填写""
//	`table`		string		从哪个表中查询不需要``包裹
//	`key`		string		需要添加的列，需要以,分割
//	`val`		string		与key对应的值，以,分割
//	`values`	string		(此项不为"",val无效)添加多行数据与key对应的值，以,分割,例(1,2),(2,3)
//	`Debug`		*log.Logger	指定log对象，没有填写nil
//	return		int64 和 error 对象，返回添加行的id
// func AddRecord(sqldb string, table string, key string, val string, values string, Debug *log.Logger) (int64, error) {
// 	dbitem, err := getRWSign(true, false)
// 	if err != nil {
// 		return 0, err
// 	}
// 	db, err := Linkmysql(false, dbitem, LSQLwithSQLDB(sqldb))
// 	if err != nil {
// 		return 0, err
// 	}
// 	redata, err := addRecord(db, table, key, val, values, Debug)
// 	db.Close()
// 	return redata, err
// }
func (p *NyaMySQL) AddRecord(table string, key string, val string, values string, Debug *log.Logger) (int64, error) {
	var dbq string = "insert into `" + table + "` (" + key + ")" + "VALUES "
	if values != "" {
		dbq += values
	} else {
		dbq += "(" + val + ")"
	}
	if Debug != nil {
		Debug.Println("\n" + dbq)
	}
	result, err := p.db.Exec(dbq)
	if err != nil {
		if Debug != nil {
			Debug.Printf("data insert faied, error:[%v]", err.Error())
		} else {
			log.Printf("data insert faied, error:[%v]", err.Error())
		}
		return 0, err
	}
	id, _ := result.LastInsertId()
	if Debug != nil {
		Debug.Printf("insert success, last id:[%d]\n", id)
	}
	return id, nil
}

//从SQL数据库中修改指定的值
//	所有关键字除*以外需要用``包裹
//	`sqldb`		string		mysql数据库名，使用默认值只需填写""
//	`table`:	从哪个表中查询不需要``包裹
//	`updata`:	需要修改的值，需要以,分割，例:`name`="aa",`age`=10
//	`where`:	需要修改行的条件，例：`id`=10
//	`Debug`		*log.Logger		指定log对象，没有填写nil
//	return		error
// func UpdataRecord(sqldb string, table string, updata string, where string, Debug *log.Logger) error {
// 	dbitem, err := getRWSign(true, false)
// 	if err != nil {
// 		return err
// 	}
// 	db, err := Linkmysql(false, dbitem, LSQLwithSQLDB(sqldb))
// 	if err != nil {
// 		return err
// 	}
// 	err = updataRecord(db, table, updata, where, Debug)
// 	db.Close()
// 	return err
// }
func (p *NyaMySQL) UpdataRecord(table string, updata string, where string, Debug *log.Logger) error {
	var dbq string = "update `" + table + "` set " + updata
	if where != "" {
		dbq += " where " + where
	}
	if Debug != nil {
		Debug.Println("\n" + dbq)
	}
	//更新uid=1的username
	result, err := p.db.Exec(dbq)
	if err != nil {
		if Debug != nil {
			Debug.Printf("update faied, error:[%v]", err.Error())
		} else {
			log.Printf("update faied, error:[%v]", err.Error())
		}
		return err
	}
	num, _ := result.RowsAffected()
	if Debug != nil {
		Debug.Printf("update success, affected rows:[%d]\n", num)
	}
	return nil
}

//从SQL数据库中删除行
//	所有关键字除*以外需要用``包裹
//	`sqldb`		string		mysql数据库名，使用默认值只需填写""
//	`table`		string		从哪个表中查询不需要``包裹
//	`key`		string		根据哪个关键字删除
//	`value`		string		关键字对应的值
//	`and`		string		删除条件的附加条件会在语句末尾添加。可以写入 and 或 or 或其他逻辑关键字以添加多个判断条件
//	`wherein`	string		以where xx in (wherein)的方式删除。wherein不为""时value无效，and 仍然有效
//	`Debug`		*log.Logger	指定log对象，没有填写nil
//	return		error
// func DeleteRecord(sqldb string, table string, key string, value string, and string, wherein string, Debug *log.Logger) error {
// 	dbitem, err := getRWSign(true, false)
// 	if err != nil {
// 		return err
// 	}
// 	db, err := Linkmysql(false, dbitem, LSQLwithSQLDB(sqldb))
// 	if err != nil {
// 		return err
// 	}
// 	err = deleteRecord(db, table, key, value, and, wherein, Debug)
// 	db.Close()
// 	return err
// }
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
	}
	result, err := p.db.Exec(dbq)
	if err != nil {
		if Debug != nil {
			Debug.Printf("delete faied, error:[%v]", err.Error())
		} else {
			log.Printf("delete faied, error:[%v]", err.Error())
		}
		return err
	}
	num, _ := result.RowsAffected()
	if Debug != nil {
		Debug.Printf("delete success, affected rows:[%d]\n", num)
	}
	return nil
}

//从SQL数据库中查找
//	所有关键字除*以外需要用``包裹
//	`sqldb`		string		mysql数据库名，使用默认值只需填写""
//	`sqlstr`	string		需要执行的SQL语句
//	`Debug`		*log.Logger	指定log对象，没有填写nil
//	return		cmap.ConcurrentMap 和 error 对象，结构为：
//	{
//		"0":{"id":1,"name":"1"},
//		"1":{"id":2,"name":"2"}
//	}
// func FreeQueryData(sqldb string, sqlstr string, Debug *log.Logger) (cmap.ConcurrentMap, error) {
// 	dbitem, err := getRWSign(true, true)
// 	if err != nil {
// 		return nil, err
// 	}
// 	db, err := Linkmysql(true, dbitem, LSQLwithSQLDB(sqldb))
// 	if err != nil {
// 		return nil, err
// 	}
// 	redata, err := freequeryData(db, sqlstr, Debug)
// 	db.Close()
// 	return redata, err
// }
func (p *NyaMySQL) FreequeryData(sqlstr string, Debug *log.Logger) (cmap.ConcurrentMap, error) {
	if Debug != nil {
		Debug.Println("\n" + sqlstr)
	}
	query, err := p.db.Query(sqlstr)
	if err != nil {
		Debug.Printf("query faied, error:[%v]", err.Error())
		return cmap.New(), err
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
	results := cmap.New()
	i := 0
	for query.Next() { //循环，让游标往下推
		if err := query.Scan(scans...); err != nil { //query.Scan查询出来的不定长值放到scans[i] = &values[i],也就是每行都放在values里
			Debug.Println(err)
			return cmap.New(), err
		}
		row := cmap.New()          //每行数据
		for k, v := range values { //每行数据是放在values里面，现在把它挪到row里
			key := cols[k]
			row.Set(key, string(v))
		}
		results.Set(strconv.Itoa(i), row) //装入结果集中
		i++
	}

	//关闭结果集（释放连接）
	query.Close()
	// fmt.Println("-----====-----")

	return results, nil
}
