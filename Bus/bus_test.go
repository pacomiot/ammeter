package bus

import (
	"encoding/hex"
	"reflect"
	"testing"
	"time"

	"github.com/yuguorong/go/log"
)

var dbgChan IChannel = nil
var dbgBus IBus = nil
var dbgtx = []byte{
	0xfe, 0xfe, 0xfe, 0xfe, 0x68, 0x04, 0x00, 0x42, 0x05, 0x08, 0x21, 0x68, 0x11, 0x04, 0x33, 0x33, 0x33, 0x33, 0x25, 0x16,
}

//fefefefe68040042050821681104333333332516
func testrecive(chnExit chan bool, r []chan interface{}) {
	tmr := time.NewTimer(time.Second * 10)
	tmr.Stop()
	loop := true
	for loop {
		select {
		case msg := <-r[0]:
			switch reflect.TypeOf(msg).Kind() {
			case reflect.Slice:
				b := msg.([]byte)
				log.Info(hex.EncodeToString(b))
				if b[0] == 0x23 && b[1] == 0x23 {
					log.Info(string(b))
				}
			case reflect.String:
				log.Info(msg.(string))
				tmr.Reset(time.Second * 10)
			}
		case <-chnExit:
			loop = false
		case <-tmr.C:
			if dbgChan != nil && dbgBus != nil {
				log.Info(dbgChan.ID(), " in timer command send")
				dbgBus.Send(dbgChan, dbgtx)
				tmr.Reset(time.Second * 20)
			}

		}
	}
}

func mountbus(name string, param ...interface{}) IBus {
	return MountBus(name, param)
}

func testDispath(buf []byte, param interface{}) interface{} {

	log.Info(string(buf))
	log.Info(hex.EncodeToString(buf))
	if hex.EncodeToString(buf) == "9521003697" {
		return hex.EncodeToString(buf)
	}

	return nil
}

var deviceList map[string][]string

func TestDtuServer(T *testing.T) {
	deviceList = map[string][]string{
		"9521003872": {"2108", "05420005", "05420003", "05420002"}, //6f
		"9521003712": {"2108", "05420001"},                         //1F
		"9521003534": {"2108", "05420006"},                         //-1F
		"9521003697": {"2108", "05420004"},                         //5c
	}

	router := make([]chan interface{}, 2)
	router[0] = make(chan interface{}, 16)
	router[1] = make(chan interface{}, 4)
	chExit := make(chan bool, 2)
	go testrecive(chExit, router)

	dbgBus = mountbus("dtuFtpServer", 10010)
	dbgBus.Init()

	dbgChan = dbgBus.OpenChannel("9521003697", router)

	time.Sleep(time.Second * 40)
	chExit <- false
	dbgBus.CloseChannel(dbgChan)
	time.Sleep(time.Second * 10)
	dbgBus.Uninit()
}
