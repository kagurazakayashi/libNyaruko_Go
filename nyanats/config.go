// 本檔案負責處理 NATS 連線的配置設定與預設值填充。
package nyanats

import "github.com/google/uuid"

// NATSConfig 定義了連線至 NATS 伺服器所需的各項參數。
// 支援透過 JSON 或 YAML 標籤進行序列化與反序列化。
type NATSConfig struct {
	// NatsServer 是 NATS 伺服器的位址（例如 "127.0.0.1:4222"）。
	NatsServer string `json:"nats_server" yaml:"nats_server"`

	// NatsUser 是用於連線驗證的使用者名稱。
	NatsUser string `json:"nats_user" yaml:"nats_user"`

	// NatsPassword 是用於連線驗證的密碼。
	NatsPassword string `json:"nats_password" yaml:"nats_password"`

	// ClientName 是此用戶端在 NATS 伺服器上顯示的識別名稱。
	ClientName string `json:"client_name" yaml:"client_name"`

	// MaxReconnects 定義了連線中斷後的最大嘗試重新連線次數。
	MaxReconnects int `json:"max_reconnects" yaml:"max_reconnects"`

	// ReconnectWait 定義了每次嘗試重新連線之間的等待秒數。
	ReconnectWait int `json:"reconnect_wait" yaml:"reconnect_wait"`

	// ConnectTimeout 定義了建立初始連線的逾時秒數。
	ConnectTimeout int `json:"connect_timeout" yaml:"connect_timeout"`

	// EncryptionKey 是全域預設的對稱加密金鑰。
	// - 注意：長度必須嚴格遵守 16, 24 或 32 字節 (Bytes)，分別對應 AES-128, AES-192, AES-256。
	// - 產生一個隨機的 32 字元 Base64 字串（適用於 256-bit 加密）:
	EncryptionKey string `json:"encryption_key" yaml:"encryption_key"`

	// ThemeKeys 可針對特定的主題（Subject）設定獨立的加密金鑰。
	// - 優先權：若主題匹配，優先使用此處的金鑰；若無匹配，則回退至全域金鑰。
	// - 明文傳輸：若值設為空字串 ""，則該主題將以「明文」傳輸，不進行任何加密。
	// - 格式要求：同全域金鑰，長度必須為 16, 24 或 32 字節。
	ThemeKeys map[string]string `json:"theme_keys" yaml:"theme_keys"`
}

// setDefaults 檢查配置項，若欄位為空值或零值，則填充預設參數。
func (c *NATSConfig) setDefaults() {
	// 若未指定伺服器位址，預設連線至本地端 4222 埠口
	if c.NatsServer == "" {
		c.NatsServer = "127.0.0.1:4222"
	}

	// 若未指定用戶端名稱，則自動生成一個隨機的 UUID 避免名稱衝突
	if c.ClientName == "" {
		c.ClientName = uuid.NewString()
	}

	// 預設最大重新連線嘗試次數為 5 次
	if c.MaxReconnects == 0 {
		c.MaxReconnects = 5
	}

	// 預設重新連線等待時間為 2 秒
	if c.ReconnectWait == 0 {
		c.ReconnectWait = 2
	}

	// 預設連線逾時時間為 10 秒
	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = 10
	}
}
