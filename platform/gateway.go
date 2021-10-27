package platform

import (
	"errors"
	"strings"

	"github.com/yuguorong/go/log"

	bus "github.com/ammeter/Bus"
	"github.com/ammeter/cloudserver"
	"github.com/ammeter/config"
	drviers "github.com/ammeter/drivers"
)

/*
gZdomIS9Hz3d7HxvcoNx
data [{"accessToken":"AZM7QSPRqsV4CUxIGldn",
"id":"1ec23004b88a020bf52b38d8579115b",
"mqttServer":"test-sbuilding.pacom.cn:1885",
"name":"FGGW10-202110020001"}]
*/
var debugName = true

const (
	DEF_VERDOR   = "GW4G"
	DEF_INF_NAME = "eth0"
)

type Gateway struct {
	Mqttcfg    bus.QMqtt
	Name       string
	Uid        string
	modelList  map[string]string
	NorthRoute interface{}
	DeviceList []drviers.IDriver
}

func (gw *Gateway) SyncCloudModelDevice(url string, model interface{}) error {
	var err = errors.New("Nil model?")
	if model != nil {
		prof := config.GetSysConfig().GetProfile("remote_config", cloudserver.GetCloudConfig())
		cs := cloudserver.GetCloudServer(prof)
		param := map[string]string{
			"gwDevId": gw.Uid,
		}
		err = cs.GetClientData(url, model, &param)
		log.Info(model)
	}
	return err
}

type netinfo struct {
	name    string
	uuid    string
	prority int
}

func priotyInf(ni *netinfo, name string) bool {
	name = strings.ToUpper(name)
	ifnames := []string{"WLAN", "ETH", "ENS", "ENO", "本地连接"}
	for i := 0; i < ni.prority && i < len(ifnames); i++ {
		if strings.Contains(name, ifnames[i]) {
			ni.name = name
			ni.prority = i
			return true
		}
	}
	return false
}

func (gw *Gateway) InitGateway(modelList map[string]string) {
	//gw.GetGatewayName()
	gw.DeviceList = make([]drviers.IDriver, 0)
	gw.modelList = modelList

}

func (gw *Gateway) StartServer(uid string) {
	//	name := "ammeter"
	//	drvU := drviers.Install("SAIR10", nil)
	for name, url := range gw.modelList {
		drv := drviers.Install(name, nil)
		model := drv.GetModel()
		if url != "" {
			if gw.SyncCloudModelDevice(url, model) != nil {
				model = config.GetSysConfig().GetProfile("gateway/"+name+"/amlist", model)
			} else {
				config.GetSysConfig().SetProfile("gateway/"+name+"/amlist", model)
			}
		}
		if sz, devlist := drv.CreateDevice(model); sz > 0 {
			for _, dev := range devlist {
				dev.SetRoute(nil, "ThingsBoards", "mqtt", "v1/gateway/telemetry", &gw.Mqttcfg)
				dev.Open()
			}
		}
	}
}

func InitGateway(mqttcfg *bus.QMqtt, modelList map[string]string) *Gateway {
	gw := &Gateway{
		Uid:     mqttcfg.Uid,
		Mqttcfg: *mqttcfg,
	}
	gw.InitGateway(modelList)
	return gw
}
