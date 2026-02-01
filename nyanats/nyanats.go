package nyanats

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid" // 需要运行 go get github.com/google/uuid
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

// setDefaults 填充默认值
func (c *NATSConfig) setDefaults() {
	if c.NatsServer == "" {
		c.NatsServer = "127.0.0.1:4222"
	}

	// 如果未指定名称，生成一个类似 "NyaNATS-550e8400-e29b..." 的唯一标识
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

func New(configString string, debug *log.Logger) *NyaNATS {
	var natsConfig NATSConfig

	if err := json.Unmarshal([]byte(configString), &natsConfig); err == nil {
		return NewC(natsConfig, debug)
	}

	if err := yaml.Unmarshal([]byte(configString), &natsConfig); err == nil {
		return NewC(natsConfig, debug)
	}
	return &NyaNATS{err: fmt.Errorf("failed to parse config"), debug: debug}
}

func NewC(config NATSConfig, debug *log.Logger) *NyaNATS {
	config.setDefaults()

	var url string
	if config.NatsUser != "" {

		url = fmt.Sprintf("nats://%s:%s@%s", config.NatsUser, config.NatsPassword, config.NatsServer)
	} else {

		url = fmt.Sprintf("nats://%s", config.NatsServer)
	}

	opts := []nats.Option{
		nats.Name(config.ClientName),
		nats.MaxReconnects(config.MaxReconnects),
		nats.ReconnectWait(time.Duration(config.ReconnectWait) * time.Second),
		nats.Timeout(time.Duration(config.ConnectTimeout) * time.Second),

		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if debug != nil {
				debug.Printf("NyaNATS: 链接断开 - %v", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			if debug != nil {
				debug.Printf("NyaNATS: 已重新连接至 %v", nc.ConnectedUrl())
			}
		}),
	}

	nc, err := nats.Connect(url, opts...)
	if err != nil {
		return &NyaNATS{err: err, debug: debug}
	}

	return &NyaNATS{
		err:      nil,
		natsConn: nc,
		debug:    debug,
	}
}

func (p *NyaNATS) Subscribe(theme string, callback func(m string) string) error {
	if p.err != nil {
		return p.err
	}

	_, err := p.natsConn.Subscribe(theme, func(m *nats.Msg) {
		if p.debug != nil {
			p.debug.Printf("Received message on [%s]", theme)
		}

		replyContent := callback(string(m.Data))

		if m.Reply != "" {
			if err := m.Respond([]byte(replyContent)); err != nil {
				if p.debug != nil {
					p.debug.Printf("Respond error: %v", err)
				}
			}
		}
	})

	return err
}

func (p *NyaNATS) Publish(theme string, message string) error {
	if p.err != nil {
		return p.err
	}
	return p.natsConn.Publish(theme, []byte(message))
}

func (p *NyaNATS) Request(theme string, message string, timeout time.Duration) (string, error) {
	if p.err != nil {
		return "", p.err
	}

	msg, err := p.natsConn.Request(theme, []byte(message), timeout)
	if err != nil {
		return "", err
	}
	return string(msg.Data), nil
}

func (p *NyaNATS) Close() {
	if p.natsConn != nil {
		p.natsConn.Close()
	}
}

func (p *NyaNATS) Error() error {
	return p.err
}
