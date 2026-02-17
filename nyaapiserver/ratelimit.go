package nyaapiserver

import (
	"sync"
	"time"
)

// ipStat 記錄單一 IP 在限流機制中的統計資訊。
// 包含目前時間視窗內的請求次數、視窗起始時間，以及封鎖截止時間。
type ipStat struct {
	count        int       // 目前時間視窗內的請求數量
	windowStart  time.Time // 目前計數視窗的起始時間
	blockedUntil time.Time // 該 IP 被封鎖到何時為止；零值代表未封鎖
}

// RateLimiter 提供基於 IP 的請求頻率限制能力。
// 其責任為：
// 1. 追蹤各 IP 在指定時間視窗內的請求次數。
// 2. 當請求次數超過限制時，暫時封鎖該 IP。
// 3. 提供目前封鎖中 IP 清單的查詢能力。
type RateLimiter struct {
	mu    sync.RWMutex
	stats map[string]*ipStat
	conf  *HttpAPIServerConfig
}

// newRateLimiter 建立並初始化一個 RateLimiter 實例。
func newRateLimiter(conf *HttpAPIServerConfig) *RateLimiter {
	return &RateLimiter{
		stats: make(map[string]*ipStat),
		conf:  conf,
	}
}

// Allow 檢查指定 IP 是否允許繼續存取。
// 判斷流程如下：
// 1. 若未啟用限流功能，直接允許。
// 2. 若該 IP 尚無紀錄，建立初始統計並允許。
// 3. 若該 IP 仍處於封鎖期間，拒絕存取。
// 4. 若目前時間已超出統計視窗，重設計數後允許。
// 5. 若未超出視窗，累加請求次數並檢查是否超過限制。
// 6. 若超過限制，設定封鎖截止時間並拒絕存取。
//
// 回傳值：
//   - true：允許本次請求
//   - false：拒絕本次請求
func (rl *RateLimiter) Allow(ip string) bool {
	if !rl.conf.EnableRateLimit {
		return true
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	stat, exists := globalLimiter.stats[ip]

	if !exists {
		globalLimiter.stats[ip] = &ipStat{
			count:       1,
			windowStart: now,
		}
		return true
	}

	// 1. 檢查是否仍在封鎖期間，若是則直接拒絕。
	if now.Before(stat.blockedUntil) {
		return false
	}

	// 2. 若目前時間已超出限制視窗，重設計數器並重新開始統計。
	if now.Sub(stat.windowStart) > rl.conf.LimitWindow {
		stat.count = 1
		stat.windowStart = now
		return true
	}

	// 3. 視窗仍有效，累加本次請求並檢查是否超出上限。
	stat.count++
	if stat.count > rl.conf.LimitRequests {
		// 當請求數超過限制時，設定封鎖截止時間。
		stat.blockedUntil = now.Add(rl.conf.BlockDuration)
		return false
	}

	return true
}

// GetBlockedIPs 取得目前仍處於封鎖狀態中的所有 IP 清單。
// 僅回傳封鎖截止時間晚於目前時間的 IP。
func (rl *RateLimiter) GetBlockedIPs() []string {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	var list []string
	now := time.Now()
	for ip, stat := range rl.stats {
		if now.Before(stat.blockedUntil) {
			list = append(list, ip)
		}
	}

	return list
}

var globalLimiter *RateLimiter
