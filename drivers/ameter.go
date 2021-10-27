package drviers

import (
	"encoding/hex"
	"reflect"
	"strconv"
	"strings"
	"time"

	bus "github.com/ammeter/Bus"
	"github.com/ammeter/config"
	"github.com/yuguorong/go/log"
)

const (
	CMD_ENERGY_TOTAL         = "00000000"
	DEF_SAMPLE_PERIOD        = 5 * time.Minute
	DEF_SAMPLE_PEER_DURATION = 30 * time.Second
	DEF_TCP_READ_TIMEOUT     = 1 * time.Minute
	POSTPONE_PERSIST_TIME    = 1 * time.Minute
)

type AmmeterModel struct {
	Id         string `json:"id" gorm:"-"`
	DevName    string `json:"devName" gorm:"primarykey"`
	Code       string `json:"code" gorm:"-"`
	DutSn      string `json:"dutSn" gorm:"-"`
	Address    string `json:"address" gorm:"-"`
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

type AmMeterHub struct {
	baseDevice
	DutSn      string
	Protocol   string
	chnExit    chan bool
	loopserver bool
	sampleTmr  *time.Timer
	amList     map[string]*Ammeter
	persitList map[string]*Ammeter
	tmrPersist *time.Timer
	queryIdx   int
}

func (dev *AmMeterHub) Close() error {
	dev.sampleTmr.Stop()
	dev.loopserver = false
	dev.chnExit <- true
	return nil
}

func (dev *AmMeterHub) OnAttach(chn bus.IChannel) {
	log.Info(dev.DutSn, " Attached!")
	chn.SetTimeout(DEF_TCP_READ_TIMEOUT)
	dev.sampleTmr.Reset(DEF_SAMPLE_PEER_DURATION)
}

func (dev *AmMeterHub) OnDetach(chn bus.IChannel) {
	log.Info(dev.DutSn, " Detached!")
	dev.sampleTmr.Stop()
}

func (dev *AmMeterHub) ChannelDispatch(stream []byte, args interface{}) bus.ChnDispResult {
	if len(stream) != 5 {
		return bus.DispatchNone
	}
	if ret := dev.baseDevice.ChannelDispatch(stream, args); ret != bus.DispatchNone {
		return ret
	}

	sid := hex.EncodeToString(stream)
	for _, am := range dev.amList {
		if am.DutSn == sid {
			log.Info("MOUNT meter: ", am.DutSn, ",", sid)
			return bus.DispatchSingle
		}
	}
	return bus.DispatchNone
}

func (dev *AmMeterHub) IssueSampleCmd(tmrTx *time.Timer) (pam *Ammeter) {
	i := 0
	for _, am := range dev.amList {
		if i == dev.queryIdx {
			pam = am
			break
		}
		i++
	}
	if pam == nil {
		return nil
	}
	pack := dev.Route[1].iproto.PackageCmd(CMD_ENERGY_TOTAL, pam.Code+pam.Address)
	dev.Route[1].ibus.Send(dev.Route[1].iChn, pack)
	tmrTx.Reset(1 * time.Second)
	return pam
}

func (dev *AmMeterHub) SchedulNextSample(tmrTx *time.Timer) {
	tmrTx.Stop()
	dev.queryIdx++
	if dev.queryIdx >= len(dev.amList) {
		dev.queryIdx = 0
		dev.sampleTmr.Reset(DEF_SAMPLE_PERIOD)
	} else {
		dev.sampleTmr.Reset(DEF_SAMPLE_PEER_DURATION)
	}
}

func (dev *AmMeterHub) AdjustTelemetry(mval interface{}, devname string) {
	vlist := mval.(map[string]interface{})
	if am, has := dev.amList[devname]; has {
		for k, v := range vlist {
			if reflect.TypeOf(v).Kind() == reflect.Float64 {
				valItem := v.(float64)
				vlist[k] = (valItem * am.TransDeno) / am.TransDiv
			}
		}

		if val, has := vlist["TotalActivePower"]; has {
			diff := val.(float64) - am.TotalEnergy
			am.TotalEnergy = val.(float64)
			vlist["ActivePowerIncrement"] = diff
			if len(dev.persitList) == 0 {
				dev.tmrPersist.Reset(POSTPONE_PERSIST_TIME)
			}
			dev.persitList[am.DevName] = am
		}
	}
}

func (dev *AmMeterHub) Run() error {
	dev.loopserver = true
	tmrTxTimout := time.NewTimer(5 * time.Second)
	dev.tmrPersist = time.NewTimer(POSTPONE_PERSIST_TIME)
	tmrTxTimout.Stop()
	//var pam *Ammeter = nil
	for dev.loopserver {
		select {
		case msg := <-dev.Route[1].router[0]:
			if reflect.TypeOf(msg).Kind() == reflect.Slice {
				b := msg.([]byte)
				log.Info(hex.EncodeToString(b))
				devname := "" //pam.DevName
				mret := dev.Route[1].iproto.ParsePacket(b, &devname)
				if mret != nil && devname != "" {
					log.Info(devname, mret)
					devname = "AM10-" + devname + dev.DutSn[:4]
					dev.AdjustTelemetry(mret, devname)
					telemetry := dev.Route[0].iproto.PackageCmd(devname, mret)
					dev.Route[0].ibus.Send(dev.Route[0].iChn, telemetry)
					dev.SchedulNextSample(tmrTxTimout)
				}
			}
		case <-tmrTxTimout.C:
			dev.SchedulNextSample(tmrTxTimout)
		case <-dev.sampleTmr.C:
			dev.IssueSampleCmd(tmrTxTimout)
		case <-dev.chnExit:
			dev.loopserver = false
		case <-dev.tmrPersist.C:
			for k, am := range dev.persitList {
				config.GetDB().Save(am)
				delete(dev.persitList, k)
			}
		}
	}
	return nil

}

func DutchanDispatch(rxin interface{}, param interface{}) interface{} {
	log.Info("DutchanDispatch")
	b := rxin.([]byte)
	if len(b) == 8 {
		return hex.EncodeToString(b)
	}
	return nil
}

func (dev *AmMeterHub) Open(param ...interface{}) error {

	if dev.DutSn != "" {
		dev.SetRoute(dev, dev.Protocol, "dtuFtpServer", dev.DutSn, 0)
	}
	dev.sampleTmr = time.NewTimer(DEF_SAMPLE_PERIOD)
	dev.sampleTmr.Stop()

	go dev.Run()
	return nil
}

func (dev *AmMeterHub) GetDevice(devname string) interface{} {
	if dev, has := dev.amList[devname]; has {
		return dev
	}
	return nil
}

/*
[{"code":"26462285","devName":"SAIR10-0000000026462285","gwDevId":"1ec2d8421b2ed30bf52b38d8579115b","id":"1ec2d84cc494690bf52b38d8579115b","protocol":"HJ212"},
 {"code":"61748803","devName":"SAIR10-0000000061748803","gwDevId":"1ec2d8421b2ed30bf52b38d8579115b","id":"1ec2d84eb428160bf52b38d8579115b","protocol":"HJ212"}]
 {"address":"05420001","code":"2108","devName":"AM10-2108054200019521","dutSn":"9521003712","gwDevId":"1ec2d8421b2ed30bf52b38d8579115b","id":"1ec2de6df5a2e40bf52b38d8579115b","protocol":"DLT645-2007","transformerRatio":"80"},
*/

type AmmeterDrv struct {
	baseDriver
}

func (drv *AmmeterDrv) ParseTransRatio(dev *Ammeter) {
	Trans := strings.Split(dev.TransRadio, "/")
	dev.TransDeno, _ = strconv.ParseFloat(Trans[0], 64)
	if len(Trans) > 1 {
		dev.TransDiv, _ = strconv.ParseFloat(Trans[1], 64)
	}
}

func (drv *AmmeterDrv) CreateDevice(model interface{}) (int, []IDevice) {
	if model != nil {
		models := model.(*[]AmmeterModel)
		for _, m := range *models {
			var hub IDevice = nil
			var has bool
			if hub, has = drv.DevList[m.DutSn]; !has {
				hub = &AmMeterHub{
					DutSn:      m.DutSn,
					Protocol:   m.Protocol,
					chnExit:    make(chan bool),
					loopserver: false,
					sampleTmr:  nil,
					amList:     make(map[string]*Ammeter),
					persitList: make(map[string]*Ammeter),
					queryIdx:   0,
				}
				hub.Probe(m.DutSn, drv)
				drv.DevList[m.DutSn] = hub
			}

			dev := &Ammeter{
				TotalEnergy: 0,
				timestamp:   0,
				TransDiv:    1,
				TransDeno:   1,
			}
			config.GetDB().Find(dev, "dev_name='"+m.DevName+"'")
			dev.AmmeterModel = m
			drv.ParseTransRatio(dev)
			hub.(*AmMeterHub).amList[dev.DevName] = dev
		}
	}
	return drv.baseDriver.CreateDevice(model)
}

func (drv *AmmeterDrv) GetModel() interface{} {
	return &[]AmmeterModel{}
}

func NewAmMeter(param interface{}) IDriver {
	am := new(AmmeterDrv)
	return am
}

func init() {
	driverReg["ammeter"] = NewAmMeter
	config.GetDB().CreateTbl(&Ammeter{})
}
