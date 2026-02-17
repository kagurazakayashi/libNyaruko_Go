package nyaapiserver

// HTTPRequest 封裝傳遞給外部程式的 HTTP 請求資訊。
// 此結構用於抽象化底層 http.Request，降低外部處理邏輯對標準函式庫細節的耦合，
// 僅保留常用且適合業務層存取的欄位。
type HTTPRequest struct {
	Method     string            // HTTP 請求方法，例如 GET、POST。
	Path       string            // 請求路徑，例如 /api/v1/user。
	RemoteAddr string            // 用戶端的實際 IP 位址。
	Headers    map[string]string // 簡化後的 HTTP 標頭集合，採用 Key-Value 形式儲存。
	Cookies    map[string]string // 簡化後的 Cookie 集合，採用 Key-Value 形式儲存。
	Params     map[string]string // 請求參數集合，包含 URL 查詢參數與 POST 表單資料。
	Body       []byte            // HTTP 請求本文的原始位元組資料。
}

// HTTPResponse 定義外部程式處理完成後，回傳給本 package 的標準回應結構。
// 呼叫端可透過此結構指定 HTTP 狀態碼、回應本文以及額外的回應標頭。
type HTTPResponse struct {
	StatusCode int               // HTTP 狀態碼，例如 200、404、500。
	Body       []byte            // 要回傳給用戶端的回應本文資料。
	Headers    map[string]string // 需要額外寫入的 HTTP 回應標頭。
}

// HandlerFunc 定義外部程式需要實作的回呼函式簽章。
// 外部程式只需提供一個符合此簽章的函式，即可接收請求並回傳處理結果。
type HandlerFunc func(req *HTTPRequest) *HTTPResponse

// ServerStats 用於匯出伺服器目前的執行統計資訊。
// 可供監控、診斷或管理介面查詢伺服器的即時狀態。
type ServerStats struct {
	TotalRequests  int64    // 伺服器啟動以來累計收到的總請求數。
	CurrentConns   int64    // 目前仍處於活躍狀態的連線數。
	TotalBytesSent int64    // 伺服器累計已傳送的總位元組數。
	TotalBytesRecv int64    // 伺服器累計已接收的總位元組數。
	BlockedIPs     []string // 目前處於封鎖狀態的惡意 IP 清單。
	Uptime         string   // 伺服器自啟動以來的持續運行時間。
}
