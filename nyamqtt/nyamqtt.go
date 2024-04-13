// MQTT 客戶端
package nyamqtt

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/tidwall/gjson"
)

// <類>
var ctx = context.Background()

type NyaMQTT NyaMQTTT
type NyaMQTTT struct {
	SubscribeTopics []string // 已訂閱的主題列表
	AutoReconnect   int      // 是否自動重新連線(毫秒等待時間，-1為禁用)

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

// NyaMQTTStatusHandler: MQTT 連線狀態發生變化時觸發
//
//	`status` int8  連線狀態:
//	-3 重新連線失敗
//	-2 正在重新連線
//	-1 連線丟失
//	 1 已連線
//	 2 將要連線
//	return   error 可能遇到的錯誤
type NyaMQTTStatusHandler func(client string, status int8, err error)

func (p *NyaMQTT) SetNyaMQTTStatusHandler(handler NyaMQTTStatusHandler) {
	p.statusHandler = handler
}

// NyaMQTTSMessageHandler: MQTT 收到新訊息時觸發
//
//	`topic`   string 主題名稱
//	`message` string 收到的訊息文字收到的訊息文字
type NyaMQTTSMessageHandler func(client string, topic string, message string)

func (p *NyaMQTT) SetNyaMQTTSMessageHandler(handler NyaMQTTSMessageHandler) {
	p.messageHandler = handler
}

// </代理方法>

// New: 建立新的 NyaMQTT 例項
//
//		`configJsonString` string 配置 JSON 字串
//		從配置 JSON 檔案中取出的本模組所需的配置段落 JSON 字串
//	 示例配置數值參考 config.template.json
//		本模組所需配置項: mqtt_addr, mqtt_port
//		本模組所需可選配置項: mqtt_client, mqtt_user, mqtt_pwd, mqtt_qos, mqtt_retained
//		`statusHandler`  NyaMQTTStatusHandler   代理方法,見該方法的註釋
//		`messageHandler` NyaMQTTSMessageHandler 代理方法,見該方法的註釋
//	 return *NyaMQTT 新的 NyaMQTT 例項
//		下一步使用 `Error()` 或 `ErrorString()` 檢查是否有錯誤
func New(configJsonString string, statusHandler NyaMQTTStatusHandler, messageHandler NyaMQTTSMessageHandler) *NyaMQTT {
	var configNG string = "NO CONFIG KEY : "
	var configKey string = "mqtt_addr"
	var mqttBroker gjson.Result = gjson.Get(configJsonString, configKey)
	if !mqttBroker.Exists() {
		return &NyaMQTT{err: fmt.Errorf(configNG + configKey)}
	}
	var broker string = mqttBroker.String()
	configKey = "mqtt_port"
	var mqttPort gjson.Result = gjson.Get(configJsonString, configKey)
	if !mqttPort.Exists() {
		return &NyaMQTT{err: fmt.Errorf(configNG + configKey)}
	}
	var port string = mqttPort.String()
	configKey = "mqtt_client"
	var mqttClientID gjson.Result = gjson.Get(configJsonString, configKey)
	var clientid string = ""
	if mqttClientID.Exists() {
		clientid = mqttClientID.String()
	} else {
		clientid = "client" + strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	configKey = "mqtt_user"
	var mqttUsername gjson.Result = gjson.Get(configJsonString, configKey)
	var user string = ""
	if mqttUsername.Exists() {
		user = mqttUsername.String()
	}
	configKey = "mqtt_pwd"
	var mqttPassword gjson.Result = gjson.Get(configJsonString, configKey)
	var password string = ""
	if mqttPassword.Exists() {
		password = mqttPassword.String()
	}
	configKey = "mqtt_qos"
	var mqttQOS gjson.Result = gjson.Get(configJsonString, configKey)
	var qos byte = 0
	if mqttQOS.Exists() {
		qos = IntToBytes(int(mqttQOS.Int()))[0]
	}
	configKey = "mqtt_timeout"
	var mqttTimeout gjson.Result = gjson.Get(configJsonString, configKey)
	var timeout time.Duration = 30
	if mqttTimeout.Exists() {
		timeout = time.Duration(mqttTimeout.Int())
	}
	configKey = "mqtt_retained"
	var mqttRetained gjson.Result = gjson.Get(configJsonString, configKey)
	var retained bool = false
	if mqttRetained.Exists() {
		retained = mqttRetained.Int() != 0
	}
	var certExists [3]bool = [3]bool{false, false, false}
	configKey = "mqtt_cert_clentca"
	var mqttClentCA gjson.Result = gjson.Get(configJsonString, configKey)
	var certClentCA string = ""
	if mqttClentCA.Exists() {
		certClentCA = mqttClentCA.String()
		certExists[0] = true
	}
	configKey = "mqtt_cert_clent"
	var mqttClent gjson.Result = gjson.Get(configJsonString, configKey)
	var certClient string = ""
	if mqttClent.Exists() {
		certClient = mqttClent.String()
		certExists[1] = true
	}
	configKey = "mqtt_cert_clentkey"
	var mqttClentKey gjson.Result = gjson.Get(configJsonString, configKey)
	var certClentkey string = ""
	if mqttClentKey.Exists() {
		certClentkey = mqttClentKey.String()
		certExists[2] = true
	}

	var urlProtocol string = "tcp"
	var opts *mqtt.ClientOptions = mqtt.NewClientOptions()
	opts.SetConnectTimeout(timeout * time.Millisecond)
	if certExists[0] {
		urlProtocol = "ssl"
		if certExists[1] && certExists[2] {
			var tlsconf *tls.Config = tlsConfig(certClentCA, certClient, certClentkey)
			opts.SetTLSConfig(tlsconf)
		} else {
			var tlsconf *tls.Config = tlsConfigWithCA(certClentCA)
			opts.SetTLSConfig(tlsconf)
		}
	}

	var uri string = urlProtocol + "://" + broker + ":" + port
	opts.AddBroker(uri)
	opts.SetClientID(clientid)
	if len(user) > 0 {
		opts.SetUsername(user)
	}
	if len(password) > 0 {
		opts.SetPassword(password)
	}

	var nyamqttobj NyaMQTT = NyaMQTT{statusHandler: statusHandler, messageHandler: messageHandler}
	nyamqttobj.hMessage = func(client mqtt.Client, msg mqtt.Message) {
		var cro mqtt.ClientOptionsReader = client.OptionsReader()
		var message string = string(msg.Payload())
		if nyamqttobj.messageHandler != nil {
			nyamqttobj.messageHandler(servURLstr(cro, msg), msg.Topic(), message)
		}
	}
	opts.SetDefaultPublishHandler(nyamqttobj.hMessage)
	nyamqttobj.hConnect = func(client mqtt.Client) {
		var cro mqtt.ClientOptionsReader = client.OptionsReader()
		if nyamqttobj.statusHandler != nil {
			nyamqttobj.statusHandler(servURLstr(cro, nil), 1, nil)
		}
	}
	opts.OnConnect = nyamqttobj.hConnect

	nyamqttobj.hConnectAttempt = func(broker *url.URL, tlsCfg *tls.Config) *tls.Config {
		if nyamqttobj.statusHandler != nil {
			nyamqttobj.statusHandler("", 2, nil)
		}
		return tlsCfg
	}
	opts.OnConnectAttempt = nyamqttobj.hConnectAttempt

	nyamqttobj.hConnectionLost = func(client mqtt.Client, err error) {
		cro := client.OptionsReader()
		if nyamqttobj.statusHandler != nil {
			nyamqttobj.statusHandler(cro.ClientID(), -1, nil)
			if nyamqttobj.AutoReconnect > -1 {
				nyamqttobj.statusHandler(cro.ClientID(), -2, nil)
				client.Disconnect(uint(nyamqttobj.AutoReconnect))
				if token := client.Connect(); token.Wait() && token.Error() != nil {
					nyamqttobj.statusHandler(cro.ClientID(), -3, token.Error())
				} else {
					nyamqttobj.SubscribeAutoRe()
				}
			}
		}
	}
	opts.OnConnectionLost = nyamqttobj.hConnectionLost

	nyamqttobj.hReconnecting = func(c mqtt.Client, co *mqtt.ClientOptions) {
		cro := c.OptionsReader()
		if nyamqttobj.statusHandler != nil {
			nyamqttobj.statusHandler(cro.ClientID(), -2, nil)
		}
	}
	opts.OnReconnecting = nyamqttobj.hReconnecting

	var client mqtt.Client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return &NyaMQTT{err: token.Error()}
	}
	return &NyaMQTT{db: client, err: nil, defaultQOS: qos, defaultRetained: retained, SubscribeTopics: []string{}}
}

func servURLstr(cro mqtt.ClientOptionsReader, msg mqtt.Message) string {
	var brokers []*url.URL = cro.Servers()
	var infos []string = []string{}
	for i, broker := range brokers {
		if msg == nil {
			infos = append(infos, fmt.Sprintf("[%d]%s", i+1, broker.String()))
		} else {
			infos = append(infos, fmt.Sprintf("[%d]%s(%s)", i+1, broker.String(), msg.Topic()))
		}
	}
	return strings.Join(infos, "; ")
}

// loadConfig: 從載入的配置檔案中載入配置
//
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

// 建立證書配置（自簽證書）
func tlsConfig(caCert string, clientCert string, clientKey string) *tls.Config {
	certpool := x509.NewCertPool()
	ca, err := os.ReadFile(caCert)
	if err != nil {
		fmt.Println("加密证书故障", err.Error())
	}
	certpool.AppendCertsFromPEM(ca)
	// Import client certificate/key pair
	clientKeyPair, err := tls.LoadX509KeyPair(clientCert, clientKey)
	if err != nil {
		fmt.Println("加密证书故障", err.Error())
	}
	return &tls.Config{
		RootCAs:            certpool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{clientKeyPair},
	}
}

// 建立證書配置（CA證書）
func tlsConfigWithCA(caCert string) *tls.Config {
	certpool := x509.NewCertPool()
	ca, err := os.ReadFile(caCert)
	if err != nil {
		fmt.Println("加密证书故障", err.Error())
	}
	certpool.AppendCertsFromPEM(ca)
	return &tls.Config{
		RootCAs: certpool,
	}
}

// IntToBytes: 將 int 轉換為 Bytes
//
//	`n`    int    整數
//	return []byte 位元組陣列
func IntToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

// Error: 獲取上一次操作時可能產生的錯誤
//
//	return error 如果有錯誤，返回錯誤物件，如果沒有錯誤返回 nil
func (p *NyaMQTT) Error() error {
	return p.err
}

// ErrorString: 獲取上一次操作時可能產生的錯誤資訊字串
//
//	return string 如果有錯誤，返回錯誤描述字串，如果沒有錯誤返回空字串
func (p *NyaMQTT) ErrorString() string {
	if p.err == nil {
		return ""
	}
	return p.err.Error()
}

// Subscribe: 訂閱主題
//
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
	if p.err == nil {
		p.SubscribeTopics = append(p.SubscribeTopics, topic)
		return true
	}
	return false
}

