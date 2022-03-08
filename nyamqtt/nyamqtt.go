package nyamqtt

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
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
	qos         byte // 0只管發即可 1至少一次 2保證相同的訊息只接收一條
	retained    bool // 持久化推送
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

// <代理方法>

//NyaMQTTStatusHandler: MQTT 連線狀態發生變化時觸發
//	`status` int8  連線狀態:
//	-2 正在重新連線
//	-1 連線丟失
//	 1 已連線
//	 2 將要連線
//	return   error 可能遇到的錯誤
type NyaMQTTStatusHandler func(status int8, err error)

func (p *NyaMQTT) SetNyaMQTTStatusHandler(handler NyaMQTTStatusHandler) {
	p.statusHandler = handler
}

//NyaMQTTSMessageHandler: MQTT 收到新訊息時觸發
//	`topic`   string 主題名稱
//	`message` string 收到的訊息文字收到的訊息文字
type NyaMQTTSMessageHandler func(topic string, message string)

func (p *NyaMQTT) SetNyaMQTTSMessageHandler(handler NyaMQTTSMessageHandler) {
	p.messageHandler = handler
}

// </代理方法>

//New: 建立新的 NyaMQTT 例項
//	`configJsonString` string 配置 JSON 字串
//	從配置 JSON 檔案中取出的本模組所需的配置段落 JSON 字串
//  示例配置數值參考 config.template.json
//	本模組所需配置項: mqtt_addr, mqtt_port
//	本模組所需可選配置項: mqtt_client, mqtt_user, mqtt_pwd, mqtt_qos, mqtt_retained
//	`statusHandler`  NyaMQTTStatusHandler   代理方法,見該方法的註釋
//	`messageHandler` NyaMQTTSMessageHandler 代理方法,見該方法的註釋
//  return *NyaMQTT 新的 NyaMQTT 例項
//	下一步使用 `Error()` 或 `ErrorString()` 檢查是否有錯誤
func New(configJsonString string, statusHandler NyaMQTTStatusHandler, messageHandler NyaMQTTSMessageHandler) *NyaMQTT {
	redisBroker, err := loadConfig(confCMap, "mqtt_addr")
	if err != nil {
		return &NyaMQTT{err: err}
	}
	redisPort, err := loadConfig(confCMap, "mqtt_port")
	if err != nil {
		return &NyaMQTT{err: err}
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
		if nyamqttobj.statusHandler != nil {
			nyamqttobj.messageHandler(msg.Topic(), message)
		}
	}
	opts.SetDefaultPublishHandler(nyamqttobj.hMessage)
	nyamqttobj.hConnect = func(client mqtt.Client) {
		if nyamqttobj.statusHandler != nil {
			nyamqttobj.statusHandler(1, nil)
		}
	}
	opts.OnConnect = nyamqttobj.hConnect

	nyamqttobj.hConnectAttempt = func(broker *url.URL, tlsCfg *tls.Config) *tls.Config {
		if nyamqttobj.statusHandler != nil {
			nyamqttobj.statusHandler(2, nil)
		}
		return tlsCfg
	}
	opts.OnConnectAttempt = nyamqttobj.hConnectAttempt

	nyamqttobj.hConnectionLost = func(client mqtt.Client, err error) {
		if nyamqttobj.statusHandler != nil {
			nyamqttobj.statusHandler(-1, nil)
		}
	}
	opts.OnConnectionLost = nyamqttobj.hConnectionLost

	nyamqttobj.hReconnecting = func(c mqtt.Client, co *mqtt.ClientOptions) {
		if nyamqttobj.statusHandler != nil {
			nyamqttobj.statusHandler(-2, nil)
		}
	}
	opts.OnReconnecting = nyamqttobj.hReconnecting

	var client mqtt.Client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return &NyaMQTT{err: token.Error()}
	}
	return &NyaMQTT{db: client, err: nil, defaultQOS: bQOS, defaultRetained: bRetained}
}

//loadConfig: 從載入的配置檔案中載入配置
//	`confCMap` cmap.ConcurrentMap 載入的配置檔案字典
//	`key`      string             配置名稱
//	return     string             配置內容
//	return     error              可能遇到的錯誤
func loadConfig(confCMap cmap.ConcurrentMap, key string) (string, error) {
	val, isExist := confCMap.Get(key)
	if !isExist {
		return "", fmt.Errorf("no config : " + key)
	}
	return val.(string), nil
}

//IntToBytes: 將 int 轉換為 Bytes
//	`n`    int    整數
//	return []byte 位元組陣列
func IntToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

//Error: 獲取上一次操作時可能產生的錯誤
//	return error 如果有錯誤，返回錯誤物件，如果沒有錯誤返回 nil
func (p *NyaMQTT) Error() error {
	return p.err
}

