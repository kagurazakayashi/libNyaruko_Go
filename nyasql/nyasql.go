package nyasql

import (
	"strings"
)

//clear: 轉義清理敏感字元
//	data string 要檢測的字串
//	mode int8   遇到敏感字元時: -1刪除 0返回空字串 1加入跳脫字元`\` 2加入一個相同字元
//	return 新增跳脫字元後的字串
//示例: fmt.Println(clear("a\"aa\"", 1)) -> a\"aa\"
func clear(data string, mode int8) string {
	var ndata string = data
	var needEscape []string = []string{"'", "\"", "`"}
	for _, escape := range needEscape {
		if mode == 0 {
			return ""
		} else if mode == -1 {
			ndata = strings.Replace(ndata, escape, "", -1)
		} else if mode == 1 {
			ndata = strings.Replace(ndata, escape, "\\"+escape, -1)
		} else if mode == 2 {
			ndata = strings.Replace(ndata, escape, escape+escape, -1)
		}
	}
	return ndata
}

//quotationMark: 在字串左右新增引號，並轉義已有引號
//	data   string 要處理的字串
//	return string 處理後的字串
func quotationMark(data string, addChar string) string {
	return addChar + clear(data, 2) + addChar
}

//SqlWhereOption: 合成條件語句 WHERE 單元 的可選引數
type SqlWhereOption struct {
	compare  []string // 比較時使用的符號
	relation []string // 條件之間的關係
}
type SqlWhereOptionT func(*SqlWhereOption)

func SqlWhereOption_compare(v []string) SqlWhereOptionT {
	return func(q *SqlWhereOption) {
		q.compare = v
	}
}
func SqlWhereOption_relation(v []string) SqlWhereOptionT {
	return func(q *SqlWhereOption) {
		q.relation = v
	}
}

//SqlWhere: 條件表示式生成
//	`table`     string 資料表名稱
//	`condition` string 要處理的鍵值字典
//	`options`   ...SqlWhereOptionT 可選配置，執行 `SqlWhereOption_*` 函式輸入
//		`compare` []string 比較時使用的符號，預設為 ["="] 。陣列長度只有 1 時，所有表示式都用此符號；陣列長度等於輸入字典長度時，則視為每個表示式單獨決定。
//		`relation` []string 條件之間的關係，預設為 ["AND"] 。
//	return string 生成的 SQL 語句片段
//示例: fmt.Println(SqlWHERE("table1", map[string]string{"aaa": "111", "bbb": "222"}))
//	-> WHERE `table1`.`aaa`="111" AND `table1`.`bbb`="222"
func SqlWHERE(table string, condition map[string]string, options ...SqlWhereOptionT) string {
	option := &SqlWhereOption{compare: []string{"="}, relation: []string{"AND"}}
	for _, o := range options {
		o(option)
	}
	var conditionLength = len(condition)
	var compareOne bool = len(option.compare) != conditionLength
	var relationOne bool = len(option.relation) != conditionLength-1
	var formulas string = "WHERE "
	var i int = 0
	for key, val := range condition {
		var nKey string = quotationMark(table, "`") + "." + quotationMark(key, "`")
		var nVal string = quotationMark(val, "\"")
		var compare string = option.compare[0]
		if !compareOne {
			compare = option.compare[i]
		}
		var formula string = nKey + compare + nVal
		if i > 0 {
			var relation string = option.relation[0]
			if !relationOne {
				relation = option.relation[i-1]
			}
			formulas += " " + relation + " "
		}
		formulas += formula
		i++
	}
	return formulas
}

//SqlINSERT: 插入表示式生成
//	`table` string 資料表名稱
//	`data`  string 要插入的鍵值字典
//	return  string 生成的 SQL 語句片段
//示例: fmt.Println(SqlINSERT("table1", map[string]string{"aaa": "111", "bbb": "222"}))
//	-> INSERT INTO `table1` (`aaa`,`bbb`) VALUES ("111","222")
func SqlINSERT(table string, data map[string]string) string {
	var allKey string = ""
	var allVal string = ""
	var first bool = true
	for key, val := range data {
		if !first {
			allKey += ","
			allVal += ","
		} else {
			first = false
		}
		allKey += quotationMark(key, "`")
		allVal += quotationMark(val, "\"")
	}
	return "INSERT INTO `" + table + "` (" + allKey + ") VALUES (" + allVal + ")"
}

//SqlORDERBY: 排序表示式生成
//	`column` []string 要進行排序的列名
//	`isDESC` []bool   是否降序排列: true:降序(從大到小), false:升序(從小到大)。陣列長度只有 1 時，所有排序方式都為它；陣列長度等於輸入陣列長度時，則視為每個表示式單獨決定。
//	return   string 生成的 SQL 語句片段
//示例: fmt.Println(SqlORDERBY([]string{"column1", "column2"}, []bool{true}))
//-> ORDER BY `column1` DESC, `column2` DESC
func SqlORDERBY(column []string, isDESC []bool) string {
	var order string = "ORDER BY "
	var descOne bool = len(column) != len(isDESC)
	for i, v := range column {
		if i > 0 {
			order += ", "
		}
		var desc bool = isDESC[0]
		if !descOne {
			desc = isDESC[i]
		}
		var sc string = ""
		if desc {
			sc = "DESC"
		} else {
			sc = "ASC"
		}
		order += quotationMark(v, "`") + " " + sc
	}
	return order
}

//SqlDELETE: 刪除表示式生成
//	`table` string 資料表名稱
//	return  string 生成的 SQL 語句片段
//示例: fmt.Println(SqlDELETE("table1"))
//-> DELETE FROM `table1`
func SqlDELETE(table string) string {
	return "DELETE FROM `" + table + "` "
}

//SqlEnd: 結束 SQL 語句
//	`sqlCmd` string SQL 語句
//	return 結束的 SQL 語句
func SqlEnd(sqlCmd []string) []string {
	for i := 0; i < len(sqlCmd); i++ {
		sqlCmd[i] += ";"
	}
	return sqlCmd
}
