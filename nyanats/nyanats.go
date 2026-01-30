package nyanats

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"gopkg.in/yaml.v3"
)

type NATSConfig struct {
	NatsServer   string `json:"nats_server" yaml:"nats_server"`
	NatsUser     string `json:"nats_user" yaml:"nats_user"`
	NatsPassword string `json:"nats_password" yaml:"nats_password"`
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

func NewC(natsConfig NATSConfig, debug *log.Logger) *NyaNATS {

	url := fmt.Sprintf("%s/%s:%s@%s", "nats:/", natsConfig.NatsUser, natsConfig.NatsPassword, natsConfig.NatsServer)

	opts := []nats.Option{
		nats.Name("NyaNATS_Client"),
		nats.MaxReconnects(5),
		nats.ReconnectWait(2 * time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if debug != nil {
				debug.Printf("NATS Disconnected: %v", err)
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
