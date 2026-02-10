// Package nyanats 提供了一個帶有加密功能的 NATS 用戶端封裝。
package nyanats

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"gopkg.in/yaml.v3"
)

// NyaNATS 是此套件的核心結構，封裝了 NATS 連線以及加解密所需的邏輯與金鑰。
type NyaNATS struct {
	err        error             // 存放最近一次發生的錯誤
	natsConn   *nats.Conn        // NATS 的底層連線實例
	debug      *log.Logger       // 用於輸出除錯資訊的日誌記錄器
	defaultKey []byte            // 預設使用的對稱加密金鑰 (AES)
	themeKeys  map[string][]byte // 針對不同主題 (Subject) 獨立設定的金鑰對照表
}

// logf 是一個內部的輔助方法，用於格式化並輸出 NyaNATS 的除錯日誌。
func (p *NyaNATS) logf(format string, v ...interface{}) {
	if p.debug != nil {
		p.debug.Printf("[NyaNATS] "+format, v...)
	}
}

// New 是工廠函式，支援傳入 JSON 或 YAML 格式的配置字串來建立 NyaNATS 實例。
func New(configString string, debug *log.Logger) *NyaNATS {
	var natsConfig NATSConfig
	// 優先嘗試解析為 JSON
	if err := json.Unmarshal([]byte(configString), &natsConfig); err == nil {
		return NewC(natsConfig, debug)
	}
	// 若失敗則嘗試解析為 YAML
	if err := yaml.Unmarshal([]byte(configString), &natsConfig); err == nil {
		return NewC(natsConfig, debug)
	}
	// 若皆失敗則回傳帶有錯誤標記的物件
	return &NyaNATS{err: fmt.Errorf("E: CONF"), debug: debug}
}

// NewC 根據傳入的 NATSConfig 結構體實例來初始化 NyaNATS。
func NewC(config NATSConfig, debug *log.Logger) *NyaNATS {
	// 1. 補齊配置的預設值
	config.setDefaults()

	// 2. 如果未傳入 Logger，則檢查環境變數 DEBUG 是否包含 "NYANATS" 來決定是否啟用日誌
	if debug == nil {
		debugEnv := os.Getenv("DEBUG")
		if strings.Contains(strings.ToUpper(debugEnv), "NYANATS") {
			debug = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
		}
	}

	p := &NyaNATS{
		debug:     debug,
		themeKeys: make(map[string][]byte),
	}

	// 3. 處理預設加密金鑰，檢查長度是否符合 AES 規範 (16, 24, 32 bytes)
	if config.EncryptionKey != "" {
		p.defaultKey = []byte(config.EncryptionKey)
		l := len(p.defaultKey)
		if l != 16 && l != 24 && l != 32 {
			p.err = fmt.Errorf("S# [](K%d) ERR: KEYLEN", l)
			fmt.Println(p.err.Error())
			return p
		}
		p.logf("S# [](K%d)", l)
	} else {
		p.logf("S# [](K0)")
	}

	// 4. 處理各主題 (Theme/Subject) 的專屬金鑰
	for theme, kStr := range config.ThemeKeys {
		if kStr == "" {
			p.themeKeys[theme] = nil
			p.logf("S# [%s](K0)", theme)
			continue
		}

		kByte := []byte(kStr)
		l := len(kByte)
		if l != 16 && l != 24 && l != 32 {
			p.err = fmt.Errorf("S# [%s](K%d) ERR: KEYLEN", theme, l)
			fmt.Println(p.err.Error())
			return p
		}
		p.themeKeys[theme] = kByte
		p.logf("S# [%s](K%d)", theme, l)
	}

	// 5. 組合 NATS 連線 URL (支援帶有帳號密碼的格式)
	scheme := "nats:/"
	url := fmt.Sprintf("%s/%s", scheme, config.NatsServer)
	if config.NatsUser != "" {
		url = fmt.Sprintf("%s/%s:%s@%s", scheme, config.NatsUser, config.NatsPassword, config.NatsServer)
	}

	// 6. 設定 NATS 連線選項 (如重連次數、超時時間與各種事件處理器)
	opts := []nats.Option{
		nats.Name(config.ClientName),
		nats.MaxReconnects(config.MaxReconnects),
		nats.ReconnectWait(time.Duration(config.ReconnectWait) * time.Second),
		nats.Timeout(time.Duration(config.ConnectTimeout) * time.Second),
		// 斷線處理
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				p.logf("L- [%s] ERR: %v", p.natsConn.Opts.Name, err)
			}
		}),
		// 重連成功處理
		nats.ReconnectHandler(func(nc *nats.Conn) { p.logf("L+ [%v]", nc.ConnectedUrl()) }),
		// 非同步錯誤處理
		nats.ErrorHandler(func(nc *nats.Conn, s *nats.Subscription, err error) { p.logf("<- [%s] ERR: %v", s.Subject, err) }),
	}

	// 7. 建立連線
	nc, err := nats.Connect(url, opts...)
	if err != nil {
		p.err = err
		return p
	}

	p.natsConn = nc
	p.logf("L+ [%s]", config.ClientName)
	return p
}

// Close 安全地關閉 NATS 連線並記錄日誌。
func (p *NyaNATS) Close() {
	if p.natsConn != nil {
		p.logf("L- [%s]", p.natsConn.Opts.Name)
		p.natsConn.Close()
	}
}

// Error 傳回 NyaNATS 實例當前的錯誤狀態，方便外部呼叫者檢查初始化是否成功。
func (p *NyaNATS) Error() error { return p.err }
