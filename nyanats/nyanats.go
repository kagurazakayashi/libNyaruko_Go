package nyanats

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"gopkg.in/yaml.v3"
)

type NATSConfig struct {
	NatsServer     string `json:"nats_server" yaml:"nats_server"`
	NatsUser       string `json:"nats_user" yaml:"nats_user"`
	NatsPassword   string `json:"nats_password" yaml:"nats_password"`
	ClientName     string `json:"client_name" yaml:"client_name"`
	MaxReconnects  int    `json:"max_reconnects" yaml:"max_reconnects"`
	ReconnectWait  int    `json:"reconnect_wait" yaml:"reconnect_wait"`
	ConnectTimeout int    `json:"connect_timeout" yaml:"connect_timeout"`
}

func (c *NATSConfig) setDefaults() {
	if c.NatsServer == "" {
		c.NatsServer = "127.0.0.1:4222"
	}
	if c.ClientName == "" {
		c.ClientName = fmt.Sprintf("NyaNATS-%s", uuid.NewString())
	}
	if c.MaxReconnects == 0 {
		c.MaxReconnects = 5
	}
	if c.ReconnectWait == 0 {
		c.ReconnectWait = 2
	}
	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = 10
	}
}

type NyaNATS struct {
	err      error
	natsConn *nats.Conn
	debug    *log.Logger
}

func (p *NyaNATS) logf(format string, v ...interface{}) {
	if p.debug != nil {
		p.debug.Printf("[NyaNATS] "+format, v...)
	}
}

func New(configString string, debug *log.Logger) *NyaNATS {
	var natsConfig NATSConfig

	tmp := &NyaNATS{debug: debug}

	if err := json.Unmarshal([]byte(configString), &natsConfig); err == nil {
		tmp.logf("检测到 JSON 配置格式")
		return NewC(natsConfig, debug)
	}
	if err := yaml.Unmarshal([]byte(configString), &natsConfig); err == nil {
		tmp.logf("检测到 YAML 配置格式")
		return NewC(natsConfig, debug)
	}

	tmp.logf("配置解析失败: 既不是合法的 JSON 也不是 YAML")
	return &NyaNATS{err: fmt.Errorf("config parse error"), debug: debug}
}

func NewC(config NATSConfig, debug *log.Logger) *NyaNATS {
	config.setDefaults()
	p := &NyaNATS{debug: debug}

	var url string
	if config.NatsUser != "" {
		url = fmt.Sprintf("nats://%s:%s@%s", config.NatsUser, config.NatsPassword, config.NatsServer)
		p.logf("正在尝试连接 (用户认证模式): %s", config.NatsServer)
	} else {
		url = fmt.Sprintf("nats://%s", config.NatsServer)
		p.logf("正在尝试连接 (匿名模式): %s", config.NatsServer)
	}

	opts := []nats.Option{
		nats.Name(config.ClientName),
		nats.MaxReconnects(config.MaxReconnects),
		nats.ReconnectWait(time.Duration(config.ReconnectWait) * time.Second),
		nats.Timeout(time.Duration(config.ConnectTimeout) * time.Second),

		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			p.logf("警告: 链接断开 - %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			p.logf("成功: 已重新连接至 %v", nc.ConnectedUrl())
		}),
		nats.ErrorHandler(func(nc *nats.Conn, s *nats.Subscription, err error) {
			p.logf("异步错误: 主题 [%s] 发生异常 - %v", s.Subject, err)
		}),
	}

	nc, err := nats.Connect(url, opts...)
	if err != nil {
		p.logf("错误: 无法建立初始连接 - %v", err)
		p.err = err
		return p
	}

	p.natsConn = nc
	p.logf("连接成功! 客户端 ID: %s", config.ClientName)
	return p
}

func (p *NyaNATS) Subscribe(theme string, callback func(m string) string) error {
	if p.err != nil {
		return p.err
	}

	_, err := p.natsConn.Subscribe(theme, func(m *nats.Msg) {
		p.logf("收到消息 <- 主题 [%s], 长度: %d 字节", theme, len(m.Data))

		replyContent := callback(string(m.Data))

		if m.Reply != "" {
			p.logf("回复响应 -> 回传地址 [%s], 长度: %d 字节", m.Reply, len(replyContent))
			if err := m.Respond([]byte(replyContent)); err != nil {
				p.logf("回复失败: %v", err)
			}
		}
	})

	if err != nil {
		p.logf("订阅失败: 主题 [%s] - %v", theme, err)
	} else {
		p.logf("已成功订阅主题: [%s]", theme)
	}
	return err
}

func (p *NyaNATS) Publish(theme string, message string) error {
	if p.err != nil {
		return p.err
	}

	err := p.natsConn.Publish(theme, []byte(message))
	if err != nil {
		p.logf("发布失败: 主题 [%s] - %v", theme, err)
	} else {
		p.logf("已发布消息 -> 主题 [%s]", theme)
	}
	return err
}

func (p *NyaNATS) Request(theme string, message string, timeout time.Duration) (string, error) {
	if p.err != nil {
		return "", p.err
	}

	p.logf("发起请求 -> 主题 [%s], 等待超时: %v", theme, timeout)
	msg, err := p.natsConn.Request(theme, []byte(message), timeout)
	if err != nil {
		p.logf("请求超时或失败: %v", err)
		return "", err
	}

	p.logf("收到响应 <- 来自主题 [%s]", theme)
	return string(msg.Data), nil
}

func (p *NyaNATS) Close() {
	if p.natsConn != nil {
		p.logf("正在关闭连接...")
		p.natsConn.Close()
		p.logf("连接已安全关闭")
	}
}

func (p *NyaNATS) Error() error {
	return p.err
}
