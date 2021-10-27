package platform

import (
	"testing"

	"github.com/ammeter/config"
	"github.com/yuguorong/go/log"
)

func TestGateway(T *testing.T) {
	//gw := InitGateway("")
	//gw.GetGatewayName()
	//gw.StartServer()
}

func TestPlatform(T *testing.T) {
	config.GetSysConfig().SetValue("Bus/DtuServer/Port", DEF_FTP_PORT)
	log.Info("process start...")
	p := PaPlatform{}
	p.SetGatewayUrl("/platform/dev/get-4G-gateway-list")
	p.SetModel("ammeter", REQ_AMMETER_URL)
	p.SetModel("SAIR10", REQ_AIR_URL)
	p.SaveModel()

}