//ErrorString: 獲取上一次操作時可能產生的錯誤資訊字串
//	return string 如果有錯誤，返回錯誤描述字串，如果沒有錯誤返回空字串
func (p *NyaMQTT) ErrorString() string {
	if p.err == nil {
		return ""
	}
	return p.err.Error()
}

//Subscribe: 訂閱主題
//	`topic`   string       主題名稱
//	`options` ...OptionConfig 可選配置，執行 `Option_*` 函式輸入
//		`qos` byte 0只管發即可 1至少一次 2保證相同的訊息只接收一條
//	此處支援的可選配置: qos
//	`bool` 訂閱是否成功，如果未成功使用 `Error()` 或 `ErrorString()` 檢查是否有錯誤
func (p *NyaMQTT) Subscribe(topic string, options ...OptionConfig) bool {
	option := &Option{qos: p.defaultQOS}
	for _, o := range options {
		o(option)
	}
	// fmt.Println("Subscribe", topic, option.qos)
	token := p.db.Subscribe(topic, option.qos, nil)
	token.Wait()
	p.err = token.Error()
	return p.err == nil
}

//SubscribeMulti: 批次訂閱主題
//	`topics`  string       主題名稱列表
//	`options` ...OptionConfig 可選配置，執行 `Option_*` 函式輸入
//		`qos` byte 0只管發即可 1至少一次 2保證相同的訊息只接收一條
//		`isErrorStop` bool 在批次操作中是否遇到錯誤就停止
//	return    bool         訂閱是否成功，如果未成功使用 `Error()` 或 `ErrorString()` 檢查是否有錯誤
func (p *NyaMQTT) SubscribeMulti(topics []string, options ...OptionConfig) bool {
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

//Unsubscribe: 退訂主題
//	`topic` string 主題名稱
//	return  bool   主題是否退訂成功，如果未成功使用 `Error()` 或 `ErrorString()` 檢查是否有錯誤
func (p *NyaMQTT) Unsubscribe(topic string) bool {
	var token mqtt.Token = p.db.Unsubscribe(topic)
	token.Wait()
	p.err = token.Error()
	return p.err == nil
}

//Unsubscribe: 批次退訂主題
//	`topics` string 主題名稱列表
//	return   bool   主題是否退訂成功，如果未成功使用 `Error()` 或 `ErrorString()` 檢查是否有錯誤
func (p *NyaMQTT) UnsubscribeMulti(topics []string) bool {
	var token mqtt.Token = p.db.Unsubscribe(topics...)
	token.Wait()
	p.err = token.Error()
	return p.err == nil
}

//Publish: 傳送訊息
//	`topic`   string 傳送到哪個主題
//	`text`    string 要傳送的文字內容
//	`options` ...OptionConfig 可選配置，執行 `Option_*` 函式輸入
//		`qos` byte 0只管發即可 1至少一次 2保證相同的訊息只接收一條
//		`retained` bool 表示 MQTT 伺服器要保留這次推送的資訊，如果有新的訂閱者出現，就會把這訊息推送給它（持久化推送）
//	此處支援的可選配置: qos, retained
//	return    bool   是否傳送成功，如果未成功使用 `Error()` 或 `ErrorString()` 檢查是否有錯誤
func (p *NyaMQTT) Publish(topic string, text string, options ...OptionConfig) bool {
	option := &Option{qos: p.defaultQOS, retained: p.defaultRetained}
	for _, o := range options {
		o(option)
	}
	// fmt.Println("Publish", topic, option.qos, option.retained, text)
	var token mqtt.Token = p.db.Publish(topic, option.qos, option.retained, text)
	token.Wait()
	p.err = token.Error()
	return p.err == nil
}

//Publish: 批次傳送訊息
//	`topicAndTexts` map[string]string 主題名稱:要傳送的文字內容 字典
//	`options` OptionConfig 可選配置,見上方該配置項的註釋
//	此處支援的可選配置: qos, retained
//	return bool 是否傳送成功，如果未成功使用 `Error()` 或 `ErrorString()` 檢查是否有錯誤
func (p *NyaMQTT) PublishMulti(topicAndTexts map[string]string, options ...OptionConfig) bool {
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

//Close: 斷開與 MQTT 伺服器的連線
//	`waitTime` uint 斷開等待時間(毫秒)，為了安全斷開建議設定
func (p *NyaMQTT) Close(waitTime uint) {
	p.db.Disconnect(waitTime)
}

//Open: 根據之前讀入的設定重新開啟連線，在 Close() 之後執行。
func (p *NyaMQTT) Open() error {
	if p.db == nil {
		sqlLiteDB, err := sql.Open(p.dbver, p.dbfile)
		p.err = err
		if err != nil {
			return err
		}
		p.db = sqlLiteDB
	}
	return nil
}