// SubscribeMulti: 批次訂閱主題
//
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

// SubscribeAuto: 自動退訂主題
// return  bool[]   主題是否退訂成功
func (p *NyaMQTT) UnsubscribeAuto() []bool {
	var isOK []bool = make([]bool, len(p.SubscribeTopics))
	for i, v := range p.SubscribeTopics {
		isOK[i] = p.Unsubscribe(v)
	}
	return isOK
}

// SubscribeAutoRe: 自動重新訂閱主題
func (p *NyaMQTT) SubscribeAutoRe() []bool {
	var isOK []bool = make([]bool, len(p.SubscribeTopics))
	for i, v := range p.SubscribeTopics {
		isOK[i] = p.Unsubscribe(v)
		isOK[i] = p.Subscribe(v)
	}
	return isOK
}

// Unsubscribe: 退訂主題
//
//	`topic` string 主題名稱
//	return  bool   主題是否退訂成功，如果未成功使用 `Error()` 或 `ErrorString()` 檢查是否有錯誤
func (p *NyaMQTT) Unsubscribe(topic string) bool {
	var token mqtt.Token = p.db.Unsubscribe(topic)
	token.Wait()
	p.err = token.Error()
	return p.err == nil
}

// Unsubscribe: 批次退訂主題
//
//	`topics` string 主題名稱列表
//	return   bool   主題是否退訂成功，如果未成功使用 `Error()` 或 `ErrorString()` 檢查是否有錯誤
func (p *NyaMQTT) UnsubscribeMulti(topics []string) bool {
	var token mqtt.Token = p.db.Unsubscribe(topics...)
	token.Wait()
	p.err = token.Error()
	return p.err == nil
}

// Publish: 傳送訊息
//
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

// Publish: 批次傳送訊息
//
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

// Close: 斷開與 MQTT 伺服器的連線
//
//	`waitTime` uint 斷開等待時間(毫秒)，為了安全斷開建議設定
func (p *NyaMQTT) Close(waitTime uint) {
	p.db.Disconnect(waitTime)
}
