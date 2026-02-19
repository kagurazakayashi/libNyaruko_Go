package nyaapiserver

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"testing"
	"time"
)

// TestRunServer 用於驗證伺服器在實際執行流程中的核心行為是否符合預期，
// 包含啟動、請求處理、統計資訊觀察，以及接收系統訊號後的優雅關閉機制。
//
// 測試流程說明：
// 1. 建立預設伺服器設定，並覆寫測試所需的連接埠與請求限制。
// 2. 註冊測試專用日誌函式，統一輸出 HTTP 日誌與目前伺服器統計資訊。
// 3. 啟動 HTTP 伺服器，供外部以手動方式發送請求驗證行為。
// 4. 阻塞等待 SIGINT 或 SIGTERM，模擬實際服務停止情境。
// 5. 接收到終止訊號後，以具逾時控制的 Context 執行優雅關閉。
//
// 注意事項：
// - 此測試屬於偏向手動操作的整合測試，不適合在全自動測試流程中直接執行。
// - 測試會持續阻塞，直到程序接收到中斷或終止訊號。
// - 若需驗證路由、統計資訊或連線關閉行為，建議搭配 curl 或瀏覽器手動測試。
func TestRunServer(t *testing.T) {
	// 初始化伺服器設定，並覆寫此次測試所需的關鍵參數。
	conf := DefaultConfig()
	conf.Port = 9000
	conf.LimitRequests = 5

	// 設定自訂日誌函式：
	// - 統一接收伺服器內部輸出的 HTTP 日誌。
	// - 當日誌內容不是內部控制訊息時，一併輸出目前統計資訊，方便觀察執行狀態。
	myLogger := func(line string) {
		if len(line) == 0 {
			return
		}

		fmt.Printf("[TestRunServer] [HTTPLOG] %s\n", line)

		if line[0] != '#' {
			// 取得目前伺服器狀態，便於同步觀察請求數、連線數與流量變化。
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

	// 定義測試用請求處理函式，依據不同路徑回傳對應結果，
	// 並輸出請求相關資訊，方便在手動測試時追蹤行為。
	handler := func(req *HTTPRequest) *HTTPResponse {
		// 輸出本次請求的核心資訊，包含 HTTP 方法、路徑與來源位址。
		fmt.Printf("[TestRunServer] [REQUEST] %s %s | From: %s\n", req.Method, req.Path, req.RemoteAddr)

		// 若請求含有標頭資料，則輸出完整內容供除錯與比對使用。
		if len(req.Headers) > 0 {
			fmt.Printf("[TestRunServer] [HEADERS] %v\n", req.Headers)
		}

		// 若請求含有查詢參數或表單參數，則輸出參數內容。
		if len(req.Params) > 0 {
			fmt.Printf("[TestRunServer] [PARAMS] %v\n", req.Params)
		}

		// 若請求攜帶 Cookie，則輸出 Cookie 內容，便於檢查會話相關資訊。
		if len(req.Cookies) > 0 {
			fmt.Printf("[TestRunServer] [COOKIES] %v\n", req.Cookies)
		}

		// 根據請求路徑執行對應路由邏輯。
		switch req.Path {
		case "/ping":
			// /ping 用於測試基本存活狀態，若帶入 t 參數則計算用戶端到伺服器的時間差。
			var latency int64 = 0
			if tStr, ok := req.Params["t"]; ok {
				if clientTime, err := strconv.ParseInt(tStr, 10, 64); err == nil {
					latency = time.Now().UnixMilli() - clientTime
				}
			}

			return JSONResponse(200, map[string]int64{
				"pong": latency,
			})

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
			// 回傳固定測試內容，並附加自訂標頭，方便驗證回應內容與標頭是否正確。
			return &HTTPResponse{
				StatusCode: 200,
				Body:       []byte("Hello, World!"),
				Headers:    map[string]string{"X-Custom-Header": "Go-Server"},
			}

		default:
			// 未命中任何已知路由時，回傳 404 Not Found。
			return &HTTPResponse{StatusCode: 404, Body: []byte("Not Found")}
		}
	}

	// 建立伺服器實例，並綁定前述設定與請求處理函式。
	srv := NewServer(conf, handler, myLogger)

	// 於獨立 goroutine 中啟動伺服器，避免阻塞目前測試主流程。
	go func() {
		if err := srv.Start("libNyaruko_Go TestRunServer", "1.0"); err != nil {
			fmt.Printf("[TestRunServer] [ERROR] %v\n", err)
		}
	}()

	// 建立系統訊號通道，等待外部送入 SIGINT 或 SIGTERM，
	// 以模擬實際部署環境中的手動停止或程序終止情境。
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 建立具逾時控制的關閉 Context，避免關閉流程因長時間等待而無限阻塞。
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 執行優雅關閉，讓既有連線與請求有機會在期限內正常完成。
	if err := srv.Stop(ctx); err != nil {
		t.Fatalf("Server forced to shutdown: %v", err)
	}
}
