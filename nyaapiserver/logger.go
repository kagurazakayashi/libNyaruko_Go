package nyaapiserver

import (
	"fmt"
	"time"
)

// GetW3CHeader 產生符合 W3C 擴充日誌格式（W3C Extended Log File Format）的標頭內容。
//
// 此函式會回傳可直接寫入日誌檔案開頭的多行標頭，內容包含：
//   - 軟體名稱（#Software）
//   - 軟體版本（#Version）
//   - 以 UTC 表示的產生日誌時間（#Date）
//   - 日誌欄位定義（#Fields）
//
// 參數：
//   - serverName：寫入 #Software 的伺服器名稱。
//   - serverVersion：寫入 #Version 的伺服器版本。
//
// 回傳值：
//   - []string：依 W3C 日誌慣例排列的標頭字串切片，可逐行輸出至日誌檔案。
func GetW3CHeader(serverName string, serverVersion string) []string {
	return []string{
		fmt.Sprintf("#Software: %s", serverName),
		fmt.Sprintf("#Version: %s", serverVersion),
		fmt.Sprintf("#Date: %s", time.Now().UTC().Format("2006-01-02 15:04:05")),
		"#Fields: date time c-ip cs-method cs-uri-stem cs-uri-query s-port cs-username c-ip cs(User-Agent) sc-status time-taken",
	}
}

// FormatW3CLine 將單筆 HTTP 請求資訊格式化為單行 W3C 日誌內容。
//
// 此函式會以目前 UTC 時間作為日誌時間，並依既定欄位順序組合輸出字串。
// 針對可選欄位，若 userAgent 或 query 為空字串，會以 "-" 作為預設佔位值，
// 以維持日誌欄位數量與格式的一致性。
//
// 參數：
//   - ip：用戶端 IP 位址。
//   - method：HTTP 請求方法，例如 GET、POST。
//   - path：請求路徑。
//   - query：查詢字串；若為空則輸出 "-"。
//   - port：服務埠號。
//   - userAgent：用戶端 User-Agent；若為空則輸出 "-"。
//   - status：HTTP 回應狀態碼。
//   - duration：請求處理耗時，最終會以毫秒表示。
//
// 回傳值：
//   - string：可直接寫入 W3C 日誌檔案的單行字串。
func FormatW3CLine(ip string, method string, path string, query string, port int, userAgent string, status int, duration time.Duration) string {
	now := time.Now().UTC()
	date := now.Format("2006-01-02")
	timeStr := now.Format("15:04:05")

	// 若未提供 User-Agent，使用 W3C 日誌常見的空值佔位符號。
	if userAgent == "" {
		userAgent = "-"
	}

	// 若查詢字串為空，使用 "-" 以避免欄位缺漏。
	if query == "" {
		query = "-"
	}

	// time-taken 以毫秒輸出，便於後續搜尋、統計與效能分析。
	return fmt.Sprintf("%s %s %s %s %s %s %d - %s %s %d %d",
		date, timeStr, ip, method, path, query, port, ip, userAgent, status, duration.Milliseconds())
}
