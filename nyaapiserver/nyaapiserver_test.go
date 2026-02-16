package nyaapiserver

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestRunServer(t *testing.T) {
	// 1. 初始化配置
	conf := DefaultConfig()
	conf.Port = 9000
	conf.LimitRequests = 5 // 為了方便測試限流，我們將閾值調低

	// 2. 定義處理邏輯
	handler := func(req *HTTPRequest) *HTTPResponse {
		// 列印接收到的關鍵資訊
		fmt.Printf("\n[Request] %s %s | From: %s\n", req.Method, req.Path, req.RemoteAddr)
		if len(req.Params) > 0 {
			fmt.Printf("  Params: %v\n", req.Params)
		}
		if len(req.Cookies) > 0 {
			fmt.Printf("  Cookies: %v\n", req.Cookies)
		}

		// 路由邏輯
		switch req.Path {
		case "/stats":
			s := GetStats()
			msg := fmt.Sprintf("Stats: Requests=%d, Conns=%d, Sent=%d bytes, BlockedIPs=%v",
				s.TotalRequests, s.CurrentConns, s.TotalBytesSent, s.BlockedIPs)
			return &HTTPResponse{StatusCode: 200, Body: []byte(msg)}

		case "/hello":
			return &HTTPResponse{
				StatusCode: 200,
				Body:       []byte("Hello, World!"),
				Headers:    map[string]string{"X-Custom-Header": "Go-Server"},
			}

		default:
			return &HTTPResponse{StatusCode: 404, Body: []byte("Not Found")}
		}
	}

	// 3. 建立並啟動伺服器
	srv := NewServer(conf, handler)

	// 在協程中啟動，避免阻塞主執行緒
	go func() {
		if err := srv.Start(); err != nil {
			fmt.Printf("Server Error: %v\n", err)
		}
	}()

	// 4. 等待訊號 (Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 5. 優雅關閉 (給定 5 秒超時)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		t.Fatalf("Server forced to shutdown: %v", err)
	}
}
