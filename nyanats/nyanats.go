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
	// EncryptionKey: 16, 24, 或 32 位字符串分别对应 AES-128, 192, 256
	EncryptionKey string `json:"encryption_key" yaml:"encryption_key"`
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
	err      error
	natsConn *nats.Conn
	debug    *log.Logger
	key      []byte // 内部存储字节类型的密钥
}

func (p *NyaNATS) logf(format string, v ...interface{}) {
	if p.debug != nil {
		p.debug.Printf("[NyaNATS] "+format, v...)
	}
}

// encrypt 使用 AES-GCM 加密数据
func (p *NyaNATS) encrypt(plaintext []byte) ([]byte, error) {
	if len(p.key) == 0 {
		return plaintext, nil
	}

	block, err := aes.NewCipher(p.key)
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

	// 封口：将加密结果追加在 nonce 后面
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// decrypt 使用 AES-GCM 解密数据
func (p *NyaNATS) decrypt(ciphertext []byte) ([]byte, error) {
	if len(p.key) == 0 {
		return ciphertext, nil
	}

	block, err := aes.NewCipher(p.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
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
	p := &NyaNATS{debug: debug}

	if config.EncryptionKey != "" {
		p.key = []byte(config.EncryptionKey)
		// 校验 AES 密钥长度
		if l := len(p.key); l != 16 && l != 24 && l != 32 {
			p.err = fmt.Errorf("E: AES KEY LEN (%d), MUST BE 16, 24, OR 32", l)
			return p
		}
		p.logf("MODE: ENCRYPTED")
	}

	var url string
	if config.NatsUser != "" {
		url = fmt.Sprintf("nats:/%s:%s@%s", config.NatsUser, config.NatsPassword, config.NatsServer)
		p.logf("CONN %s@%s", config.NatsUser, config.NatsServer)
	} else {
		url = fmt.Sprintf("nats:/%s", config.NatsServer)
		p.logf("CONN %s", config.NatsServer)
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
		p.logf("E CON: %v", err)
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

		// 解密输入
		data, err := p.decrypt(m.Data)
		if err != nil {
			p.logf("E DECRYPT: %v", err)
			return
		}

		replyContent := callback(string(data))

		if m.Reply != "" {
			// 加密响应
			encryptedReply, err := p.encrypt([]byte(replyContent))
			if err != nil {
				p.logf("E ENC-RESP: %v", err)
				return
			}

			p.logf("-> %s , LEN: %d", m.Reply, len(encryptedReply))
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

	data, err := p.encrypt([]byte(message))
	if err != nil {
		return err
	}

	err = p.natsConn.Publish(theme, data)
	if err != nil {
		p.logf("E PUB: %s : %v", theme, err)
	} else {
		p.logf("-> %s (ENC:%v)", theme, len(p.key) > 0)
	}
	return err
}

func (p *NyaNATS) Request(theme string, message string, timeout time.Duration) (string, error) {
	if p.err != nil {
		return "", p.err
	}

	data, err := p.encrypt([]byte(message))
	if err != nil {
		return "", err
	}

	p.logf("-> %s", theme)
	msg, err := p.natsConn.Request(theme, data, timeout)
	if err != nil {
		p.logf("E REQ: %s : %v", theme, err)
		return "", err
	}

	// 解密返回的消息
	decryptedData, err := p.decrypt(msg.Data)
	if err != nil {
		p.logf("E DECRYPT-RES: %v", err)
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

func (p *NyaNATS) Error() error {
	return p.err
}
