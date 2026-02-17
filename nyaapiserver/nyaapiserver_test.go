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

// TestRunServer 用於驗證伺服器的啟動、請求處理與優雅關閉流程是否正常。
// 此測試會：
// 1. 建立預設伺服器設定並覆寫必要參數。
// 2. 註冊自訂日誌輸出函式，以便觀察請求與伺服器統計資訊。
// 3. 啟動測試用 HTTP 伺服器。
// 4. 等待系統中斷訊號後執行優雅關閉。
// 注意：此測試會阻塞至接收到 SIGINT 或 SIGTERM，較適合手動整合測試情境。
func TestRunServer(t *testing.T) {
	// 初始化伺服器設定，並調整測試所需參數。
	conf := DefaultConfig()
	conf.Port = 9000
	conf.LimitRequests = 5

	// 由呼叫端統一處理 HTTP 日誌輸出，並同步列印目前伺服器統計資訊。
	conf.Logger = func(line string) {
		if len(line) == 0 {
			return
		}
		fmt.Printf("[TestRunServer] [HTTPLOG] %s\n", line)
		if line[0] != '#' {
			// 取得目前伺服器狀態，便於觀察連線與流量變化。
			stats := GetStats()
			fmt.Printf(
				"[TestRunServer] [STATUS] Requests=%d, Conn=%d, Sent=%d, Recv=%d, Uptime=%s, BlockedIP=%v\n",
				stats.TotalRequests,
				stats.CurrentConns,
				stats.TotalBytesSent,
				stats.TotalBytesRecv,
				stats.Uptime,
				stats.BlockedIPs,
			)
		}
	}

	// 定義測試用請求處理函式，依據不同路徑回傳對應內容。
	handler := func(req *HTTPRequest) *HTTPResponse {
		// 列印請求的核心資訊，協助追蹤請求來源與路由行為。
		fmt.Printf("\n[TestRunServer] [REQUEST] %s %s | From: %s\n", req.Method, req.Path, req.RemoteAddr)

		// 當請求包含查詢參數時，輸出其內容以利除錯。
		if len(req.Params) > 0 {
			fmt.Printf("[TestRunServer] [PARAMS] %v\n", req.Params)
		}

		// 當請求包含 Cookie 時，輸出其內容以利驗證會話或狀態資訊。
		if len(req.Cookies) > 0 {
			fmt.Printf("[TestRunServer] [COOKIES] %v\n", req.Cookies)
		}

		// 根據請求路徑執行對應路由邏輯。
		switch req.Path {
		case "/stats":
			// 回傳目前伺服器統計資訊，供外部快速檢視執行狀態。
			s := GetStats()
			msg := fmt.Sprintf(
				"[TestRunServer] Stats: Requests=%d, Conns=%d, Sent=%d bytes, BlockedIPs=%v",
				s.TotalRequests,
				s.CurrentConns,
				s.TotalBytesSent,
				s.BlockedIPs,
			)
			return &HTTPResponse{StatusCode: 200, Body: []byte(msg)}

		case "/hello":
			// 回傳固定測試內容，並附加自訂標頭供驗證。
			return &HTTPResponse{
				StatusCode: 200,
				Body:       []byte("Hello, World!"),
				Headers:    map[string]string{"X-Custom-Header": "Go-Server"},
			}

		default:
			// 未命中路由時回傳 404。
			return &HTTPResponse{StatusCode: 404, Body: []byte("Not Found")}
		}
	}

	// 建立伺服器實例，並綁定前述設定與請求處理函式。
	srv := NewServer(conf, handler)

	// 於獨立 goroutine 中啟動伺服器，避免阻塞目前測試流程。
	go func() {
		if err := srv.Start("libNyaruko_Go TestRunServer", "1.0"); err != nil {
			fmt.Printf("[TestRunServer] [ERROR] %v\n", err)
		}
	}()

	// 等待系統中斷訊號，以模擬手動停止服務的場景。
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 建立具逾時控制的關閉 Context，避免關閉流程無限等待。
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 執行優雅關閉，確保既有連線有機會正常結束。
	if err := srv.Stop(ctx); err != nil {
		t.Fatalf("Server forced to shutdown: %v", err)
	}
}
