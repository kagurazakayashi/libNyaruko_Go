package nyanats

import (
	"log"
	"testing"
	"time"
)

func TestNyaNATS_FullConfig(t *testing.T) {
	// 1. 这是一个覆盖了所有功能的完整配置示例
	conf := NATSConfig{
		// 基础连接（对应你的 nats-server.conf）
		NatsServer:   "127.0.0.1:4222",
		NatsUser:     "admin",
		NatsPassword: "password123",

		// 运行控制
		ClientName:     "Tester-Device-01", // 如果为空，工具类会自动生成 UUID
		MaxReconnects:  10,                 // 允许重连 10 次
		ReconnectWait:  1,                  // 重连间隔 1 秒
		ConnectTimeout: 5,                  // 连接超时 5 秒

		// 加密设置
		// 全局默认密钥 (必须是 16, 24 或 32 字节)
		EncryptionKey: "global-default-key-32-char-long!",

		// 特定主题的差异化密钥映射
		ThemeKeys: map[string]string{
			"order.private": "special-order-key-32-char-long!!", // 订单主题用独立密钥
			"log.public":    "",                                 // 明确指定不加密
		},
	}

	// 2. 初始化客户端
	// 使用 log.Default() 可以在测试时看到 [NyaNATS] 的打出来的 debug 日志
	client := NewC(conf, log.Default())
	if err := client.Error(); err != nil {
		t.Fatalf("无法连接到 NATS Server: %v", err)
	}
	defer client.Close()

	// 3. 测试：特定主题加密通信 (order.private)
	testTopic := "order.private"
	testMsg := "Sensitive Order Content"
	done := make(chan string, 1)

	err := client.Subscribe(testTopic, func(m string) string {
		t.Logf("[Test] 收到并解密了消息: %s", m)
		done <- m
		return "ACK"
	})
	if err != nil {
		t.Fatalf("订阅失败: %v", err)
	}

	// 发送消息
	err = client.Publish(testTopic, testMsg)
	if err != nil {
		t.Errorf("发布失败: %v", err)
	}

	// 等待回调触发
	select {
	case received := <-done:
		if received != testMsg {
			t.Errorf("数据不一致！期望 %s, 得到 %s", testMsg, received)
		}
	case <-time.After(3 * time.Second):
		t.Error("测试超时：未收到订阅消息，请检查 NATS Server 是否运行以及密钥是否匹配")
	}
}
