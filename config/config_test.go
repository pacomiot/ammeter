package config

import (
	"testing"
	"time"

	"github.com/yuguorong/go/log"
)

func TestConfigSet(T *testing.T) {
	cfg := GetSysConfig()
	cfg.LoadConfig()
	vi := cfg.GetValue("test", 1)
	cfg.SetValue("test", 2)
	cfg.SetValue("test", 2)
	cfg.SetValue("test", "2")
	log.Info(vi)

	cfg.Save()

}

type CloudServerConf struct {
	ServerUrl string
	AppId     string
	AppSecret string
}

func TestLoadConf(T *testing.T) {
	sys := GetSysConfig()
	sys.LoadConfig()
	conf := sys.GetValue("cloudserver/config", &CloudServerConf{})
	log.Info(conf)

	res := &CloudServerConf{
		AppId:     "fvWmjGCU",
		AppSecret: "054e7df0881eff8328858092f9e8ac0b0f356676",
		ServerUrl: "https://test-admin.pacom.cn",
	}
	sys.SetValue("cloudserver/config", res)
	conf = sys.GetValue("cloudserver/config", &CloudServerConf{})
	log.Info(conf)
	sys.Save()

}

type AmmeterModel struct {
	Id         string `json:"id" gorm:"-"`
	DevName    string `json:"devName" gorm:"primarykey"`
	Code       string `json:"code" `
	DutSn      string `json:"dutSn" gorm:"-"`
	Address    string `json:"address"`
	Protocol   string `json:"protocol" gorm:"-"`
	GwDevId    string `json:"gwDevId" gorm:"-"`
	TransRadio string `json:"transformerRatio" gorm:"-"`
}

type Ammeter struct {
	AmmeterModel
	timestamp   int32
	TotalEnergy float64
	TransDeno   float64 ` gorm:"-"`
	TransDiv    float64 ` gorm:"-"`
}

func TestDBConf(T *testing.T) {
	GetDB().CreateTbl(&Ammeter{})

	v0 := &Ammeter{
		TransDeno: 3,
		TransDiv:  2,
	}
	GetDB().Find(v0)
	log.Info(v0)

	v0.DevName = ""
	GetDB().Find(v0, "dev_name='test2'")
	log.Info(v0)

	if v0.DevName != "" {
		v1 := &Ammeter{
			TotalEnergy: 1373.78,
			timestamp:   int32(time.Now().Unix()),
			AmmeterModel: AmmeterModel{
				DevName: "test3",
				Code:    "2108",
				Address: "123456",
			},
		}

		GetDB().Save(v1)
	} else {
		v0.TotalEnergy += 100
		GetDB().Save(v0)
	}

}
