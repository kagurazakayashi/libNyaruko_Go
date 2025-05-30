// Redis 資料庫操作
package nyaredis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	redis "github.com/go-redis/redis/v8"
)

type RedisDBConfig struct {
	Address      string `json:"redis_addr" yaml:"redis_addr"`
	Port         string `json:"redis_port" yaml:"redis_port"`
	Password     string `json:"redis_pwd" yaml:"redis_pwd"`
	DB           int    `json:"redis_db" yaml:"redis_db"`
	PoolSize     int    `json:"redis_poolSize" yaml:"redis_poolSize"`
	MinIdleConns int    `json:"redis_minIdleConns" yaml:"redis_minIdleConns"`
}

// <類>
var ctx = context.Background()

type NyaRedis NyaRedisT
type NyaRedisT struct {
	db  *redis.Client
	err error
}

// </類>

// <可選配置>
type Option struct {
	isDelete    bool // 在查詢完成後刪除此條目
	autoDelete  int  // 資料條目的超時時間（秒）
	isErrorStop bool // 在批次操作中是否遇到錯誤就停止
}
type OptionConfig func(*Option)

func Option_isDelete(v bool) OptionConfig {
	return func(p *Option) {
		p.isDelete = v
	}
}
func Option_autoDelete(v int) OptionConfig {
	return func(p *Option) {
		p.autoDelete = v
	}
}
func Option_isErrorStop(v bool) OptionConfig {
	return func(p *Option) {
		p.isErrorStop = v
	}
}

// </可選配置>

// New: 建立新的 NyaRedis 例項
//
//	`configJsonString` string 配置 JSON 字串
//	從配置 JSON 檔案中取出的本模組所需的配置段落 JSON 字串
//	示例配置數值參考 config.template.json
//	本模組所需配置項: redis_addr, redis_port, redis_pwd, redis_db
//	return NyaRedis 新的 NyaRedis 例項
//	下一步使用 `Error()` 或 `ErrorString()` 檢查是否有錯誤
func New(configJsonString string) *NyaRedis {
	return NewDB(configJsonString, -1, 255)
}

// New: 建立新的 NyaRedis 例項
//
//	`configJsonString` string 配置 JSON 字串
//	從配置 JSON 檔案中取出的本模組所需的配置段落 JSON 字串
//	示例配置數值參考 config.template.json
//	本模組所需配置項: redis_addr, redis_port, redis_pwd, redis_db
//	return NyaRedis 新的 NyaRedis 例項
//	下一步使用 `Error()` 或 `ErrorString()` 檢查是否有錯誤
func NewDB(configJsonString string, dbID int, maxDB int) *NyaRedis {
	var redisConfig RedisDBConfig
	err := json.Unmarshal([]byte(configJsonString), &redisConfig)
	if err != nil {
		return &NyaRedis{err: err}
	}
	return NewC(redisConfig, dbID, maxDB)
}

// NewC: 建立新的 NyaRedis 例項
//
//	`config` RedisDBConfig 配置
//	下一步使用 `Error()` 或 `ErrorString()` 檢查是否有錯誤
func NewC(config RedisDBConfig, dbID int, maxDB int) *NyaRedis {
	var dbid int = dbID
	if !(0 <= dbID && dbID <= maxDB) {
		dbid = config.DB
	}

	if config.PoolSize == 0 {
		config.PoolSize = 20
	}
	if config.MinIdleConns == 0 {
		config.MinIdleConns = 5
	}

	var nRedisDB *redis.Client = redis.NewClient(&redis.Options{
		Addr:         config.Address + ":" + config.Port,
		Password:     config.Password,
		DB:           dbid,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
	})
	_, err := nRedisDB.Ping(nRedisDB.Context()).Result()
	if err != nil {
		return &NyaRedis{err: err}
	}
	return &NyaRedis{db: nRedisDB, err: nil}
}

// Close: 關閉資料庫連線
func (p *NyaRedis) Close() {
	if p.db != nil {
		p.db.Close()
		p.db = nil
	}
}

// Error: 獲取上一次操作時可能產生的錯誤
//
//	return error 如果有錯誤，返回錯誤物件，如果沒有錯誤返回 nil
func (p *NyaRedis) Error() error {
	return p.err
}

