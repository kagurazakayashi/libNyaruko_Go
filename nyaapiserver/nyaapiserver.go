package nyaapiserver

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

// Server 是 HTTP API 伺服器的主體
type Server struct {
	httpServer *http.Server
	config     *HttpAPIServerConfig
	handler    HandlerFunc
}

// NewServer 建立一個新的伺服器例項
func NewServer(conf *HttpAPIServerConfig, handler HandlerFunc) *Server {

	s := &Server{
		config:  conf,
		handler: handler,
	}

	// 初始化底層的 http.Server
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		Handler:      s, // 讓 Server 結構體處理所有請求
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
		IdleTimeout:  conf.IdleTimeout,
	}

	return s
}

// Start 啟動伺服器
func (s *Server) Start() error {
	fmt.Printf("Starting server on %s:%d...\n", s.config.Host, s.config.Port)

	// 判斷是否啟用 HTTPS
	if s.config.TLSCertFile != "" && s.config.TLSKeyFile != "" {
		return s.httpServer.ListenAndServeTLS(s.config.TLSCertFile, s.config.TLSKeyFile)
	}
	return s.httpServer.ListenAndServe()
}

// Stop 優雅地關閉伺服器
func (s *Server) Stop(ctx context.Context) error {
	fmt.Println("Shutting down server...")
	return s.httpServer.Shutdown(ctx)
}

// ServeHTTP 實現了 http.Handler 介面，是所有請求的入口
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// 獲取客戶端 IP (處理代理情況)
	s.getClientIP(r)

	// 解析並轉換請求資料
	reqData := s.parseRequest(r)

	// 回撥外部程式邏輯
	resp := s.handler(reqData)

	// 處理並寫回響應
	if resp == nil {
		resp = &HTTPResponse{StatusCode: 500}
	}

	// 設定響應頭
	for k, v := range resp.Headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(resp.StatusCode)

	// 寫入 Body 並統計傳送流量
	if resp.Body != nil {
		w.Write(resp.Body)
	}
}

// parseRequest 將標準 http.Request 轉換為自定義的 HTTPRequest
func (s *Server) parseRequest(r *http.Request) *HTTPRequest {
	customReq := &HTTPRequest{
		Method:     r.Method,
		Path:       r.URL.Path,
		RemoteAddr: s.getClientIP(r),
		Headers:    make(map[string]string),
		Cookies:    make(map[string]string),
		Params:     make(map[string]string),
	}

	// 解析 Header
	for k := range r.Header {
		customReq.Headers[k] = r.Header.Get(k)
	}

	// 解析 Cookie
	for _, cookie := range r.Cookies() {
		customReq.Cookies[cookie.Name] = cookie.Value
	}

	// 解析 URL 引數
	query := r.URL.Query()
	for k := range query {
		customReq.Params[k] = query.Get(k)
	}

	// 解析 Post 表單 (如果是 POST 請求)
	if strings.Contains(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		r.ParseForm()
		for k := range r.PostForm {
			customReq.Params[k] = r.PostForm.Get(k)
		}
	}

	// 讀取 Body
	body, _ := io.ReadAll(r.Body)
	customReq.Body = body
	defer r.Body.Close()

	return customReq
}

// getClientIP 獲取真實的 IP
func (s *Server) getClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		return strings.Split(xForwardedFor, ",")[0]
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
