// HTTP API 服务器
package nyaapiserver

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// Server 為 HTTP API 伺服器的核心實作，負責：
// 1. 初始化底層 http.Server 與相關設定。
// 2. 接收並分派所有進入的 HTTP 請求。
// 3. 整合限流、流量統計、連線統計與 W3C 日誌輸出。
// 4. 提供伺服器啟動、優雅關閉與狀態查詢能力。
type Server struct {
	httpServer *http.Server
	config     *HttpAPIServerConfig
	handler    HandlerFunc
	logger     LoggerFunc
}

// NewServer 建立並初始化一個 Server 實例。
//
// 此函式會：
// 1. 依照傳入設定初始化全域限流器。
// 2. 建立 Server 結構。
// 3. 建立底層 http.Server，並將目前 Server 自身註冊為請求處理入口。
//
// 參數：
//   - conf: HTTP API 伺服器設定。
//   - handler: 實際處理業務邏輯的回呼函式。
//
// 回傳：
//   - 已完成初始化的 Server 實例。
func NewServer(conf *HttpAPIServerConfig, handler HandlerFunc, logger LoggerFunc) *Server {
	// 初始化全域限流器，供所有請求共用。
	globalLimiter = newRateLimiter(conf)

	s := &Server{
		config:  conf,
		handler: handler,
		logger:  logger,
	}

	// 初始化底層 http.Server，並將所有請求統一交由 Server.ServeHTTP 處理。
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		Handler:      s,
		ReadTimeout:  Second(conf.ReadTimeout),
		WriteTimeout: Second(conf.WriteTimeout),
		IdleTimeout:  Second(conf.IdleTimeout),
	}

	return s
}

// Start 啟動伺服器並開始監聽連線。
//
// 此函式會先輸出 W3C 標頭資訊（若已設定 Logger），再依照 TLS 設定決定：
// 1. 啟動 HTTPS 伺服器。
// 2. 或啟動一般 HTTP 伺服器。
//
// 參數：
//   - serverName: 伺服器名稱，用於 W3C 標頭輸出。
//   - serverVersion: 伺服器版本，用於 W3C 標頭輸出。
//
// 回傳：
//   - 底層 ListenAndServe 或 ListenAndServeTLS 所回傳的錯誤。
func (s *Server) Start(serverName string, serverVersion string) error {
	// 若有設定 Logger，先輸出 W3C 標頭與啟動資訊。
	if s.logger != nil {
		lines := GetW3CHeader(serverName, serverVersion)
		for _, line := range lines {
			s.logger(line)
		}
		s.logger(fmt.Sprintf("#Listening: %s:%d", s.config.Host, s.config.Port))
	}

	// 若已提供 TLS 憑證與私鑰，則以 HTTPS 模式啟動。
	if s.config.TLSCertFile != "" && s.config.TLSKeyFile != "" {
		return s.httpServer.ListenAndServeTLS(s.config.TLSCertFile, s.config.TLSKeyFile)
	}

	// 否則以一般 HTTP 模式啟動。
	return s.httpServer.ListenAndServe()
}

// Stop 以優雅關閉模式停止伺服器。
//
// 此函式會通知底層 http.Server 停止接受新連線，並等待既有請求於 ctx
// 所限定的生命週期內完成，適合用於正式環境的平滑下線流程。
//
// 參數：
//   - ctx: 控制關閉等待時間與取消時機的 Context。
//
// 回傳：
//   - Shutdown 執行結果所回傳的錯誤。
func (s *Server) Stop(ctx context.Context) error {
	if s.logger != nil {
		s.logger(fmt.Sprintf("#Stopping: %s", time.Now().UTC().Format("2006-01-02 15:04:05")))
	}
	return s.httpServer.Shutdown(ctx)
}

