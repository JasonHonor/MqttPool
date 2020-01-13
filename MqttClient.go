package MqttPool

import (
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"os"

	//"runtime"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/gogf/gf/frame/g"
)

type MqttClient struct {
	c mqtt.Client
}

var gMqttClientOnce sync.Once
var gMqttClient *MqttClient

var gGetHostNameOnce sync.Once
var gHostName string

func getHostName(NeedRand bool) string {
	gGetHostNameOnce.Do(func() {
		var err error
		gHostName, err = os.Hostname()
		if err != nil {
			gHostName = ""

			rand.Seed(time.Now().UnixNano())

			var NamePart = make([]string, 8)
			//生成10个0-99之间的随机数
			for i := 0; i < 7; i++ {
				NamePart[i] = fmt.Sprintf("%d", rand.Intn(10))
			}

			gHostName = strings.Join(NamePart, "")

			//TODO: 将主机名写入到配置文件中
		} else {
			if NeedRand {
				rand.Seed(time.Now().UnixNano())

				var NamePart = make([]string, 8)
				//生成10个0-99之间的随机数
				for i := 0; i < 7; i++ {
					NamePart[i] = fmt.Sprintf("%d", rand.Intn(10))
				}

				gHostName += strings.Join(NamePart, "")
			}
		}
	})
	return gHostName
}

//Init once.
func InitMqttClient(clientIds ...string) *MqttClient {
	gMqttClientOnce.Do(func() {

		vServerUrl := g.Cfg().Get("client.mqtt-server")

		if vServerUrl == nil {
			log.Println("No client.mqtt-server defined.")
			return
		}

		if vServerUrl != nil {

			serverUrl := vServerUrl.(string)

			var clientId string
			if len(clientIds) == 0 {
				clientId = getHostName(true)
			} else {
				clientId = clientIds[0]
			}

			gMqttClient = newMqttClient(serverUrl, clientId)

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

	opts.SetAutoReconnect(true)

	opts.SetKeepAlive(60 * time.Second)

	opts.SetMaxReconnectInterval(30 * time.Second)

	//连接丢失处理
	opts.SetConnectionLostHandler(func(mqtt.Client, error) {
		log.Printf("Connection Lost.\n")
	})

	//连接成功处理
	opts.SetOnConnectHandler(func(mqtt.Client) {
		log.Printf("Connection created.\n")
	})

	opts.SetReconnectingHandler(func(mqtt.Client, *mqtt.ClientOptions) {
		log.Printf("Connection recreating.\n")
	})

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
func (mqttClient MqttClient) Connect(nInterval time.Duration) bool {

	if mqttClient.c.IsConnected() {
		//log.Println("mqttClient connected.")
		return true
	}

	var iPos = 0
	for {

		iPos++

		token := mqttClient.c.Connect()
		if token == nil {
			log.Printf("mqttClient token =nil\n")
		} else {
			if token.Wait() && token.Error() == nil {
				return true
			} else {
				log.Printf("MqttClient Connect %d Error:%s\r\n", iPos, token.Error())
				time.Sleep(nInterval)
			}
		}
	}

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

func MqttPublishMsg(mqttContext *MqttContext, nRetryInterval time.Duration) bool {

	var bCompleted bool
	mqttClient := InitMqttClient()

	if mqttClient.Connect(nRetryInterval) {
		mqttClient.Publish(mqttContext.Topic, &mqttContext.Message)
		log.Println("msg published.")
		bCompleted = true
	} else {
		bCompleted = false
	}

	return bCompleted
}

func MqttPublishMsgRet(mqttContext *MqttContext, nRetryInterval time.Duration) bool {

	var bCompleted bool

	mqttClient := InitMqttClient()

	if mqttClient.Connect(nRetryInterval) {
		mqttClient.Publish(mqttContext.Topic, &mqttContext.Message)
		bCompleted = true
	} else {

		bCompleted = false
	}

	return bCompleted
}

func MqttSubscribeTopic(nRetryInterval time.Duration,
	topic string, callback mqtt.MessageHandler) (bool, bool) {

	var bConnected, bCompleted bool

	mqttClient := InitMqttClient()

	if mqttClient.Connect(nRetryInterval) {

		bConnected = true

		mqttClient.Subscribe(topic, callback)
		log.Println("Topic " + topic + " subscribed.")

		bCompleted = true

	} else {

		bConnected = false
		bCompleted = false

	}

	return bConnected, bCompleted
}
