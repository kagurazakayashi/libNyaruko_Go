package nyaapiserver

import (
	"sync/atomic"
	"time"
)

// statistics 內部統計結構體
type statistics struct {
	startTime      time.Time
	totalRequests  int64
	currentConns   int64
	totalBytesSent int64
	totalBytesRecv int64
}

// 全域性統計例項
var globalStats = &statistics{
	startTime: time.Now(),
}

// recordRequest 增加總請求數
func recordRequest() {
	atomic.AddInt64(&globalStats.totalRequests, 1)
}

// recordConnIn 增加當前活躍連線
func recordConnIn() {
	atomic.AddInt64(&globalStats.currentConns, 1)
}

// recordConnOut 減少當前活躍連線
func recordConnOut() {
	atomic.AddInt64(&globalStats.currentConns, -1)
}

// recordTrafficSent 統計傳送的位元組數
func recordTrafficSent(bytes int) {
	atomic.AddInt64(&globalStats.totalBytesSent, int64(bytes))
}

// recordTrafficRecv 統計接收的位元組數
func recordTrafficRecv(bytes int) {
	atomic.AddInt64(&globalStats.totalBytesRecv, int64(bytes))
}
