package nyanats

import (
	"log"
	"strings"
	"testing"
	"time"
)

func TestNyaNATS_MultiScenario(t *testing.T) {
	// --- 1. 構造生產級配置 ---
	conf := NATSConfig{
		NatsServer:    "127.0.0.1:4222",
		NatsUser:      "admin",
		NatsPassword:  "password123",
		EncryptionKey: "GLOBAL_BACKUP_KEY_32_CHARS_LONG!", // 全域性預設金鑰
		ThemeKeys: map[string]string{
			"theme.dedicated": "DEDICATED_SECRETKEY_32_CHAR_LONG", // 專用金鑰
			"theme.plain":     "",                                 // 顯式指定：明文傳輸
		},
	}

	// 初始化客戶端
	client := NewC(conf, log.Default())
	if err := client.Error(); err != nil {
		t.Fatalf("無法連線 NATS: %v", err)
	}
	defer client.Close()

	separator := strings.Repeat("=", 50)

	// --- 2. 場景 A：使用預設全域性金鑰 (主題：theme.default) ---
	t.Run("GlobalKey_Scenario", func(t *testing.T) {
		topic := "theme.default"
		t.Logf("\n%s\n[場景A] 預設金鑰測試 | 主題: %s\n%s", separator, topic, separator)

		// 訂閱端
		client.Subscribe(topic, func(m string) string {
			t.Logf("[訂閱收到] 解密後內容: %s", m)
			return "訂閱端解密成功！"
		})

		// 模式1：Publish (傳送)
		client.Publish(topic, "這是一條使用全域性金鑰加密的訊息")

		// 模式2：Request (傳送並獲取回覆)
		resp, _ := client.Request(topic, "詢問請求(全域性)", 1*time.Second)
		t.Logf("[請求結果] 收到加密回覆並解密: %s", resp)
	})

	// --- 3. 場景 B：使用專用金鑰 (主題：theme.dedicated) ---
	t.Run("DedicatedKey_Scenario", func(t *testing.T) {
		topic := "theme.dedicated"
		t.Logf("\n%s\n[場景B] 專用金鑰測試 | 主題: %s\n%s", separator, topic, separator)

		// 訂閱端
		client.Subscribe(topic, func(m string) string {
			t.Logf("[訂閱收到] 解密後內容: %s", m)
			return "訂閱端收到了專用金鑰加密的內容！"
		})

		// 模式1：Publish
		client.Publish(topic, "這是一條使用專用金鑰加密的訊息")

		// 模式2：Request
		resp, _ := client.Request(topic, "專用金鑰請求", 1*time.Second)
		t.Logf("[請求結果] 收到專用金鑰加密回覆: %s", resp)
	})

	// --- 4. 場景 C：明文傳輸 (主題：theme.plain) ---
	t.Run("Plaintext_Scenario", func(t *testing.T) {
		topic := "theme.plain"
		t.Logf("\n%s\n[場景C] 明文傳輸測試 | 主題: %s\n%s", separator, topic, separator)

		// 訂閱端
		client.Subscribe(topic, func(m string) string {
			t.Logf("[訂閱收到] 原始內容: %s", m)
			return "訂閱端收到了明文內容！"
		})

		// 模式1：Publish
		client.Publish(topic, "這是一條沒有加密的明文訊息")

		// 模式2：Request
		resp, _ := client.Request(topic, "明文詢問", 1*time.Second)
		t.Logf("[請求結果] 收到明文回覆: %s", resp)
	})

	// 為了讓非同步的 Publish 日誌完整輸出
	time.Sleep(500 * time.Millisecond)
}
