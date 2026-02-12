package nyaapiserver

// HTTPRequest 封裝了傳遞給外部程式的請求資訊
// 遮蔽了底層 http.Request 的複雜性，只暴露常用欄位
type HTTPRequest struct {
	Method     string            // 請求方法 (GET, POST, etc.)
	Path       string            // 請求路徑 (例如 /api/v1/user)
	RemoteAddr string            // 客戶端真實 IP 地址
	Headers    map[string]string // 簡化的請求頭
	Cookies    map[string]string // 簡化的 Cookie (Key-Value)
	Params     map[string]string // 包含 URL 引數和 POST 表單資料
	Body       []byte            // 請求體原始資料
}

// HTTPResponse 外部程式處理完邏輯後，需要返回給本 package 的結構體
type HTTPResponse struct {
	StatusCode int               // HTTP 狀態碼 (如 200, 404, 500)
	Body       []byte            // 返回給客戶端的資料
	Headers    map[string]string // 需要額外設定的響應頭
}

// HandlerFunc 是外部程式需要實現的回撥函式原型
// 外部程式只需要寫一個符合這個簽名的函式，即可處理所有邏輯
type HandlerFunc func(req *HTTPRequest) *HTTPResponse

// ServerStats 用於匯出伺服器當前的執行狀態
type ServerStats struct {
	TotalRequests  int64    // 自啟動以來的總請求數
	CurrentConns   int64    // 當前活躍連線數
	TotalBytesSent int64    // 已傳送的總流量 (位元組)
	TotalBytesRecv int64    // 已接收的總流量 (位元組)
	BlockedIPs     []string // 當前正處於封鎖狀態的惡意 IP 列表
	Uptime         string   // 伺服器執行時間
}
