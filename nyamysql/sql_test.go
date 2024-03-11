package nyamysql

import (
	"fmt"
	"testing"
)

var mysqlconfig string = `{
	"mysql_user": "dbuser",
	"mysql_pwd": "dbpwd",
	"mysql_addr": "127.0.0.1",
	"mysql_port": "3306",
	"mysql_db": "test",
	"mysql_limit": "100"
}`

func TestQueryCMD(t *testing.T) {
	nyaMS := New(mysqlconfig)
	if nyaMS.Error() != nil {
		fmt.Println("MySQL DB Link error:", nyaMS.Error().Error())
		return
	}
	sqls := "select * from `test` where `id`=?;select * from `region` where `id`=?"
	values := [][]interface{}{{1}, {3}}

	qd, err := nyaMS.QueryDataCMD(sqls, values...)
	if err != nil {
		fmt.Println("QueryDataCMD error:", err.Error())
		return
	}
	fmt.Println(qd)
}

func TestQueryData(t *testing.T) {
	nyaMS := New(mysqlconfig)
	if nyaMS.Error() != nil {
		fmt.Println("MySQL DB Link error:", nyaMS.Error().Error())
		return
	}
	where := "`id`=?"
	oreder := "`id` asc"
	limit := "0,10"
	values := []interface{}{"1"}
	qd, err := nyaMS.QueryData("*", "test", where, oreder, limit, values...)
	if err != nil {
		fmt.Println("QueryData error:", err.Error())
		return
	}
	fmt.Println(qd)
}

func TestAdd(t *testing.T) {
	nyaMS := New(mysqlconfig)
	if nyaMS.Error() != nil {
		fmt.Println("MySQL DB Link error:", nyaMS.Error().Error())
		return
	}
	key := "`id`,`name`"
	value := "?,?"
	values := []interface{}{1, "test"}
	_, err := nyaMS.AddRecord("test", key, value, "", values...)
	if err != nil {
		fmt.Println("Add error:", err.Error())
		return
	}
}

func TestAdds(t *testing.T) {
	nyaMS := New(mysqlconfig)
	if nyaMS.Error() != nil {
		fmt.Println("MySQL DB Link error:", nyaMS.Error().Error())
		return
	}
	key := "`id`,`name`"
	value := "(?,?),(?,?)"
	values := []interface{}{2, "test1", 3, "test2"}
	_, err := nyaMS.AddRecord("test", key, "", value, values...)
	if err != nil {
		fmt.Println("Add error:", err.Error())
		return
	}
}

func TestUpdate(t *testing.T) {
	nyaMS := New(mysqlconfig)
	if nyaMS.Error() != nil {
		fmt.Println("MySQL DB Link error:", nyaMS.Error().Error())
		return
	}
	set := "`name`=?"
	where := "`id`=?"
	values := []interface{}{"test3", 3}
	_, err := nyaMS.UpdateRecord("test", set, where, values...)
	if err != nil {
		fmt.Println("Update error:", err.Error())
		return
	}
}

func TestDelete(t *testing.T) {
	nyaMS := New(mysqlconfig)
	if nyaMS.Error() != nil {
		fmt.Println("MySQL DB Link error:", nyaMS.Error().Error())
		return
	}
	wherein := "?,?,?"
	and := "or `name`=?"
	values := []interface{}{3, 5, 6, "test2"}
	result, err := nyaMS.DeleteRecord("test", "id", wherein, and, values...)
	if err != nil {
		fmt.Println("Delete error:", err.Error())
		return
	}
	fmt.Println(result)
}

func TestDeleteNoPK(t *testing.T) {
	nyaMS := New(mysqlconfig)
	if nyaMS.Error() != nil {
		fmt.Println("MySQL DB Link error:", nyaMS.Error().Error())
		return
	}
	keys := []string{"id", "name"}
	values := []interface{}{3, "test3", 4, "test4"}
	err := nyaMS.DeleteRecordNoPK("test", keys, values...)
	if err != nil {
		fmt.Println("Delete error:", err.Error())
		return
	}
}
