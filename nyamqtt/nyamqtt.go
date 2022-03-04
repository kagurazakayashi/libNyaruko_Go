package nyamqtt

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	cmap "github.com/orcaman/concurrent-map"
)

// <类>
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
}

// </类>

type NyaMQTTStatusHandler func(int8, error)

func (p NyaMQTT) SetNyaMQTTStatusHandler(handler NyaMQTTStatusHandler) {
	p.statusHandler = handler
}

type NyaMQTTSMessageHandler func(uint16, string, string)

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
	return NyaMQTT{db: client, err: nil}
}

func loadConfig(confCMap cmap.ConcurrentMap, key string) (string, error) {
	val, isExist := confCMap.Get(key)
	if !isExist {
		return "", fmt.Errorf("no config : " + key)
	}
	return val.(string), nil
}

func (p NyaMQTT) Error() {

}

// var messagePubHandler mqtt.MessageHandler =

// var connectHandler mqtt.OnConnectHandler =

// var connectLostHandler mqtt.ConnectionLostHandler =

// var reconnectHandler mqtt.ReconnectHandler =

// var connectAttemptHandler mqtt.ConnectionAttemptHandler =
