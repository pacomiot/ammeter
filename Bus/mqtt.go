package bus

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/ammeter/config"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/yuguorong/go/log"
)

type QMqtt struct {
	baseBus
	User    string      `json:"accessToken"`
	MqttUrl string      `json:"mqttserver"`
	Uid     string      `json:"id"`
	Name    string      `json:"name"`
	Pswd    string      `json:-`
	Qos     byte        `json:-`
	cnn     MQTT.Client `json:-`
	cliLock *sync.Mutex `json:-`
}

type CbRegistMqttSubs func(cnn *QMqtt) int

var cbMqttEdgeConfig MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	log.Infof("R_TOPIC: %s\n", msg.Topic())
	log.Infof("R_MSG: %s\n", msg.Payload())
}

func (mqtt *QMqtt) subTopic(subtopic string, Qos byte, hdlr MQTT.MessageHandler) {
	log.Info("MQTT sub :[" + subtopic + "]")
	if token := mqtt.cnn.Subscribe(subtopic, Qos, hdlr); token.Wait() && token.Error() != nil {
		log.Info(token.Error())
	}
}

func (mqtt *QMqtt) PublishObj(topic string, qos byte, in interface{}) {
	jsonBytes, _ := json.Marshal(in)
	sjson := string(jsonBytes)
	log.Info(topic)
	log.Info(sjson)
	mqtt.cnn.Publish(topic, qos, true, sjson)
}

//MqttDisocnnect mqtt disconnect
func (mqtt *QMqtt) Disocnnect() {
	mqtt.cliLock.Lock()
	mqtt.cnn.Disconnect(0)
	mqtt.cnn = nil
	mqtt.cliLock.Unlock()
}

//MqttConnServer mqtt init and connect server
func (mqtt *QMqtt) ConnServer(retry int, cbSubs CbRegistMqttSubs) error {
	suid := mqtt.Uid //"8df8de45-efa6-419a-a46c-acf30f017da5"
	if mqtt.cnn != nil {
		return errors.New("Mqtt already connected")
	}

	opts := MQTT.NewClientOptions().AddBroker(mqtt.MqttUrl)
	opts.SetClientID(suid)
	opts.SetUsername(mqtt.User)
	opts.SetPassword(mqtt.Pswd)
	opts.SetDefaultPublishHandler(cbMqttEdgeConfig)

	mqtt.cliLock.Lock()
	defer mqtt.cliLock.Unlock()
	//create and start a client using the above ClientOptions
	mqtt.cnn = MQTT.NewClient(opts)
	for ; retry != 0; retry-- {
		if token := mqtt.cnn.Connect(); token.Wait() && token.Error() == nil {
			if cbSubs != nil {
				cbSubs(mqtt)
			}
			log.Info("MQTT connect OK!")
			return nil
		} else {
			log.Error("Retry to connect the MQTT server!! ", token.Error())
		}
		time.Sleep(time.Duration(3) * time.Second)
	}
	return errors.New("Fault Error! can not connect MQTT server!!!")
}

func (mqtt *QMqtt) StartServer() error {
	return mqtt.ConnServer(-1, AppMqttSubs)
}

func (mqtt *QMqtt) StopServer() {
	mqtt.Disocnnect()
	log.Info("Stop MQTT server")
}

func (mqtt *QMqtt) Init() error {
	mqtt.baseBus.Init()
	go mqtt.StartServer()
	return nil
}

func (mqtt *QMqtt) Uninit() {
	mqtt.baseBus.Uninit()
	mqtt.StopServer()
}

func (mqtt *QMqtt) Send(chn IChannel, buff interface{}) (int, error) {
	mqtt.PublishObj(chn.ID(), mqtt.Qos, buff)
	return 0, nil
}

func (mqtt *QMqtt) Recive(chn IChannel, buff interface{}) (int, error) {
	return 0, nil
}

func AppMqttSubs(mqtt *QMqtt) int {
	mqtt.subTopic("/sub/default", 1, cbMqttEdgeConfig)
	return 0
}

func NewMqtt(param []interface{}) IBus {
	mq := &QMqtt{
		User:    "gZdomIS9Hz3d7HxvcoNx",
		Pswd:    "",
		MqttUrl: "test-sbuilding.pacom.cn:1885",
		Name:    "mqtt",
		cnn:     nil,
		cliLock: new(sync.Mutex),
	}

	if len(param) >= 1 {
		var cfgmq *QMqtt = nil
		switch param[0].(type) {
		case string:
			cfg := config.GetSysConfig().GetProfile(param[0].(string), mq)
			if cfg != nil {
				cfgmq = cfg.(*QMqtt)
			}
		case *QMqtt:
			cfgmq = param[0].(*QMqtt)
		}
		if cfgmq != nil {
			mq.User = cfgmq.User
			mq.MqttUrl = cfgmq.MqttUrl
			mq.Pswd = cfgmq.Pswd
			mq.Name = cfgmq.Name
		}
	}
	return mq
}

func init() {
	BusReg["mqtt"] = NewMqtt
}
