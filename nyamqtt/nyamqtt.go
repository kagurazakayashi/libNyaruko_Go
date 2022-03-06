package nyamqtt

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"net/url"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	cmap "github.com/orcaman/concurrent-map"
)

// <類>
var ctx = context.Background()

type NyaMQTT NyaMQTTT
type NyaMQTTT struct {
	db              mqtt.Client
	err             error
	statusHandler   NyaMQTTStatusHandler
	messageHandler  NyaMQTTSMessageHandler
	hConnect        mqtt.OnConnectHandler
	hConnectionLost mqtt.ConnectionLostHandler
	hReconnecting   mqtt.ReconnectHandler
	hConnectAttempt mqtt.ConnectionAttemptHandler
	hMessage        mqtt.MessageHandler
	defaultQOS      byte
	defaultRetained bool
	qos             byte
	retained        bool
}

// </類>

// <可選配置>
type Option struct {
	qos         byte
	retained    bool // 表示mqtt伺服器要保留這次推送的資訊，如果有新的訂閱者出現，就會把這訊息推送給它（持久化推送）
	isErrorStop bool // 在批次操作中是否遇到錯誤就停止
}
type OptionConfig func(*Option)

func Option_qos(v byte) OptionConfig {
	return func(p *Option) {
		p.qos = v
	}
}
func Option_retained(v bool) OptionConfig {
	return func(p *Option) {
		p.retained = v
	}
}
func Option_isErrorStop(v bool) OptionConfig {
	return func(p *Option) {
		p.isErrorStop = v
	}
}

// </可選配置>

type NyaMQTTStatusHandler func(status int8, err error)

func (p NyaMQTT) SetNyaMQTTStatusHandler(handler NyaMQTTStatusHandler) {
	p.statusHandler = handler
}

type NyaMQTTSMessageHandler func(messageID uint16, topic string, message string)

func (p NyaMQTT) SetNyaMQTTSMessageHandler(handler NyaMQTTSMessageHandler) {
	p.messageHandler = handler
}

func New(confCMap cmap.ConcurrentMap, statusHandler NyaMQTTStatusHandler, messageHandler NyaMQTTSMessageHandler) NyaMQTT {
	redisBroker, err := loadConfig(confCMap, "mqtt_addr")
	if err != nil {
		return NyaMQTT{err: err}
	}
	redisPort, err := loadConfig(confCMap, "mqtt_port")
	if err != nil {
		return NyaMQTT{err: err}
	}
	var opts *mqtt.ClientOptions = mqtt.NewClientOptions()
	var uri string = "tcp://" + redisBroker + ":" + redisPort
	opts.AddBroker(uri)
	redisClientID, err := loadConfig(confCMap, "mqtt_client")
	if err != nil || len(redisClientID) == 0 {
		redisClientID = "client" + strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	opts.SetClientID(redisClientID)
	redisUsername, err := loadConfig(confCMap, "mqtt_user")
	if err == nil && len(redisUsername) > 0 {
		opts.SetUsername(redisUsername)
	}
	redisPassword, err := loadConfig(confCMap, "mqtt_pwd")
	if err == nil && len(redisPassword) > 0 {
		opts.SetPassword(redisPassword)
	}
	var bQOS byte = 0
	qos, err := loadConfig(confCMap, "mqtt_qos")
	if err == nil && len(qos) > 0 {
		// bQOS = []byte(qos)[0]
		qosi, err := strconv.Atoi(qos)
		if err == nil && qosi >= 0 && qosi <= 2 {
			bQOS = IntToBytes(qosi)[0]
		}
	}
	var bRetained bool = false
	retained, err := loadConfig(confCMap, "mqtt_retained")
	if err == nil && len(retained) > 0 {
		bRetained = retained != "0"
	}

	var nyamqttobj NyaMQTT = NyaMQTT{statusHandler: statusHandler, messageHandler: messageHandler}
	nyamqttobj.hMessage = func(client mqtt.Client, msg mqtt.Message) {
		var message string = string(msg.Payload())
		nyamqttobj.messageHandler(msg.MessageID(), msg.Topic(), message)
	}
	opts.SetDefaultPublishHandler(nyamqttobj.hMessage)
	nyamqttobj.hConnect = func(client mqtt.Client) {
		nyamqttobj.statusHandler(1, nil)
	}
	opts.OnConnect = nyamqttobj.hConnect

	nyamqttobj.hConnectAttempt = func(broker *url.URL, tlsCfg *tls.Config) *tls.Config {
		nyamqttobj.statusHandler(2, nil)
		return tlsCfg
	}
	opts.OnConnectAttempt = nyamqttobj.hConnectAttempt

	nyamqttobj.hConnectionLost = func(client mqtt.Client, err error) {
		nyamqttobj.statusHandler(-1, nil)
	}
	opts.OnConnectionLost = nyamqttobj.hConnectionLost

	nyamqttobj.hReconnecting = func(c mqtt.Client, co *mqtt.ClientOptions) {
		nyamqttobj.statusHandler(-2, nil)
	}
	opts.OnReconnecting = nyamqttobj.hReconnecting

	var client mqtt.Client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return NyaMQTT{err: token.Error()}
	}
	return NyaMQTT{db: client, err: nil, defaultQOS: bQOS, defaultRetained: bRetained}
}

