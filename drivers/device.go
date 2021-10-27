package drviers

import (
	"encoding/hex"
	"reflect"

	bus "github.com/ammeter/Bus"
	"github.com/ammeter/protocol"
	"github.com/yuguorong/go/log"
)

type IDevice interface {
	Probe(name string, drv IDriver) error
	Open(...interface{}) error
	Close() error
	Ctrl(...interface{}) error
	Suspend() error
	Resume() error
	GetDevice(string) interface{}
	SetRoute(evt bus.IBusEvent, prot string, busName string, chn string, param ...interface{}) int
}

type routePath struct {
	iChn   bus.IChannel
	ibus   bus.IBus
	iproto protocol.IProtocol
	router []chan interface{}
}

type baseDevice struct {
	DeviceName string
	drv        IDriver
	Route      []routePath
}

func (dev *baseDevice) Probe(name string, driver IDriver) error {
	dev.Route = make([]routePath, 0)
	dev.DeviceName = name
	dev.drv = driver
	return nil
}

//bus.IBusEvent
func (dev *baseDevice) OnAttach(chn bus.IChannel) {
}

func (dev *baseDevice) OnDetach(chn bus.IChannel) {
}

func (dev *baseDevice) ChannelDispatch(stream []byte, args interface{}) bus.ChnDispResult {
	log.Info(dev.drv.Name(), "-", dev.DeviceName, " try Dispatch: ", hex.EncodeToString(stream))
	k := reflect.TypeOf(args).Kind()
	if k == reflect.Ptr {
		iprot := args.(protocol.IProtocol)
		return bus.ChnDispResult(iprot.ChannelDispatch(stream))
	}
	return bus.DispatchNone
}

func (dev *baseDevice) Create(model interface{}) interface{} {
	return nil
}

func (dev *baseDevice) Open(...interface{}) error {
	return nil
}

func (dev *baseDevice) Close() error {
	return nil
}

func (dev *baseDevice) Ctrl(...interface{}) error {
	return nil
}

func (dev *baseDevice) GetDevice(string) interface{} {
	return dev
}

func (dev *baseDevice) Suspend() error {
	return nil
}

func (dev *baseDevice) Resume() error {
	return nil
}

func (dev *baseDevice) SetRoute(evt bus.IBusEvent, prot string, busName string, chn string, param ...interface{}) int {
	r := new(routePath)
	r.router = make([]chan interface{}, 1)
	r.router[0] = make(chan interface{}, 16)

	r.iproto = protocol.LoadProtocol(prot)
	r.iproto.Init(prot)

	r.ibus = bus.MountBus(busName, param)
	r.ibus.Init()
	r.iChn = r.ibus.OpenChannel(chn, r.router)
	r.iChn.SetEvent(evt, r.iproto)

	dev.Route = append(dev.Route, *r)
	return len(dev.Route)
}
