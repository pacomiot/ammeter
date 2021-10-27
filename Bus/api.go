package bus

import "time"

//Channel Dispatch not in event because event and dispatch handler not same one.

type ChnDispResult int

const (
	DispatchSingle ChnDispResult = -1 + iota
	DispatchNone
	DispatchMulti
)

//Dispatch maybe a protocol object
type IBusEvent interface {
	OnAttach(chn IChannel)
	OnDetach(chn IChannel)
	ChannelDispatch(stream []byte, argEvt interface{}) ChnDispResult
}

type IChannel interface {
	ID() string
	GetChan(id int) chan interface{}
	SetTimeout(time.Duration)
	SetEvent(evt IBusEvent, evtArgs interface{})
	GetEvent() IBusEvent
	GetBus() IBus
}

type IBus interface {
	Init() error
	Uninit()
	OpenChannel(chnID interface{}, router []chan interface{}) IChannel
	CloseChannel(chn IChannel) error
	ResetChannel(chn IChannel)
	ScanChannel(stream []byte, conn int) IChannel
	Send(chn IChannel, buff interface{}) (int, error)
	Recive(chn IChannel, buff interface{}) (int, error)
}