func loadConfig(confCMap cmap.ConcurrentMap, key string) (string, error) {
	val, isExist := confCMap.Get(key)
	if !isExist {
		return "", fmt.Errorf("no config : " + key)
	}
	return val.(string), nil
}

func IntToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

//Error: 獲取上一次操作時可能產生的錯誤
//	return error 如果有錯誤，返回錯誤物件，如果沒有錯誤返回 nil
func (p NyaMQTT) Error() error {
	return p.err
}

//ErrorString: 獲取上一次操作時可能產生的錯誤資訊字串
//	return string 如果有錯誤，返回錯誤描述字串，如果沒有錯誤返回空字串
func (p NyaMQTT) ErrorString() string {
	if p.err == nil {
		return ""
	}
	return p.err.Error()
}

func (p NyaMQTT) Subscribe(topic string, options ...OptionConfig) bool {
	option := &Option{qos: p.defaultQOS}
	for _, o := range options {
		o(option)
	}
	fmt.Println("Subscribe", topic, option.qos)
	token := p.db.Subscribe(topic, option.qos, nil)
	token.Wait()
	p.err = token.Error()
	return p.err == nil
}

func (p NyaMQTT) SubscribeMulti(topics []string, options ...OptionConfig) bool {
	option := &Option{qos: p.defaultQOS, isErrorStop: false}
	for _, o := range options {
		o(option)
	}
	var isOK = true
	for _, v := range topics {
		if !p.Subscribe(v, options...) {
			isOK = false
			if option.isErrorStop {
				return false
			}
		}
	}
	return isOK
}

func (p NyaMQTT) Unsubscribe(topic string) bool {
	var token mqtt.Token = p.db.Unsubscribe(topic)
	token.Wait()
	p.err = token.Error()
	return p.err == nil
}

func (p NyaMQTT) UnsubscribeMulti(topics []string) bool {
	var token mqtt.Token = p.db.Unsubscribe(topics...)
	token.Wait()
	p.err = token.Error()
	return p.err == nil
}

func (p NyaMQTT) Publish(topic string, text string, options ...OptionConfig) bool {
	option := &Option{qos: p.defaultQOS, retained: p.defaultRetained}
	for _, o := range options {
		o(option)
	}
	fmt.Println("Publish", topic, option.qos, option.retained, text)
	var token mqtt.Token = p.db.Publish(topic, option.qos, option.retained, text)
	token.Wait()
	p.err = token.Error()
	return p.err == nil
}

func (p NyaMQTT) PublishMulti(topicAndTexts map[string]string, options ...OptionConfig) bool {
	option := &Option{qos: p.defaultQOS, retained: p.defaultRetained, isErrorStop: false}
	for _, o := range options {
		o(option)
	}
	var isOK = true
	for key, value := range topicAndTexts {
		if !p.Publish(key, value, options...) {
			isOK = false
			if option.isErrorStop {
				return false
			}
		}
	}
	return isOK
}

func (p NyaMQTT) Close(waitTime uint) {
	p.db.Disconnect(waitTime)
}
