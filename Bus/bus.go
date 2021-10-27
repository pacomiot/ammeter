package bus

import (
	"reflect"
	"strconv"
	"sync"
	"time"
)

const (
	CHAN_INPUT  = 0
	CHAN_OUTPUT = 1
)

type CbBusChanDisp func(stream []byte, param interface{}) interface{}
type BusChannel struct {
	chnID     string
	mountCnt  int
	event     IBusEvent
	bus       IBus
	timeout   time.Duration
	conIdList []int
	evtArg    interface{}
	chin      chan interface{}
	chout     chan interface{}
}

func (chn *BusChannel) ID() string {
	return chn.chnID
}

func (chn *BusChannel) GetChan(id int) chan interface{} {
	switch id {
	case 0:
		return chn.chin
	case 1:
		return chn.chout
	}
	return nil
}

func (chn *BusChannel) SetTimeout(timeout time.Duration) {

}

func (chn *BusChannel) SetEvent(evt IBusEvent, evtArgs interface{}) {
	chn.event = evt
	chn.evtArg = evtArgs
}

func (chn *BusChannel) GetEvent() IBusEvent {
	return chn.event
}

func (chn *BusChannel) GetBus() IBus {
	return chn.bus
}

type baseBus struct {
	BusId   string
	ChnList map[string]IChannel
	mutex   *sync.Mutex
}

func (bus *baseBus) Init() error {
	if bus.ChnList == nil {
		bus.ChnList = make(map[string]IChannel)
		bus.mutex = new(sync.Mutex)
	}

	return nil
}

func (bus *baseBus) Uninit() {

}

func (bus *baseBus) stringId(chnID interface{}) string {
	switch reflect.TypeOf(chnID).Kind() {
	case reflect.String:
		return chnID.(string)
	case reflect.Int:
		return strconv.Itoa(chnID.(int))
	case reflect.Int64:
		return strconv.FormatInt(chnID.(int64), 10)
	}
	return ""
}

func (bus *baseBus) OpenChannel(chnID interface{}, router []chan interface{}) IChannel {
	schnid := bus.stringId(chnID)
	if chn, has := bus.ChnList[schnid]; has {
		return chn
	}
	c := &BusChannel{
		chnID:     schnid,
		event:     nil,
		mountCnt:  0,
		bus:       bus,
		conIdList: make([]int, 0),
		timeout:   0,
	}
	if len(router) > 1 {
		c.chout = router[1]
	}
	if len(router) > 0 {
		c.chin = router[0]
	}

	bus.mutex.Lock()
	bus.ChnList[schnid] = c
	bus.mutex.Unlock()

	return c
}
func (dtu *baseBus) ResetChannel(chn IChannel) {
}

func (bus *baseBus) CloseChannel(chn IChannel) error {
	id := chn.ID()
	bus.mutex.Lock()
	delete(bus.ChnList, id)
	bus.mutex.Unlock()
	basechn := chn.(*BusChannel)
	basechn.mountCnt = 0
	if basechn.event != nil {
		basechn.event.OnDetach(chn)
	}
	return nil
}

func (bus *baseBus) ScanChannel(stream []byte, connID int) IChannel {
	for _, ichn := range bus.ChnList {
		chn := ichn.(*BusChannel)
		if chn.event == nil || chn.mountCnt < 0 {
			continue
		}
		if ret := chn.event.ChannelDispatch(stream, chn.evtArg); ret != DispatchNone {
			chn.mountCnt += int(ret)
			chn.conIdList = append(chn.conIdList, connID)
			chn.event.OnAttach(ichn)
			return ichn
		}
	}
	return nil
}

func (bus *baseBus) Send(chn IChannel, buff interface{}) (int, error) {
	return 0, nil
}

func (bus *baseBus) Recive(chn IChannel, buff interface{}) (int, error) {
	return 0, nil
}

type funcRegBus func(param []interface{}) IBus
type funcGetID func(string, []interface{}) string

var BusList map[string]IBus
var BusGetID map[string]funcGetID
var BusReg map[string]funcRegBus

func MountBus(name string, param []interface{}) IBus {
	busid := name
	if f, has := BusGetID[name]; has {
		busid = f(name, param)
	}
	if p, has := BusList[busid]; has && p != nil {
		return p
	}
	if f, has := BusReg[name]; has && f != nil {
		b := f(param)
		BusList[busid] = b
		return b
	}
	return nil
}

func init() {
	BusList = make(map[string]IBus)
	BusReg = make(map[string]funcRegBus)
	BusGetID = make(map[string]funcGetID)
}
