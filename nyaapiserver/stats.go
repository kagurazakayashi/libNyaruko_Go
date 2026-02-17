package nyaapiserver

import (
	"sync/atomic"
	"time"
)

// statistics 定義伺服器執行期間的內部統計資料。
// 此結構體用於追蹤服務啟動時間、累計請求數、目前活躍連線數，
// 以及累計傳送與接收的流量位元組數。
// 各欄位的讀寫預期透過 atomic 套件進行，以確保在並行環境下的資料一致性。
type statistics struct {
	// startTime 為服務啟動時間，用於計算運行時間等統計資訊。
	startTime time.Time

	// totalRequests 為自服務啟動以來累計處理的請求總數。
	totalRequests int64

	// currentConns 為目前仍維持中的活躍連線數。
	currentConns int64

	// totalBytesSent 為自服務啟動以來累計送出的位元組總數。
	totalBytesSent int64

	// totalBytesRecv 為自服務啟動以來累計接收的位元組總數。
	totalBytesRecv int64
}

// globalStats 為全域統計實例，用於彙整整個伺服器程序的執行統計資料。
// 其 startTime 於套件初始化時設定為目前時間，作為統計起點。
var globalStats = &statistics{
	startTime: time.Now(),
}

// recordRequest 將累計請求數加一。
// 此函式應於每次成功接收到新請求時呼叫，以更新全域請求統計。
func recordRequest() {
	atomic.AddInt64(&globalStats.totalRequests, 1)
}

// recordConnIn 將目前活躍連線數加一。
// 此函式應於新連線建立並納入追蹤時呼叫。
func recordConnIn() {
	atomic.AddInt64(&globalStats.currentConns, 1)
}

// recordConnOut 將目前活躍連線數減一。
// 此函式應於連線結束或釋放時呼叫，以反映最新的連線狀態。
func recordConnOut() {
	atomic.AddInt64(&globalStats.currentConns, -1)
}

// recordTrafficSent 累計送出的流量位元組數。
// 參數 bytes 表示本次送出的位元組數，會以原子操作方式累加至全域統計。
func recordTrafficSent(bytes int) {
	atomic.AddInt64(&globalStats.totalBytesSent, int64(bytes))
}

// recordTrafficRecv 累計接收的流量位元組數。
// 參數 bytes 表示本次接收的位元組數，會以原子操作方式累加至全域統計。
func recordTrafficRecv(bytes int) {
	atomic.AddInt64(&globalStats.totalBytesRecv, int64(bytes))
}
