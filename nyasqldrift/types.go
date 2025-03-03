package nyasqldrift

import (
	"regexp"
	"strings"
)

// typeMapping 是一個對映，用於將 MySQL 資料型別轉換為 SQLite 資料型別
var typeMapping = map[string]string{
	"TINYINT":   "INTEGER", // TINYINT 轉換為 INTEGER
	"SMALLINT":  "INTEGER", // SMALLINT 轉換為 INTEGER
	"MEDIUMINT": "INTEGER", // MEDIUMINT 轉換為 INTEGER
	"INT":       "INTEGER", // INT 轉換為 INTEGER
	"INTEGER":   "INTEGER", // INTEGER 轉換為 INTEGER
	"BIGINT":    "INTEGER", // BIGINT 轉換為 INTEGER
	"FLOAT":     "REAL",    // FLOAT 轉換為 REAL
	"DOUBLE":    "REAL",    // DOUBLE 轉換為 REAL
	"DECIMAL":   "REAL",    // DECIMAL 轉換為 REAL
	"CHAR":      "TEXT",    // CHAR 轉換為 TEXT
	"VARCHAR":   "TEXT",    // VARCHAR 轉換為 TEXT
	"TEXT":      "TEXT",    // TEXT 轉換為 TEXT
	"LONGTEXT":  "TEXT",    // LONGTEXT 轉換為 TEXT
	"DATE":      "TEXT",    // DATE 轉換為 TEXT
	"DATETIME":  "TEXT",    // DATETIME 轉換為 TEXT
	"TIMESTAMP": "TEXT",    // TIMESTAMP 轉換為 TEXT
	"TIME":      "TEXT",    // TIME 轉換為 TEXT
	"BOOL":      "INTEGER", // BOOL 轉換為 INTEGER
	"BOOLEAN":   "INTEGER", // BOOLEAN 轉換為 INTEGER
}

// reverseTypeMapping 是一個對映，用於將 SQLite 資料型別轉換為 MySQL 資料型別
var reverseTypeMapping = map[string]string{
	"INTEGER": "INT",          // INTEGER 轉換為 INT
	"REAL":    "DOUBLE",       // REAL 轉換為 DOUBLE
	"TEXT":    "VARCHAR(255)", // 可改为 TEXT，如需严格保守
	"BLOB":    "BLOB",         // BLOB 轉換為 BLOB
	"NUMERIC": "DECIMAL",      // NUMERIC 轉換為 DECIMAL
}

// replaceDataTypes 函式用於替換字串中的資料型別
//
// 引數:
//
//	line: 需要處理的字串
//
// 返回值:
//
//	返回處理後的字串
//
// 備註:
//
//	該函式使用正則表示式匹配字串中的資料型別，並根據 typeMapping 對映表將其替換為對應的 SQLite 資料型別。
func replaceDataTypes(line string) string {
	// 使用正則表示式匹配資料型別定義
	re := regexp.MustCompile(`(?i)(\w+)\s+([a-zA-Z]+)(\([0-9,]+\))?`)
	// 查詢匹配項
	matches := re.FindStringSubmatch(line)

	// 檢查是否至少有三個匹配項
	if len(matches) >= 3 {
		// 將資料型別轉換為大寫
		typ := strings.ToUpper(matches[2])
		// 初始化字尾為空字串
		suffix := ""
		// 檢查是否有第四個匹配項（即型別定義中的括號部分）
		if len(matches) >= 4 {
			suffix = matches[3]
		}

		// 在型別對映表中查詢匹配的資料型別
		sqliteType, ok := typeMapping[typ]
		// 如果找到對應的SQLite型別
		if ok {
			// 將原始型別替換為SQLite型別
			line = strings.Replace(line, typ+suffix, sqliteType, 1)
		}
	}

	return line
}

// restoreDataTypes 將給定字串中的資料型別替換為MySQL相容的資料型別
//
// 引數：
// line - 待處理的字串
//
// 返回值：
// 處理後的字串
func restoreDataTypes(line string) string {
	// 編譯正則表示式
	re := regexp.MustCompile(`(?i)(\w+)\s+([a-zA-Z]+)(\s|,|$)`)

	// 查詢匹配字串的子匹配
	matches := re.FindStringSubmatch(line)

	// 判斷子匹配的數量是否大於等於3
	if len(matches) >= 3 {
		// 將第二個子匹配轉換為大寫
		typ := strings.ToUpper(matches[2])

		// 從反向型別對映中獲取MySQL型別
		mysqlType, ok := reverseTypeMapping[typ]

		// 如果在反向型別對映中存在對應的MySQL型別
		if ok {
			// 使用正則表示式替換原始字串中的型別
			line = re.ReplaceAllString(line, "${1} "+mysqlType+"${3}")
		}
	}

	return line
}
