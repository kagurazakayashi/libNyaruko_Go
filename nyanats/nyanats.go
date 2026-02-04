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
	NatsServer     string `json:"nats_server" yaml:"nats_server"`
	NatsUser       string `json:"nats_user" yaml:"nats_user"`
	NatsPassword   string `json:"nats_password" yaml:"nats_password"`
	ClientName     string `json:"client_name" yaml:"client_name"`
	MaxReconnects  int    `json:"max_reconnects" yaml:"max_reconnects"`
	ReconnectWait  int    `json:"reconnect_wait" yaml:"reconnect_wait"`
	ConnectTimeout int    `json:"connect_timeout" yaml:"connect_timeout"`

	EncryptionKey string `json:"encryption_key" yaml:"encryption_key"`

	ThemeKeys map[string]string `json:"theme_keys" yaml:"theme_keys"`
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
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("E: CIPHER SHORT")
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
		if l := len(p.defaultKey); l != 16 && l != 24 && l != 32 {
			p.err = fmt.Errorf("E: DEFAULT KEY LEN %d", l)
			return p
		}
	}

	for theme, kStr := range config.ThemeKeys {
		kByte := []byte(kStr)
		if l := len(kByte); l != 16 && l != 24 && l != 32 {
			p.err = fmt.Errorf("E: THEME [%s] KEY LEN %d", theme, l)
			return p
		}
		p.themeKeys[theme] = kByte
		p.logf("THEME KEY LOADED: %s", theme)
	}

	url := fmt.Sprintf("nats://%s", config.NatsServer)
	if config.NatsUser != "" {
		url = fmt.Sprintf("nats://%s:%s@%s", config.NatsUser, config.NatsPassword, config.NatsServer)
	}

	opts := []nats.Option{
		nats.Name(config.ClientName),
		nats.MaxReconnects(config.MaxReconnects),
		nats.ReconnectWait(time.Duration(config.ReconnectWait) * time.Second),
		nats.Timeout(time.Duration(config.ConnectTimeout) * time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) { p.logf("UNLINK %v", err) }),
		nats.ReconnectHandler(func(nc *nats.Conn) { p.logf("LINK %v", nc.ConnectedUrl()) }),
		nats.ErrorHandler(func(nc *nats.Conn, s *nats.Subscription, err error) { p.logf("E LINK: %s : %v", s.Subject, err) }),
	}

	nc, err := nats.Connect(url, opts...)
	if err != nil {
		p.err = err
		return p
	}

	p.natsConn = nc
	p.logf("LINKID: %s", config.ClientName)
	return p
}

func (p *NyaNATS) Subscribe(theme string, callback func(m string) string) error {
	if p.err != nil {
		return p.err
	}

	_, err := p.natsConn.Subscribe(theme, func(m *nats.Msg) {
		p.logf("<- %s , LEN: %d", theme, len(m.Data))

		data, err := p.decrypt(m.Subject, m.Data)
		if err != nil {
			p.logf("E DECRYPT [%s]: %v", m.Subject, err)
			return
		}

		replyContent := callback(string(data))

		if m.Reply != "" {

			encryptedReply, err := p.encrypt(m.Subject, []byte(replyContent))
			if err != nil {
				p.logf("E ENC-RESP: %v", err)
				return
			}
			if err := m.Respond(encryptedReply); err != nil {
				p.logf("E SND: %s : %v", m.Reply, err)
			}
		}
	})

	if err != nil {
		p.logf("E SUB : %s : %v", theme, err)
	} else {
		p.logf("SUB: %s", theme)
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
		p.logf("E PUB: %s : %v", theme, err)
	} else {
		p.logf("-> %s", theme)
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

	p.logf("-> %s", theme)
	msg, err := p.natsConn.Request(theme, data, timeout)
	if err != nil {
		p.logf("E REQ: %s : %v", theme, err)
		return "", err
	}

	decryptedData, err := p.decrypt(theme, msg.Data)
	if err != nil {
		p.logf("E DECRYPT-RES [%s]: %v", theme, err)
		return "", err
	}
	p.logf("<- %s", theme)
	return string(decryptedData), nil
}

func (p *NyaNATS) Close() {
	if p.natsConn != nil {
		p.logf("OFF")
		p.natsConn.Close()
	}
}

func (p *NyaNATS) Error() error { return p.err }
