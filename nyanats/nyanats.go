package nyanats

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"gopkg.in/yaml.v3"
)

type NATSConfig struct {
	NatsServer     string            `json:"nats_server" yaml:"nats_server"`
	NatsUser       string            `json:"nats_user" yaml:"nats_user"`
	NatsPassword   string            `json:"nats_password" yaml:"nats_password"`
	ClientName     string            `json:"client_name" yaml:"client_name"`
	MaxReconnects  int               `json:"max_reconnects" yaml:"max_reconnects"`
	ReconnectWait  int               `json:"reconnect_wait" yaml:"reconnect_wait"`
	ConnectTimeout int               `json:"connect_timeout" yaml:"connect_timeout"`
	EncryptionKey  string            `json:"encryption_key" yaml:"encryption_key"`
	ThemeKeys      map[string]string `json:"theme_keys" yaml:"theme_keys"`
}

func (c *NATSConfig) setDefaults() {
	if c.NatsServer == "" {
		c.NatsServer = "127.0.0.1:4222"
	}
	if c.ClientName == "" {
		c.ClientName = uuid.NewString()
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
	err        error
	natsConn   *nats.Conn
	debug      *log.Logger
	defaultKey []byte
	themeKeys  map[string][]byte
}

func (p *NyaNATS) logf(format string, v ...interface{}) {
	if p.debug != nil {
		p.debug.Printf("[NyaNATS] "+format, v...)
	}
}

func (p *NyaNATS) getKey(theme string) []byte {
	if key, ok := p.themeKeys[theme]; ok {
		return key
	}
	return p.defaultKey
}

func (p *NyaNATS) encrypt(theme string, plaintext []byte) ([]byte, error) {
	key := p.getKey(theme)
	if len(key) == 0 {
		return plaintext, nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (p *NyaNATS) decrypt(theme string, ciphertext []byte) ([]byte, error) {
	key := p.getKey(theme)
	if len(key) == 0 {
		return ciphertext, nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	l := len(ciphertext)
	if l < nonceSize {
		return nil, fmt.Errorf("E: KEY %d < %d", l, nonceSize)
	}

	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, actualCiphertext, nil)
}

func New(configString string, debug *log.Logger) *NyaNATS {
	var natsConfig NATSConfig
	if err := json.Unmarshal([]byte(configString), &natsConfig); err == nil {
		return NewC(natsConfig, debug)
	}
	if err := yaml.Unmarshal([]byte(configString), &natsConfig); err == nil {
		return NewC(natsConfig, debug)
	}
	return &NyaNATS{err: fmt.Errorf("E: CONF"), debug: debug}
}

func NewC(config NATSConfig, debug *log.Logger) *NyaNATS {
	config.setDefaults()
	p := &NyaNATS{
		debug:     debug,
		themeKeys: make(map[string][]byte),
	}

	if config.EncryptionKey != "" {
		p.defaultKey = []byte(config.EncryptionKey)
		l := len(p.defaultKey)
		if l != 16 && l != 24 && l != 32 {
			p.err = fmt.Errorf("KEYLENERR %d", l)
			p.logf("S# [](%d) ERR: LEN", l)
			return p
		}
		p.logf("S# [](%d)", l)
	} else {
		p.logf("S# [](0)")
	}

	for theme, kStr := range config.ThemeKeys {

		if kStr == "" {
			p.themeKeys[theme] = nil
			p.logf("S# [%s](0)", theme)
			continue
		}

		kByte := []byte(kStr)
		l := len(kByte)
		if l != 16 && l != 24 && l != 32 {
			p.err = fmt.Errorf("S# [%s](%d) ERR: KEYLEN", theme, l)
			return p
		}
		p.themeKeys[theme] = kByte
		p.logf("S# [%s](%d)", theme, l)
	}

	scheme := "nats:/"
	url := fmt.Sprintf("%s/%s", scheme, config.NatsServer)
	if config.NatsUser != "" {
		url = fmt.Sprintf("%s/%s:%s@%s", scheme, config.NatsUser, config.NatsPassword, config.NatsServer)
	}

	opts := []nats.Option{
		nats.Name(config.ClientName),
		nats.MaxReconnects(config.MaxReconnects),
		nats.ReconnectWait(time.Duration(config.ReconnectWait) * time.Second),
		nats.Timeout(time.Duration(config.ConnectTimeout) * time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				p.logf("L- [%s] ERR: %v", p.natsConn.Opts.Name, err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) { p.logf("L+ [%v]", nc.ConnectedUrl()) }),
		nats.ErrorHandler(func(nc *nats.Conn, s *nats.Subscription, err error) { p.logf("L+ [%s] ERR: %v", s.Subject, err) }),
	}

	nc, err := nats.Connect(url, opts...)
	if err != nil {
		p.err = err
		return p
	}

	p.natsConn = nc
	p.logf("L+ [%s]", config.ClientName)
	return p
}

func (p *NyaNATS) Subscribe(theme string, callback func(m string) string) error {
	if p.err != nil {
		return p.err
	}

	_, err := p.natsConn.Subscribe(theme, func(m *nats.Msg) {

		data, err := p.decrypt(m.Subject, m.Data)
		if err != nil {
			p.logf("<- [%s](ERR) %v", m.Subject, err)
			return
		}
		p.logf("<- [%s](%d) %s", theme, len(m.Data), data)

		replyContent := callback(string(data))

		if m.Reply != "" {

			encryptedReply, err := p.encrypt(m.Subject, []byte(replyContent))
			if err != nil {
				p.logf("-> [%s](ERR) %v", m.Reply, err)
				return
			}
			if err := m.Respond(encryptedReply); err != nil {
				p.logf("-> [%s](ERR) %v", m.Reply, err)
			}
		}
	})

	if err != nil {
		p.logf("S+ [%s](ERR) %v", theme, err)
	} else {
		p.logf("S+ [%s]", theme)
	}
	return err
}

func (p *NyaNATS) Publish(theme string, message string) error {
	if p.err != nil {
		return p.err
	}
	data, err := p.encrypt(theme, []byte(message))
	if err != nil {
		return err
	}
	err = p.natsConn.Publish(theme, data)
	if err != nil {
		p.logf("-> [%s](ERR) %v", theme, err)
	} else {
		p.logf("-> [%s](%d) %s", theme, len(data), message)
	}
	return err
}

func (p *NyaNATS) Request(theme string, message string, timeout time.Duration) (string, error) {
	if p.err != nil {
		return "", p.err
	}
	data, err := p.encrypt(theme, []byte(message))
	if err != nil {
		return "", err
	}

	p.logf("-> [%s](%d) %v", theme, len(data), message)
	msg, err := p.natsConn.Request(theme, data, timeout)
	if err != nil {
		p.logf("-> [%s](ERR) %v", theme, err)
		return "", err
	}

	decryptedData, err := p.decrypt(theme, msg.Data)
	if err != nil {
		p.logf("<- [%s](ERR) %v", theme, err)
		return "", err
	}
	p.logf("<- [%s](%d) %s", theme, len(msg.Data), decryptedData)
	return string(decryptedData), nil
}

func (p *NyaNATS) Close() {
	if p.natsConn != nil {
		p.logf("L- [%s]", p.natsConn.Opts.Name)
		p.natsConn.Close()
	}
}

func (p *NyaNATS) Error() error { return p.err }
