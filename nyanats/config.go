// 本檔案負責處理 NATS 連線的配置設定與預設值填充。
package nyanats

import "github.com/google/uuid"

// NatsConfig 定義了連線至 NATS 伺服器所需的各項參數。
// 支援透過 JSON 或 YAML 標籤進行序列化與反序列化。
type NatsConfig struct {
	// NatsServerHost 是 NATS 伺服器的主機名稱或 IP 位址（例如 "127.0.0.1"）。
	NatsServerHost string `json:"nats_server_host" yaml:"nats_server_host"`

	// NatsServerPort 是 NATS 伺服器的埠號（例如 4222）。
	NatsServerPort int `json:"nats_server_port" yaml:"nats_server_port"`

	// NatsUser 是用於連線驗證的使用者名稱。
	NatsUser string `json:"nats_user" yaml:"nats_user"`

	// NatsPassword 是用於連線驗證的密碼。
	NatsPassword string `json:"nats_password" yaml:"nats_password"`

	// NatsClient 是此用戶端在 NATS 伺服器上顯示的識別名稱。
	NatsClient string `json:"client_name" yaml:"client_name"`

	// NatsMaxReconnects 定義了連線中斷後的最大嘗試重新連線次數。
	NatsMaxReconnects int `json:"max_reconnects" yaml:"max_reconnects"`

	// NatsReconnectWait 定義了每次嘗試重新連線之間的等待秒數。
	NatsReconnectWait int `json:"reconnect_wait" yaml:"reconnect_wait"`

	// NatsConnectTimeout 定義了建立初始連線的逾時秒數。
	NatsConnectTimeout int `json:"connect_timeout" yaml:"connect_timeout"`

	// NatsEncryptionKey 是全域預設的對稱加密金鑰。
	// - 注意：長度必須嚴格遵守 16, 24 或 32 字節 (Bytes)，分別對應 AES-128, AES-192, AES-256。
	// - 產生一個隨機的 32 字元 Base64 字串（適用於 256-bit 加密）:
	NatsEncryptionKey string `json:"encryption_key" yaml:"encryption_key"`

	// NatsThemeKeys 可針對特定的主題（Subject）設定獨立的加密金鑰。
	// - 優先權：若主題匹配，優先使用此處的金鑰；若無匹配，則回退至全域金鑰。
	// - 明文傳輸：若值設為空字串 ""，則該主題將以「明文」傳輸，不進行任何加密。
	// - 格式要求：同全域金鑰，長度必須為 16, 24 或 32 字節。
	NatsThemeKeys map[string]string `json:"theme_keys" yaml:"theme_keys"`
}

// setDefaults 檢查配置項，若欄位為空值或零值，則填充預設參數。
func (c *NatsConfig) setDefaults() {
	// 若未指定伺服器主機，預設為本機位址
	if c.NatsServerHost == "" {
		c.NatsServerHost = "127.0.0.1"
	}

	// 若未指定伺服器埠號，預設為 4222
	if c.NatsServerPort == 0 {
		c.NatsServerPort = 4222
	}

	// 若未指定用戶端名稱，則自動生成一個隨機的 UUID 避免名稱衝突
	if c.NatsClient == "" {
		c.NatsClient = uuid.NewString()
	}

	// 預設最大重新連線嘗試次數為 5 次
	if c.NatsMaxReconnects == 0 {
		c.NatsMaxReconnects = 5
	}

	// 預設重新連線等待時間為 2 秒
	if c.NatsReconnectWait == 0 {
		c.NatsReconnectWait = 2
	}

	// 預設連線逾時時間為 10 秒
	if c.NatsConnectTimeout == 0 {
		c.NatsConnectTimeout = 10
	}
}
