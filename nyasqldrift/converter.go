package nyasqldrift

import (
	"regexp"
	"strings"
)

// convertCreateTableMySQLToSQLite 將 MySQL 風格的 SQL 語句轉換為 SQLite 風格的 SQL 語句
//
// 引數:
//
//	mysqlSQL: MySQL 風格的 SQL 語句
//
// 返回值:
//
//	string: 轉換後的 SQLite 風格的 SQL 語句
//	error: 如果發生錯誤，則返回錯誤資訊
func convertCreateTableMySQLToSQLite(mysqlSQL string) (string, error) {
	// 編譯正則表示式，用於匹配以 ") ENGINE=" 開頭的字串，並忽略大小寫
	reEngine := regexp.MustCompile(`(?i)\)\s*ENGINE=.*?;`)
	// 使用正則表示式替換掉原始 SQL 語句中的 ENGINE 部分
	mysqlSQL = reEngine.ReplaceAllString(mysqlSQL, ");")

	// 按行分割 SQL 語句
	lines := strings.Split(mysqlSQL, "\n")
	// 遍歷每一行
	for i, line := range lines {
		// 去除行首尾的空白字元
		line = strings.TrimSpace(line)
		// 如果行是以 "CREATE TABLE" 開頭（忽略大小寫），或者行為空，或者行為 ");"，則跳過該行
		if strings.HasPrefix(strings.ToUpper(line), "CREATE TABLE") || line == "" || line == ");" {
			continue
		}
		// 替換資料型別
		line = replaceDataTypes(line)
		// 如果行中包含 "AUTO_INCREMENT"（忽略大小寫）
		if strings.Contains(strings.ToUpper(line), "AUTO_INCREMENT") {
			// 替換掉 "AUTO_INCREMENT"
			line = strings.ReplaceAll(line, "AUTO_INCREMENT", "")
			line = strings.ReplaceAll(line, "auto_increment", "")
			// 如果行中不包含 "PRIMARY KEY"（忽略大小寫）
			if !strings.Contains(strings.ToUpper(line), "PRIMARY KEY") {
				// 在行末新增 " PRIMARY KEY AUTOINCREMENT"
				line += " PRIMARY KEY AUTOINCREMENT"
			} else {
				// 替換 "PRIMARY KEY" 為 "PRIMARY KEY AUTOINCREMENT"
				line = strings.ReplaceAll(line, "PRIMARY KEY", "PRIMARY KEY AUTOINCREMENT")
			}
		}

		// 更新行內容
		lines[i] = line
	}
	// 將修改後的行重新拼接成完整的 SQL 語句
	newSQL := strings.Join(lines, "\n")
	return newSQL, nil
}

// convertCreateTableSQLiteToMySQL 將 SQLite 的建立表語句轉換為 MySQL 相容的建立表語句
// 引數：
// sqliteSQL：要轉換的 SQLite 建立表語句
// 返回值：
// string：轉換後的 MySQL 建立表語句
// error：可能出現的錯誤
func convertCreateTableSQLiteToMySQL(sqliteSQL string) (string, error) {
	// 將輸入的SQLite SQL語句按行分割
	lines := strings.Split(sqliteSQL, "\n")
	for i, line := range lines {
		// 去除每行首尾的空白字元
		line = strings.TrimSpace(line)
		// 如果當前行以"CREATE TABLE"開頭，或者為空行，或者為以");"結尾的行，則跳過處理
		if strings.HasPrefix(strings.ToUpper(line), "CREATE TABLE") || line == "" || line == ");" {
			continue
		}
		// 恢復資料型別
		line = restoreDataTypes(line)
		// 如果當前行包含"AUTOINCREMENT"
		if strings.Contains(strings.ToUpper(line), "AUTOINCREMENT") {
			// 將"AUTOINCREMENT"替換為空字串
			line = strings.ReplaceAll(line, "AUTOINCREMENT", "")
			// 如果當前行包含"PRIMARY KEY"
			if strings.Contains(strings.ToUpper(line), "PRIMARY KEY") {
				// 將"PRIMARY KEY"替換為"AUTO_INCREMENT PRIMARY KEY"
				line = strings.ReplaceAll(line, "PRIMARY KEY", "AUTO_INCREMENT PRIMARY KEY")
			} else {
				// 在當前行末尾新增" AUTO_INCREMENT PRIMARY KEY"
				line += " AUTO_INCREMENT PRIMARY KEY"
			}
		}

		// 更新lines陣列中的當前行
		lines[i] = line
	}
	// 將處理後的行重新拼接成完整的SQL語句
	newSQL := strings.Join(lines, "\n")
	// 去除newSQL末尾的分號
	newSQL = strings.TrimSuffix(newSQL, ";")
	// 在newSQL末尾新增" ENGINE=InnoDB DEFAULT CHARSET=utf8;"
	newSQL += " ENGINE=InnoDB DEFAULT CHARSET=utf8;"
	return newSQL, nil
}
