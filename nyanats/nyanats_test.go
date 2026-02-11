// Package nyanats 提供 NATS 客戶端的測試案例，包含加密與明文傳輸場景。
// DEBUG=NYANATS
package nyanats

import (
	"crypto/rand"
	"log"
	"strings"
	"testing"
	"time"
)

// TestNyaNATS_MultiScenario 測試 NyaNATS 在不同加密配置下的多種場景，
// 包含全域金鑰加密、專用金鑰加密以及明文傳輸。
func TestNyaNATS_MultiScenario(t *testing.T) {
	// --- 0. 產生並輸出一個範例隨機金鑰 (不作後續用途) ---
	// "產生的範例隨機金鑰 (32字元): %s"
	exampleKey := GenerateRandomKey(32)
	t.Logf("Generated example random key (32 chars): %s", exampleKey)

	// --- 1. 構造生產級配置 ---
	// 16 位金鑰：對應 AES-128
	// 24 位金鑰：對應 AES-192
	// 32 位金鑰：對應 AES-256
	conf := NatsConfig{
		NatsServer:        "127.0.0.1:4222",
		NatsUser:          "admin",
		NatsClient:        "GO Test Client",
		NatsPassword:      "password123",
		NatsEncryptionKey: "GLOBAL_BACKUP_KEY_32_CHARS_LONG!", // 全域性預設金鑰
		NatsThemeKeys: map[string]string{
			"theme.dedicated": "DEDICATED_SECRETKEY_32_CHAR_LONG", // 專用金鑰
			"theme.plain":     "",                                 // 顯式指定：明文傳輸
		},
	}

	// 初始化客戶端並檢查連線狀態
	client := NewC(conf, log.Default())
	if err := client.Error(); err != nil {
		// "無法連線 NATS: %v"
		t.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer client.Close()

	// 產生用於日誌分隔的橫線
	separator := strings.Repeat("=", 50)

	// --- 2. 場景 A：使用預設全域金鑰 (主題：theme.default) ---
	t.Run("GlobalKey_Scenario", func(t *testing.T) {
		topic := "theme.default"
		// "\n%s\n[場景A] 預設金鑰測試 | 主題: %s\n%s"
		t.Logf("\n%s\n[Scenario A] Default Key Test | Topic: %s\n%s", separator, topic, separator)

		// 訂閱端處理邏輯
		client.Subscribe(topic, func(m string) string {
			// "[訂閱收到] 解密後內容: %s"
			t.Logf("[Sub Received] Decrypted content: %s", m)
			// "訂閱端解密成功！"
			return "Subscriber decryption successful!"
		})

		// 模式 1：Publish (傳送加密訊息)
		// "這是一條使用全域性金鑰加密的訊息"
		client.Publish(topic, "This is a message encrypted using the global key")

		// 模式 2：Request (傳送請求並獲取回覆)
		// "詢問請求(全域性)"
		resp, _ := client.Request(topic, "Inquiry Request (Global)", 1*time.Second)
		// "[請求結果] 收到加密回覆並解密: %s"
		t.Logf("[Req Result] Received encrypted reply and decrypted: %s", resp)
	})

	// --- 3. 場景 B：使用專用金鑰 (主題：theme.dedicated) ---
	t.Run("DedicatedKey_Scenario", func(t *testing.T) {
		topic := "theme.dedicated"
		// "\n%s\n[場景B] 專用金鑰測試 | 主題: %s\n%s"
		t.Logf("\n%s\n[Scenario B] Dedicated Key Test | Topic: %s\n%s", separator, topic, separator)

		// 訂閱端處理邏輯
		client.Subscribe(topic, func(m string) string {
			// "[訂閱收到] 解密後內容: %s"
			t.Logf("[Sub Received] Decrypted content: %s", m)
			// "訂閱端收到了專用金鑰加密的內容！"
			return "Subscriber received content encrypted with dedicated key!"
		})

		// 模式 1：Publish
		// "這是一條使用專用金鑰加密的訊息"
		client.Publish(topic, "This is a message encrypted using a dedicated key")

		// 模式 2：Request
		// "專用金鑰請求"
		resp, _ := client.Request(topic, "Dedicated Key Request", 1*time.Second)
		// "[請求結果] 收到專用金鑰加密回覆: %s"
		t.Logf("[Req Result] Received reply encrypted with dedicated key: %s", resp)
	})

	// --- 4. 場景 C：明文傳輸 (主題：theme.plain) ---
	t.Run("Plaintext_Scenario", func(t *testing.T) {
		topic := "theme.plain"
		// "\n%s\n[場景C] 明文傳輸測試 | 主題: %s\n%s"
		t.Logf("\n%s\n[Scenario C] Plaintext Transmission Test | Topic: %s\n%s", separator, topic, separator)

		// 訂閱端處理邏輯
		client.Subscribe(topic, func(m string) string {
			// "[訂閱收到] 原始內容: %s"
			t.Logf("[Sub Received] Plaintext content: %s", m)
			// "訂閱端收到了明文內容！"
			return "Subscriber received plaintext content!"
		})

		// 模式 1：Publish
		// "這是一條沒有加密的明文訊息"
		client.Publish(topic, "This is a plaintext message without encryption")

		// 模式 2：Request
		// "明文詢問"
		resp, _ := client.Request(topic, "Plaintext Inquiry", 1*time.Second)
		// "[請求結果] 收到明文回覆: %s"
		t.Logf("[Req Result] Received plaintext reply: %s", resp)
	})

	// 暫停一小段時間以確保非同步的 Publish 日誌能完整輸出到控制台
	time.Sleep(500 * time.Millisecond)
}

// GenerateRandomKey 產生指定長度的隨機字串，僅包含大小寫字母與數字。
func GenerateRandomKey(n int) string {
	// 使用 strings.Builder 是 Go 處理字串拼接的最佳實踐，
	// 它能有效減少記憶體分配 (Memory Allocation) 的次數，提升效能。
	var b strings.Builder
	// 1. 加入數字 '0' 到 '9' (ASCII 48-57)
	for i := '0'; i <= '9'; i++ {
		b.WriteRune(i)
	}
	// 2. 加入大寫字母 'A' 到 'Z' (ASCII 65-90)
	for i := 'A'; i <= 'Z'; i++ {
		b.WriteRune(i)
	}
	// 3. 加入小寫字母 'a' 到 'z' (ASCII 97-122)
	for i := 'a'; i <= 'z'; i++ {
		b.WriteRune(i)
	}
	// 定義允許的字元集：確保每個字元都是單一位元組 (Single-byte)，
	// 這樣產生的字串長度將會精確等於其位元組長度，符合 AES 的嚴格要求。
	charset := b.String()

	// 初始化一個長度為 n 的位元組切片 (Byte Slice)。
	c := make([]byte, n)

	// 使用 crypto/rand 而非 math/rand。
	// crypto/rand 提供的是「加密安全擬隨機數產生器」(CSPRNG)，
	// 它是基於作業系統提供的熵來源（如 Linux 的 /dev/urandom），
	// 其產生的隨機數無法被預測，適合用於安全敏感的金鑰產生。
	if _, err := rand.Read(c); err != nil {
		// 若發生系統級錯誤（例如熵池耗盡），則回傳空字串。
		// "產生隨機數失敗"
		return ""
	}

	// 遍歷剛產生的隨機位元組切片。
	for i := range c {
		// 將隨機的位元組數值對字元集長度取餘數 (Modulo)。
		// 這樣可以將任何 0-255 的數值映射回我們定義的 charset 索引範圍內 (0-61)。
		// 雖然這在極小機率下會有些微的偏態 (Bias)，但對於產生一般金鑰已足夠安全。
		c[i] = charset[int(c[i])%len(charset)]
	}

	// 將處理後的位元組切片轉換成最終的字串格式回傳。
	return string(c)
}