// ErrorString: 獲取上一次操作時可能產生的錯誤資訊字串
//
//	return string 如果有錯誤，返回錯誤描述字串，如果沒有錯誤返回空字串
func (p *NyaRedis) ErrorString() string {
	if p.err == nil {
		return ""
	}
	return p.err.Error()
}

// SetString: 向資料庫中新增字串資料
//
//		`key`  string 資料名稱
//		`val`  string 資料內容
//	 `options` ...OptionConfig 可選配置，執行 `Option_*` 函式輸入
//			`autoDelete` int 資料條目的超時時間(秒)，預設 `0` (不限時)
//		return bool 操作是否成功，若不成功可使用 `Error()` 或 `ErrorString()` 獲取錯誤資訊
func (p *NyaRedis) SetString(key string, val string, options ...OptionConfig) bool {
	option := &Option{autoDelete: 0}
	for _, o := range options {
		o(option)
	}
	if option.autoDelete == 0 {
		p.err = p.db.Set(ctx, key, val, 0).Err()
	} else {
		p.err = p.db.Set(ctx, key, val, time.Duration(option.autoDelete)*time.Second).Err()
	}
	return p.err == nil
}

// GetValTime: 從資料庫中取出字串值
//
//	`key`  string 資料名稱
//	`options` ...OptionConfig 可選配置，執行 `Option_*` 函式輸入
//		`isDelete` bool 是否在查詢完成後刪除此條目，預設值 `false`
//	return
//		time.Duration 取出的資料時間。如果不成功則返回空時間，可使用 `Error()` 或 `ErrorString()` 檢查是否發生錯誤或獲取錯誤資訊
func (p *NyaRedis) GetTTL(key string, options ...OptionConfig) time.Duration {
	option := &Option{isDelete: false}
	for _, o := range options {
		o(option)
	}

	ttl, err := p.db.TTL(ctx, key).Result()
	p.err = err
	if err != nil {
		return ttl
	}
	if option.isDelete {
		p.err = p.db.Del(ctx, key).Err()
		if p.err != nil {
			return ttl
		}
	}
	return ttl
}

// GetString: 從資料庫中取出字串值
//
//	`key`  string 資料名稱
//	`options` ...OptionConfig 可選配置，執行 `Option_*` 函式輸入
//		`isDelete` bool 是否在查詢完成後刪除此條目，預設值 `false`
//	return string 取出的字串。如果不成功則返回空字串，可使用 `Error()` 或 `ErrorString()` 檢查是否發生錯誤或獲取錯誤資訊
func (p *NyaRedis) GetString(key string, options ...OptionConfig) string {
	option := &Option{isDelete: false}
	for _, o := range options {
		o(option)
	}
	val, err := p.db.Get(ctx, key).Result()
	p.err = err
	if err != nil {
		return ""
	}
	if option.isDelete {
		p.err = p.db.Del(ctx, key).Err()
		if p.err != nil {
			return val
		}
	}
	return val
}

// GetValTime: 從資料庫中取出字串值
//
//	`key`  string 資料名稱
//	`options` ...OptionConfig 可選配置，執行 `Option_*` 函式輸入
//		`isDelete` bool 是否在查詢完成後刪除此條目，預設值 `false`
//	return
//		string 取出的字串。如果不成功則返回空字串，可使用 `Error()` 或 `ErrorString()` 檢查是否發生錯誤或獲取錯誤資訊
//		time.Duration 取出的資料時間。如果不成功則返回空時間，可使用 `Error()` 或 `ErrorString()` 檢查是否發生錯誤或獲取錯誤資訊
func (p *NyaRedis) GetStringAndTTL(key string, options ...OptionConfig) (string, time.Duration) {
	option := &Option{isDelete: false}
	for _, o := range options {
		o(option)
	}

	var ttl time.Duration
	val, err := p.db.Get(ctx, key).Result()
	p.err = err
	if err != nil {
		return "", ttl
	}
	ttl, err = p.db.TTL(ctx, key).Result()
	p.err = err
	if err != nil {
		return "", ttl
	}
	if option.isDelete {
		p.err = p.db.Del(ctx, key).Err()
		if p.err != nil {
			return val, ttl
		}
	}
	return val, ttl
}

