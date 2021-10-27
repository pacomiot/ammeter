package cloudserver

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/yuguorong/go/log"
)

const (
	REQ_AMGW_URL    = "/platform/dev/get-4G-gateway-list"
	REQ_AMMETER_URL = "/platform/dev/get-ammeter-list"
)

//4G 网关
type GateWay4GModel struct {
	AccessToken string `json:"accessToken"`
	MqttServer  string `json:"mqttserver"`
	Id          string `json:"id"`
	Name        string `json:"name"`
}

type AmmeterModel struct {
	Id       string `json:"id"`
	DevName  string `json:"devName"`
	Code     string `json:"code"`
	DutSn    string `json:"dutSn"`
	Address  string `json:"address"`
	Protocol string `json:"protocol"`
	GwDevId  string `json:"gwDevId"`
}

func TestForGetAmmeterGwList(t *testing.T) {
	// t.Error(1)
	conf := getTestConf()
	cloudServer := GetCloudServer(conf)
	gwList := []GateWay4GModel{}
	err := cloudServer.GetClientData(REQ_AMGW_URL, &gwList, nil)
	if err != nil || len(gwList) == 0 {
		t.Error("test for getAmmeter error,please check the gw data  in server")
	}
	// for i := 0; i < len(gwList); i++ {
	// 	d := gwList[i]
	// 	log.Println("query ammeter by gw", d)
	// 	ammeters := cloudServer.GetAmmeters(d.Id)
	// 	log.Println("gw ammeters:", ammeters)
	// }
}

func TestForGetAmmeterList(t *testing.T) {
	// t.Error(1)
	conf := getTestConf()
	cloudServer := GetCloudServer(conf)
	gwList := []GateWay4GModel{}
	err := cloudServer.GetClientData(REQ_AMGW_URL, &gwList, nil)
	if err != nil || len(gwList) == 0 {
		t.Error("test for getAmmeter error,please check the gw data  in server")
	}
	checkRes := false
	for i := 0; i < len(gwList); i++ {
		d := gwList[i]
		param := map[string]string{
			"gwDevId": d.Id,
		}
		listMeters := []AmmeterModel{}
		err := cloudServer.GetClientData(REQ_AMMETER_URL, &listMeters, &param)
		if err == nil && !checkRes && len(listMeters) > 1 {
			log.Info(listMeters)
			checkRes = true
			break
		}
	}
	if !checkRes {
		t.Error("test for getAmmeter error,please check the ammeter in your server ")
	}
}

func getTestConf() *CloudServerConf {
	res := CloudServerConf{
		AppId:     "fvWmjGCU",
		AppSecret: "054e7df0881eff8328858092f9e8ac0b0f356676",
		ServerUrl: "https://test-admin.pacom.cn",
	}
	return &res
}

func TestForPageChangeData(t *testing.T) {
	t.Log(1)
	p := PageModel{
		Total: 1,
		Rows: []interface{}{
			GateWay4GModel{AccessToken: "xx"}, GateWay4GModel{MqttServer: "xx123"},
		},
	}

	var dd []GateWay4GModel
	p.ChangeData(&dd)
	t.Log(dd)
}

func TestConf(t *testing.T) {
	f, _ := os.OpenFile("cloudconf.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer f.Close()
	jsonStu, _ := json.Marshal(GetCloudConfig())
	f.WriteString(string(jsonStu))
	log.Info("SysConfig saved!")
}
