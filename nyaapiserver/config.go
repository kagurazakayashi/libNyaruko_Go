package nyaapiserver

import (
	"time"
)

// HttpAPIServerConfig 定義了 HTTP API 伺服器的詳細配置項
type HttpAPIServerConfig struct {
	// --- 基礎網路配置 ---
	Host string // 監聽的 IP 或域名 (例如 "0.0.0.0" 或 "127.0.0.1")
	Port int    // 監聽的埠 (例如 8080)

	// --- HTTPS 配置 ---
	// 如果這兩個欄位不為空，伺服器將自動以 HTTPS 模式啟動
	TLSCertFile string // TLS 證書檔案路徑 (.crt / .pem)
	TLSKeyFile  string // TLS 私鑰檔案路徑 (.key)

	// --- 超時與連線配置 ---
	ReadTimeout  time.Duration // 讀取客戶端請求（包括請求頭和請求體）的超時時間
	WriteTimeout time.Duration // 伺服器處理並向客戶端寫入響應的超時時間
	IdleTimeout  time.Duration // 開啟 Keep-Alive 時，空閒連線的超時時間

	// --- 安全與限流配置 (針對單個 IP) ---
	EnableRateLimit bool          // 是否開啟 IP 限流防護
	LimitRequests   int           // 在指定時間視窗內允許的最大請求次數
	LimitWindow     time.Duration // 限流的時間視窗長度
	BlockDuration   time.Duration // 超過限流閾值後，將該 IP 封鎖的時長
}

// DefaultConfig 提供一套開箱即用的預設配置
func DefaultConfig() *HttpAPIServerConfig {
	return &HttpAPIServerConfig{
		Host:            "0.0.0.0",
		Port:            8080,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
		EnableRateLimit: true,
		LimitRequests:   50,               // 預設配置：單 IP 允許 50次請求
		LimitWindow:     1 * time.Second,  // 預設配置：每 1 秒
		BlockDuration:   10 * time.Minute, // 預設配置：觸發頻率限制後，臨時拉黑該 IP 10 分鐘
	}
}