// ServeHTTP 實作 http.Handler 介面，為所有 HTTP 請求的唯一統一入口。
//
// 處理流程包含：
// 1. 請求耗時統計。
// 2. 連線數與請求數統計。
// 3. 用戶端真實 IP 解析。
// 4. 限流與封鎖檢查。
// 5. 請求轉換為自定義 HTTPRequest。
// 6. 呼叫業務邏輯處理器。
// 7. 回應寫入與流量統計。
// 8. W3C 存取日誌輸出。
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 記錄請求開始時間，用於後續計算請求總耗時。
	startTime := time.Now()

	// 統計目前活躍連線數，並在請求結束後確保正確遞減。
	recordConnIn()
	defer recordConnOut()

	// 累加總請求數。
	recordRequest()

	// 取得用戶端真實 IP，優先考慮反向代理轉送資訊。
	ip := s.getClientIP(r)

	// 先執行限流與封鎖檢查；若不允許則直接回應 429。
	if !globalLimiter.Allow(ip) {
		statusCode := http.StatusTooManyRequests
		http.Error(w, "Too Many Requests / IP Blocked", statusCode)

		// 即使請求遭拒，仍保留一筆完整 W3C 日誌，便於稽核與追蹤。
		if s.logger != nil {
			logLine := FormatW3CLine(ip, r.Method, r.URL.Path, r.URL.RawQuery,
				s.config.Port, r.UserAgent(), statusCode, time.Since(startTime))
			s.logger(logLine)
		}
		return
	}

	// 將標準 http.Request 轉換為內部使用的 HTTPRequest 結構。
	reqData := s.parseRequest(r)

	// 統計接收流量，以 Body 位元組數為準。
	recordTrafficRecv(len(reqData.Body))

	// 呼叫外部注入的業務邏輯處理器產生回應。
	resp := s.handler(reqData)

	// 若處理器異常回傳 nil，則補上預設 500 狀態碼，避免回應不完整。
	if resp == nil {
		resp = &HTTPResponse{StatusCode: http.StatusInternalServerError}
	}

	// 寫入回應標頭。
	if resp.Headers != nil {
		for k, v := range resp.Headers {
			w.Header().Set(k, v)
		}
	}

	// 寫入 HTTP 狀態碼。
	w.WriteHeader(resp.StatusCode)

	// 寫入回應本文，並統計送出流量。
	if resp.Body != nil {
		bytesSent, _ := w.Write(resp.Body)
		recordTrafficSent(bytesSent)
	}

	// 請求完成後輸出 W3C 標準日誌。
	if s.logger != nil { // 改为 s.logger
		logLine := FormatW3CLine(ip, r.Method, r.URL.Path, r.URL.RawQuery,
			s.config.Port, r.UserAgent(), resp.StatusCode, time.Since(startTime))
		s.logger(logLine)
	}
}

// parseRequest 將標準庫的 http.Request 轉換為內部使用的 HTTPRequest。
//
// 轉換內容包含：
// 1. HTTP 方法與路徑。
// 2. 用戶端來源 IP。
// 3. Header、Cookie 與查詢參數。
// 4. x-www-form-urlencoded 表單資料。
// 5. 原始 Body 內容。
//
// 參數：
//   - r: 原始 HTTP 請求物件。
//
// 回傳：
//   - 轉換完成的 HTTPRequest 指標。
func (s *Server) parseRequest(r *http.Request) *HTTPRequest {
	customReq := &HTTPRequest{
		Method:     r.Method,
		Path:       r.URL.Path,
		RemoteAddr: s.getClientIP(r),
		Headers:    make(map[string]string),
		Cookies:    make(map[string]string),
		Params:     make(map[string]string),
	}

	// 複製所有 Header，統一轉為簡單字串映射結構。
	for k := range r.Header {
		customReq.Headers[k] = r.Header.Get(k)
	}

	// 複製所有 Cookie，方便上層直接依名稱讀取。
	for _, cookie := range r.Cookies() {
		customReq.Cookies[cookie.Name] = cookie.Value
	}

	// 解析 URL Query 參數。
	query := r.URL.Query()
	for k := range query {
		customReq.Params[k] = query.Get(k)
	}

	// 若為表單提交，則補充解析 POST 表單欄位並合併至 Params。
	if strings.Contains(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		r.ParseForm()
		for k := range r.PostForm {
			customReq.Params[k] = r.PostForm.Get(k)
		}
	}

	// 讀取完整 Body 內容供後續業務邏輯使用。
	body, _ := io.ReadAll(r.Body)
	customReq.Body = body
	defer r.Body.Close()

	return customReq
}

// getClientIP 解析並回傳請求的真實來源 IP。
//
// 解析策略：
// 1. 若存在 X-Forwarded-For，優先取其第一個 IP。
// 2. 否則退回使用 r.RemoteAddr 解析來源位址。
//
// 參數：
//   - r: 原始 HTTP 請求物件。
//
// 回傳：
//   - 解析後的用戶端 IP 字串。
func (s *Server) getClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		return strings.Split(xForwardedFor, ",")[0]
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// GetStats 回傳目前伺服器的執行統計資訊。
//
// 統計內容包含：
// 1. 總請求數。
// 2. 當前活躍連線數。
// 3. 累計傳送與接收位元組數。
// 4. 目前封鎖中的 IP 清單。
// 5. 自啟動以來的運行時間。
func GetStats() ServerStats {
	return ServerStats{
		TotalRequests:  globalStats.totalRequests,
		CurrentConns:   globalStats.currentConns,
		TotalBytesSent: globalStats.totalBytesSent,
		TotalBytesRecv: globalStats.totalBytesRecv,
		BlockedIPs:     globalLimiter.GetBlockedIPs(),
		Uptime:         time.Since(globalStats.startTime).String(),
	}
}
