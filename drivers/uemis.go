package drviers

import (
	"errors"

	"github.com/yuguorong/go/log"
)

type Uemis struct {
	baseDevice
	chnExit    chan bool
	loopserver bool
}

func (u *Uemis) Run() error {
	u.loopserver = true
	for u.loopserver {
		select {
		case msg := <-u.Route[1].router[0]:
			b := msg.([]byte)
			devName := ""
			mret := u.Route[1].iproto.ParsePacket(b, &devName)
			if mret != nil && devName != "" {
				v := mret.(map[string]interface{})
				dev := devName[8:]
				dev = u.DeviceName + "-00000000" + dev
				log.Info(dev, v)
				telemetry := u.Route[0].iproto.PackageCmd(dev, mret)
				u.Route[0].ibus.Send(u.Route[0].iChn, telemetry)
			}
		case <-u.chnExit:
			u.loopserver = false

		}
	}
	return nil
}

func (u *Uemis) Open(...interface{}) error {
	lenr := len(u.Route)
	if lenr < 1 {
		return errors.New("wrong route config")
	} else if lenr < 2 {
		u.SetRoute(u, "HJ212", "dtuFtpServer", "HJ212", 0)
	}
	if !u.loopserver {
		go u.Run()
	}
	return nil
}

func (u *Uemis) Close() error {
	u.loopserver = false
	u.chnExit <- true
	return nil
}

func (u *Uemis) GetDevice(s string) interface{} {
	return u
}

type SAirModel struct {
	Id       string `json:"id"`
	DevName  string `json:"devName"`
	Code     string `json:"code"`
	Protocol string `json:"protocol"`
	GwDevId  string `json:"gwDevId"`
}

type UemisDrv struct {
	baseDriver
}

func (drv *UemisDrv) GetModel() interface{} {
	return &[]SAirModel{}
}

func (drv *UemisDrv) CreateDevice(model interface{}) (int, []IDevice) {
	drv.baseDriver.CreateDevice(model)
	dev := new(Uemis)
	dev.loopserver = false
	dev.Probe(drv.Name(), drv)
	drv.baseDriver.DevList[drv.Name()] = dev
	return drv.baseDriver.CreateDevice(model)
}

func NewUemis(param interface{}) IDriver {
	drv := new(UemisDrv)
	return drv
}

func init() {
	driverReg["SAIR10"] = NewUemis
}
