package nyaapiserver

import (
	"sync"
	"time"
)

// ipStat 記錄單個 IP 的訪問詳情
type ipStat struct {
	count        int       // 當前視窗內的請求數
	windowStart  time.Time // 當前視窗的開始時間
	blockedUntil time.Time // 封鎖截止時間
}

// RateLimiter 頻率限制器
type RateLimiter struct {
	mu    sync.RWMutex
	stats map[string]*ipStat
	conf  *HttpAPIServerConfig
}

func newRateLimiter(conf *HttpAPIServerConfig) *RateLimiter {
	return &RateLimiter{
		stats: make(map[string]*ipStat),
		conf:  conf,
	}
}

// Allow 檢查該 IP 是否被允許訪問。如果不允許（被封鎖或觸發限流），返回 false
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

	// 1. 檢查是否正處於封鎖期
	if now.Before(stat.blockedUntil) {
		return false
	}

	// 2. 檢查時間視窗是否已過期，過期則重置計數
	if now.Sub(stat.windowStart) > rl.conf.LimitWindow {
		stat.count = 1
		stat.windowStart = now
		return true
	}

	// 3. 增加計數並檢查是否觸發閾值
	stat.count++
	if stat.count > rl.conf.LimitRequests {
		// 觸發限流，計算封鎖截止時間
		stat.blockedUntil = now.Add(rl.conf.BlockDuration)
		return false
	}

	return true
}

// GetBlockedIPs 獲取當前所有正在被封鎖的 IP 列表
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
