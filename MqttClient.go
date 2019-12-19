package MqttPool

import (
	"crypto/tls"
	"log"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/gogf/gf/frame/g"
)

type MqttClient struct {
	c      mqtt.Client
	rwlock sync.RWMutex //lock for client
}

var gMqttClientOnce sync.Once
var gMqttClient *MqttClient

//Init once.
func InitMqttClient() *MqttClient {
	gMqttClientOnce.Do(func() {

		vServerUrl := g.Cfg().Get("client.mqtt-server")

		if vServerUrl == nil {
			log.Println("No client.mqtt_server defined.")
			return
		}

		if vServerUrl != nil {

			serverUrl := vServerUrl.(string)

			gMqttClient = newMqttClient(serverUrl, "txt1234")

			//gMqttClient.Connect(3, 1)
		}
	})
	return gMqttClient
}

func newMqttClient(serverUri, clientId string) *MqttClient {

	client := new(MqttClient)

	// set the protocol, ip and port of the broker. tcp://localhost:1883,ssl://127.0.0.1:883
	opts := mqtt.NewClientOptions().AddBroker(serverUri)

	// set the id to the client.
	opts.SetClientID(clientId)

	//support ssl protocol.
	if strings.HasPrefix(serverUri, "ssl://") {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}

		opts.SetTLSConfig(tlsConfig)
	}

	c := mqtt.NewClient(opts)

	// create a new client
	client.c = c

	return client
}

//try connect with failed retry interval.
func (mqttClient MqttClient) Connect(nRetry int, nInterval time.Duration) bool {

	log.Println("mqttClient Connecting")

	mqttClient.rwlock.RLock()

	if mqttClient.c.IsConnected() {
		log.Println("mqttClient connected.")
		return true
	} else {
		log.Println("mqttClient disconnected.")
	}

	mqttClient.rwlock.RUnlock()

	//we are going to try connecting for max 10 times to the server if the connection fails.
	mqttClient.rwlock.Lock()

	for i := 0; i < nRetry; i++ {

		token := mqttClient.c.Connect()
		if token == nil {
			log.Printf("mqttClient token =nil\n")
		} else {
			if token.Wait() && token.Error() == nil {
				return true
			} else {
				log.Printf("MqttClient Connect Error:%s\r\n", token.Error())
				time.Sleep(nInterval)
			}
		}
	}
	mqttClient.rwlock.Unlock()
	return false
}

func (mqttClient MqttClient) Disconnect() {
	log.Printf("mqttClient.Disconnect\n")
	mqttClient.c.Disconnect(250)
}

//publish message
func (mqttClient MqttClient) Publish(topic string, msg *string) bool {
	//topic string, qos byte, retained bool, payload interface{}
	//QoS 0(At most once)」、「QoS 1(At least once)」、「QoS 2(Exactly once」
	token := mqttClient.c.Publish(topic, 1, true, *msg)

	if token.Wait() && token.Error() == nil {
		return true
	} else {
		log.Printf("MqttClient Publish Error:%s\r\n", token.Error())
	}
	return false
}

//subscribe topic
func (mqttClient MqttClient) Subscribe(topic string, callback mqtt.MessageHandler) bool {
	//topic string, qos byte, retained bool, payload interface{}
	//QoS 0(At most once)」、「QoS 1(At least once)」、「QoS 2(Exactly once」
	token := mqttClient.c.Subscribe(topic, 1, callback)

	if token.Wait() && token.Error() == nil {
		return true
	} else {
		log.Printf("MqttClient Subscribe Error:%s\r\n", token.Error())
	}
	return false
}

func MqttPublishMsg(mqttContext *MqttContext, nReconnectLimit int, nRetryInterval time.Duration) {

	mqttClient := InitMqttClient()

	if mqttClient.Connect(nReconnectLimit, nRetryInterval) {
		mqttClient.Publish(mqttContext.Topic, &mqttContext.Message)

		log.Println("msg published.")
	}
}
