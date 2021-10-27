package main

import (
	"net"

	"github.com/yuguorong/go/log"

	"github.com/ammeter/config"
	"github.com/ammeter/platform"
)

//Port 端口
const (
	REQ_AMGW_URL    = "/platform/dev/get-4G-gateway-list"
	REQ_AMMETER_URL = "/platform/dev/get-ammeter-list"
	REQ_AIR_URL     = "/platform/dev/get-sair-list"
	DEF_FTP_PORT    = 10010
)

var Port string

var maxConnCount = 100

var connclientSize = 0
var connlist map[string]net.Conn

var deviceList map[string][]string

func init() {
	connlist = make(map[string]net.Conn)
	deviceList = map[string][]string{
		"9521003872": {"2108", "05420005", "05420003", "05420002"}, //6f
		"9521003712": {"2108", "05420001"},                         //1F
		"9521003534": {"2108", "05420006"},                         //-1F
		"9521003697": {"2108", "05420004"},                         //5c
	}
}

func main() {

	config.GetSysConfig().SetValue("Bus/DtuServer/Port", DEF_FTP_PORT)
	log.Info("process start...")
	p := platform.PaPlatform{}
	p.SetGatewayUrl("/platform/dev/get-4G-gateway-list")
	p.SetModel("ammeter", REQ_AMMETER_URL)
	p.SetModel("SAIR10", REQ_AIR_URL)
	p.SaveModel()

	platform.StartServer()
}
