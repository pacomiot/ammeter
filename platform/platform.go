package platform

import (
	"github.com/yuguorong/go/log"

	bus "github.com/ammeter/Bus"
	"github.com/ammeter/cloudserver"
	"github.com/ammeter/config"
)

const (
	DEF_REQ_AMGW_URL = "/platform/dev/get-4G-gateway-list"
	REQ_AMMETER_URL  = "/platform/dev/get-ammeter-list"
	REQ_AIR_URL      = "/platform/dev/get-sair-list"
	DEF_FTP_PORT     = 10010
)

type PaPlatform struct {
	gwUrl       string
	GatewayList []*Gateway
	modelList   map[string]string
}

func (p *PaPlatform) SyncCloudGWConfig() *[]bus.QMqtt {
	if p.gwUrl == "" {
		p.gwUrl = DEF_REQ_AMGW_URL
	}
	mqttList := []bus.QMqtt{}
	prof := config.GetSysConfig().GetProfile("remote_config", cloudserver.GetCloudConfig())
	cs := cloudserver.GetCloudServer(prof)
	err := cs.GetClientData(p.gwUrl, &mqttList, nil)
	if err != nil {
		return nil
	}
	config.GetSysConfig().SetValue("mqtt_list", &mqttList)
	config.GetSysConfig().Save()
	return &mqttList
}

func (p *PaPlatform) LoadGatewayProfile() {
	p.GatewayList = make([]*Gateway, 0)
	mqttList := &[]bus.QMqtt{}
	config.GetSysConfig().GetValue("mqtt_list", mqttList)
	if mqrmt := p.SyncCloudGWConfig(); mqrmt != nil {
		mqttList = mqrmt
		config.GetSysConfig().SetValue("mqtt_list", mqttList)
	}
	for _, mq := range *mqttList {
		gw := InitGateway(&mq, p.modelList)
		p.GatewayList = append(p.GatewayList, gw)
		gw.StartServer(mq.Uid)
	}

}

func (p *PaPlatform) SetModel(sname string, surl string) {
	if p.modelList == nil {
		p.modelList = make(map[string]string)
	}
	p.modelList[sname] = surl
}

func (p *PaPlatform) SaveModel() {
	config.GetSysConfig().SetProfile("model_list", &p.modelList)
}

func (p *PaPlatform) LoadModles() {
	p.modelList = make(map[string]string)
	list := config.GetSysConfig().GetProfile("model_list", &p.modelList)
	if list != nil {
		p.modelList = *list.(*map[string]string)
	}
}

func (p *PaPlatform) SetGatewayUrl(url string) {
	p.gwUrl = url
}

var mainloop chan interface{}

func StartServer() {
	p := &PaPlatform{}
	p.LoadModles()
	p.LoadGatewayProfile()
	mainloop = make(chan interface{})

	v := <-mainloop
	log.Info("exit now", v)

}
