package nyaapiserver

import (
	"time"
)

// LoggerFunc 定義日誌處理函式型別，用於接收單行日誌內容。
// 呼叫端可自行實作此函式，以整合檔案寫入、標準輸出或外部日誌系統。
type LoggerFunc func(line string)

// HttpAPIServerConfig 定義 HTTP API 伺服器的執行設定。
// 此結構集中管理網路監聽、TLS、逾時控制、連線行為、限流策略與日誌輸出等相關參數，
// 供伺服器初始化與啟動流程使用。
type HttpAPIServerConfig struct {
	// --- 基礎網路設定 ---

	// Host 為伺服器綁定的監聽位址，可為 IP 位址或網域名稱。
	// 例如："0.0.0.0" 代表監聽所有可用網路介面，"127.0.0.1" 則僅接受本機連線。
	Host string

	// Port 為伺服器監聽的連接埠號。
	// 例如：8080、8443。
	Port int

	// --- HTTPS 設定 ---

	// TLSCertFile 為 TLS 憑證檔案路徑，常見副檔名包含 .crt 或 .pem。
	// 當此欄位與 TLSKeyFile 皆為非空字串時，伺服器應以 HTTPS 模式啟動。
	TLSCertFile string

	// TLSKeyFile 為 TLS 私鑰檔案路徑。
	// 當此欄位與 TLSCertFile 皆為非空字串時，伺服器應以 HTTPS 模式啟動。
	TLSKeyFile string

	// --- 逾時與連線設定 ---

	// ReadTimeout 為讀取客戶端請求的逾時時間，
	// 通常涵蓋請求標頭與請求主體的接收階段。
	ReadTimeout time.Duration

	// WriteTimeout 為伺服器處理請求後，將回應寫回客戶端的逾時時間。
	WriteTimeout time.Duration

	// IdleTimeout 為 Keep-Alive 連線在閒置狀態下可維持的最長時間。
	// 超過此時間後，伺服器可主動關閉閒置連線以釋放資源。
	IdleTimeout time.Duration

	// --- 安全與限流設定（以單一 IP 為單位） ---

	// EnableRateLimit 表示是否啟用以來源 IP 為基礎的請求頻率限制機制。
	EnableRateLimit bool

	// LimitRequests 為單一 IP 在指定時間視窗內可接受的最大請求數。
	LimitRequests int

	// LimitWindow 為限流統計所使用的時間視窗長度。
	LimitWindow time.Duration

	// BlockDuration 為來源 IP 超出限流門檻後的封鎖持續時間。
	BlockDuration time.Duration

	// --- 日誌設定 ---

	// Logger 為可選的日誌輸出處理函式。
	// 若為 nil，表示不主動輸出日誌，由呼叫端自行決定是否處理。
	Logger LoggerFunc
}

// DefaultConfig 回傳一組可直接使用的預設伺服器設定。
// 此預設值適合一般開發或基礎部署情境，呼叫端可依實際需求覆寫各欄位。
func DefaultConfig() *HttpAPIServerConfig {
	return &HttpAPIServerConfig{
		Host:            "0.0.0.0",
		Port:            8080,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
		EnableRateLimit: true,
		LimitRequests:   50,               // 預設值：單一 IP 於單一時間視窗內最多允許 50 次請求。
		LimitWindow:     1 * time.Second,  // 預設值：限流統計時間視窗為 1 秒。
		BlockDuration:   10 * time.Minute, // 預設值：觸發限流後，封鎖該 IP 10 分鐘。
		Logger:          nil,              // 預設值：不輸出日誌，由呼叫端自行注入處理函式。
	}
}