// GetStringAll: 從資料庫中批次取出字串值
//
//	`keyPattern` string 資料名稱(包含萬用字元 `*`, 例如 `prefix*` )
//	`options` ...OptionConfig 可選配置，執行 `Option_*` 函式輸入
//		`isDelete`    bool 是否在查詢完成後刪除此條目，預設值 `false`
//		`isErrorStop` bool 在批次操作中是否遇到錯誤就停止，否則忽略錯誤，會記錄最近一次錯誤，預設值 `false`
//	return map[string]string 取出的資料鍵值字典。如果不成功則返回空字典，可使用 `Error()` 或 `ErrorString()` 檢查是否發生錯誤或獲取錯誤資訊
func (p *NyaRedis) GetStringAll(keyPattern string, options ...OptionConfig) map[string]string {
	option := &Option{isDelete: false, isErrorStop: false}
	for _, o := range options {
		o(option)
	}
	var data map[string]string = make(map[string]string)
	iter := p.db.Scan(ctx, 0, keyPattern, 0).Iterator()
	for iter.Next(ctx) {
		var key string = iter.Val()
		var val string = p.GetString(key, options...)
		if p.err != nil {
			if option.isErrorStop {
				return nil
			}
		} else {
			data[key] = val
		}
	}
	if p.err = iter.Err(); p.err != nil {
		return nil
	}
	return data
}

// Keys: 按照萬用字元字串獲取鍵名稱列表
//
//	`keyPattern` string 資料名稱(包含萬用字元 `*`, 例如 `prefix*` )
//	return 取出的鍵名字串陣列。如果不成功則返回空陣列，可使用 `Error()` 或 `ErrorString()` 檢查是否發生錯誤或獲取錯誤資訊
func (p *NyaRedis) Keys(keyPattern string) []string {
	keys, err := p.db.Keys(ctx, keyPattern).Result()
	p.err = err
	if err != nil {
		return []string{}
	}
	return keys
}

// Delete: 刪除資料
//
//	`keys` []string 要刪除的資料名稱陣列（可刪除多條）
//	`options` ...OptionConfig 可選配置，執行 `Option_*` 函式輸入
//	`isErrorStop` bool 在批次操作中是否遇到錯誤就停止，否則忽略錯誤，會記錄最近一次錯誤，預設值 `false`
//	return bool 操作是否成功，只要批次操作中有一次失敗即為 false ，若不成功可使用 `Error()` 或 `ErrorString()` 獲取錯誤資訊
func (p *NyaRedis) Delete(keys []string, options ...OptionConfig) error {
	option := &Option{isErrorStop: false}
	for _, o := range options {
		o(option)
	}
	for _, k := range keys {
		p.err = p.db.Del(ctx, k).Err()
		if p.err != nil && option.isErrorStop {
			return fmt.Errorf("Delete key:", k, "error[", p.err.Error(), "]")
		}
	}
	return nil
}

// DeleteMulti: 根據萬用字元資料名稱刪除資料
//
//	`keyPattern` string 資料名稱(包含萬用字元 `*`, 例如 `prefix*` )
//	`options` ...OptionConfig 可選配置，執行 `Option_*` 函式輸入
//	`isErrorStop` bool 在批次操作中是否遇到錯誤就停止，否則忽略錯誤，會記錄最近一次錯誤，預設值 `false`
//	return bool 操作是否成功，只要批次操作中有一次失敗即為 false ，若不成功可使用 `Error()` 或 `ErrorString()` 獲取錯誤資訊
func (p *NyaRedis) DeleteMulti(keyPattern string, options ...OptionConfig) bool {
	option := &Option{isErrorStop: false}
	for _, o := range options {
		o(option)
	}
	iter := p.db.Scan(ctx, 0, keyPattern, 0).Iterator()
	for iter.Next(ctx) {
		p.err = p.db.Del(ctx, iter.Val()).Err()
		if p.err != nil && option.isErrorStop {
			return false
		}
	}
	if p.err = iter.Err(); p.err != nil {
		return false
	}
	return true
}
