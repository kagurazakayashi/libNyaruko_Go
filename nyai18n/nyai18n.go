// 多語言支援
package nyai18n

import (
	"bufio"
	"os"
	"strings"

	"github.com/cloudfoundry/jibber_jabber"
)

var (
	// 語言
	Language string = "auto"
	// 語言映射
	languageMap map[string]string
)

// 初始化語言檔案
// 注意：基於 Language 變數，如果 Language 為空或為 auto，則會自動設定為系統語言，如果找不到對應的語言，則使用第一個語言
// `languageFile` string 語言檔案路徑
// `isReload` bool 是否重新載入，使用時用 false 即可
// 語言檔案格式: `任意檔名.ini`
// ```ini
// [en-US]
// key1=value1
// key2=value2
// [zh-CN]
// key1=值1
// key2=值2
// ; 註釋
// ```
func LoadLanguageFile(languageFile string, isReload bool) error {
	// 檢查語言是否為空，如果是則自動設定
	if len(Language) == 0 || Language == "auto" {
		AutoSetLanguage()
	}

	// 開啟檔案
	file, err := os.Open(languageFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// 建立一個讀取器
	scanner := bufio.NewScanner(file)

	// 用來儲存結果的 map
	languageMap = make(map[string]string)

	// 用來標記是否處於目標段
	inSection := false

	// 逐行讀取檔案
	var firstSection string = ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 檢查是否是新的節開始
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// 檢查當前節名是否為目標節名
			currentSection := line[1 : len(line)-1]
			if len(firstSection) == 0 {
				firstSection = currentSection
			}
			inSection = (currentSection == Language)
		} else if inSection && len(line) > 0 && line[0] != ';' {
			// 解析鍵值對
			if kv := strings.SplitN(line, "=", 2); len(kv) == 2 {
				key := strings.TrimSpace(kv[0])
				value := strings.TrimSpace(kv[1])
				languageMap[key] = value
			}
		}
	}

	// 檢查掃描是否有錯誤
	if err := scanner.Err(); err != nil {
		return err
	}

	// 如果沒有找到目標段，則使用第一個段
	if len(languageMap) == 0 && !isReload {
		Language = firstSection
		LoadLanguageFile(languageFile, true)
	}
	return nil
}

// 自動設定語言
func AutoSetLanguage() string {
	syslang, err := jibber_jabber.DetectIETF()
	if err != nil && len(syslang) > 0 {
		return ""
	}
	Language = syslang
	return syslang
}

// 取得多語言文字
// `title` string 標題
func GetMultilingualText(title string) string {
	if value, ok := languageMap[title]; ok {
		return value
	}
	return title
}
