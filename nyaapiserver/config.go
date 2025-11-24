package nyaapiserver

import "time"

// LoggerFunc 定義單行日誌輸出的處理函式型別。
// 呼叫端可透過實作此函式，將日誌整合至標準輸出、檔案、集中式日誌平台或其他外部觀測系統。
// 傳入參數 line 應視為已完成格式化的單行訊息，實作者通常不應再對其進行結構拆解。
type LoggerFunc func(line string)

// HttpAPIServerConfig 定義 HTTP API 伺服器的執行期設定。
// 此結構集中管理伺服器啟動與運作所需的核心參數，包含監聽位址、TLS、連線逾時、限流控制與日誌輸出策略。
// 呼叫端可先使用 DefaultConfig 取得建議預設值，再依部署環境、流量模型與安全需求進行覆寫。
type HttpAPIServerConfig struct {
	// --- 基礎網路設定 ---

	// Host 為伺服器綁定的監聽位址，可為 IP 位址或網域名稱。
	// 例如："0.0.0.0" 代表監聽所有可用網路介面，"127.0.0.1" 則僅接受本機連線。
	Host string `json:"httpapiserver_host" yaml:"httpapiserver_host"`

	// Port 為伺服器監聽的連接埠號。
	// 實際部署時，應確認該連接埠未被其他服務占用，並與防火牆或反向代理設定一致。
	Port int `json:"httpapiserver_port" yaml:"httpapiserver_port"`

	// --- HTTPS 設定 ---

	// TLSCertFile 為 TLS 憑證檔案路徑，常見副檔名包含 .crt 或 .pem。
	// 當 TLSCertFile 與 TLSKeyFile 皆為非空字串時，伺服器應以 HTTPS 模式啟動。
	// 若任一欄位為空，通常表示沿用純 HTTP，或由上游代理終止 TLS。
	TLSCertFile string `json:"httpapiserver_tls_cert_file" yaml:"httpapiserver_tls_cert_file"`

	// TLSKeyFile 為 TLS 私鑰檔案路徑。
	// 此欄位應與 TLSCertFile 配對使用，且檔案權限應妥善控管，以避免私鑰外洩風險。
	TLSKeyFile string `json:"httpapiserver_tls_key_file" yaml:"httpapiserver_tls_key_file"`

	// --- 逾時與連線設定 ---

	// ReadTimeout 為讀取客戶端請求的逾時時間。
	// 一般包含請求標頭與請求主體的接收階段，可用於降低慢速連線或惡意請求長時間占用資源的風險。
	ReadTimeout int `json:"httpapiserver_read_timeout" yaml:"httpapiserver_read_timeout"`

	// WriteTimeout 為伺服器完成請求處理後，將回應寫回客戶端的逾時時間。
	// 合理設定此值可避免回應傳輸階段因對端過慢而長時間占用連線資源。
	WriteTimeout int `json:"httpapiserver_write_timeout" yaml:"httpapiserver_write_timeout"`

	// IdleTimeout 為 Keep-Alive 連線在閒置狀態下可維持的最長時間。
	// 超過此時間後，伺服器可主動關閉閒置連線，以回收檔案描述符與連線相關資源。
	IdleTimeout int `json:"httpapiserver_idle_timeout" yaml:"httpapiserver_idle_timeout"`

	// --- 安全與限流設定（以單一 IP 為單位） ---

	// EnableRateLimit 表示是否啟用以來源 IP 為基礎的請求頻率限制機制。
	// 啟用後，可降低暴力請求、惡意探測或短時間高頻流量對服務穩定性的影響。
	EnableRateLimit bool `json:"httpapiserver_enable_rate_limit" yaml:"httpapiserver_enable_rate_limit"`

	// LimitRequests 為單一 IP 在指定時間視窗內可接受的最大請求數。
	// 當請求數超過此門檻時，系統可視為觸發限流處置。
	LimitRequests int `json:"httpapiserver_limit_requests" yaml:"httpapiserver_limit_requests"`

	// LimitWindow 為限流統計所使用的時間視窗長度。
	// 此欄位會與 LimitRequests 搭配使用，以定義單位時間內的最大可接受請求量。
	LimitWindow int `json:"httpapiserver_limit_window" yaml:"httpapiserver_limit_window"`

	// BlockDuration 為來源 IP 超出限流門檻後的封鎖持續時間。
	// 適當的封鎖時間有助於抑制重複攻擊，但也應避免因設定過長而影響正常使用者恢復存取。
	BlockDuration int `json:"httpapiserver_block_duration" yaml:"httpapiserver_block_duration"`
}

// DefaultConfig 回傳一組可直接使用的預設伺服器設定。
// 此預設值適用於一般開發環境或基礎部署情境，提供相對保守且實務上常見的初始參數。
// 在正式環境中，仍建議依實際網路拓樸、流量型態、安全要求與 SLA 目標進一步調整。
func DefaultConfig() *HttpAPIServerConfig {
	return &HttpAPIServerConfig{
		Host:            "0.0.0.0",
		Port:            8080,
		ReadTimeout:     10,
		WriteTimeout:    15,
		IdleTimeout:     120,
		EnableRateLimit: true,
		LimitRequests:   30,  // 預設值：單一 IP 於單一時間視窗內最多允許 30 次請求。
		LimitWindow:     1,   // 預設值：限流統計時間視窗為 1 秒。
		BlockDuration:   300, // 預設值：觸發限流後，封鎖該 IP 5 分鐘。
	}
}

func Second(second int) time.Duration {
	return time.Duration(second) * time.Second
}
