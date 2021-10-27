package bus

import (
	"encoding/hex"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/yuguorong/go/log"

	"github.com/ammeter/config"
)

const (
	maxConnCount = 100
	moduleName   = "dtuFtpServer"
	DEF_FTP_PORT = 10010
)

type DtuServer struct {
	baseBus
	Port      int
	name      string
	lis       net.Listener
	connlist  map[int]net.Conn
	clientNum int
	loop      bool
}

func (dtu *DtuServer) ResetChannel(chn IChannel) {

}

func (dtu *DtuServer) closeConn(conn net.Conn, connid int, chn IChannel) {
	dtu.mutex.Lock()
	if _, has := dtu.connlist[connid]; has {
		delete(dtu.connlist, connid)
		dtu.clientNum--
	}
	dtu.mutex.Unlock()

	conn.Close()
	if chn != nil {
		basechn := chn.(*BusChannel)
		log.Infof("client [%s] close\n", basechn.chnID)
		basechn.mountCnt = 0
		if basechn.event != nil {
			basechn.event.OnDetach(chn)
		}
		for i, id := range basechn.conIdList {
			if id == connid {
				basechn.conIdList[i] = -1
				break
			}
		}
	}
}

func (dtu *DtuServer) LookupChannel(buf []byte, idx int, conn net.Conn) IChannel {
	return dtu.baseBus.ScanChannel(buf, idx)
}

func (dtu *DtuServer) ClientConnect(conn net.Conn, idx int) {
	var ic IChannel = nil
	remoteAddr := conn.RemoteAddr().String()
	defer dtu.closeConn(conn, idx, ic)

	var err error
	var buf [1024]byte
	n := 0
	for ic == nil {
		n, err = conn.Read(buf[:])
		if err != nil {
			log.Infof("read from %s header faild err:[%v]\n", remoteAddr, err)
			return
		}
		ic = dtu.LookupChannel(buf[:n], idx, conn)
	}

	chnin := ic.GetChan(0)
	if chnin == nil {
		panic("no chan for read message")
	}
	chnout := ic.GetChan(1)
	dtu.name = hex.EncodeToString(buf[:n])

	for {
		smeter := hex.EncodeToString(buf[:n])
		dtu.mutex.Lock()
		chnin <- buf[:n]
		dtu.mutex.Unlock()

		if chnout != nil {
			after := time.After(time.Second * 60)
			select {
			case msg := <-chnout:
				conn.Write(msg.([]byte))
			case <-after:
				break
			}
		}
		// log.Printf("rev data from %s msg:%s\n", conn.RemoteAddr().String(), string(buf[:n]))
		log.Infof("[%s]rev data from %s msg(%d):[%s]\n", time.Now().Format("2006-01-02 15:04:05"), remoteAddr, n, smeter)

		if ic.(*BusChannel).timeout != 0 {
			conn.SetReadDeadline(time.Now().Add(ic.(*BusChannel).timeout))
		}
		n, err = conn.Read(buf[:])
		if err != nil {
			log.Errorf("read from %s msg faild err:[%v]\n", ic.ID(), err)
			dtu.closeConn(conn, idx, ic)
			break
		}
	}
}

func (dtu *DtuServer) StartServer() {
	defer func() {
		dtu.lis.Close()
	}()
	idx := 0

	for dtu.loop {
		if dtu.clientNum >= maxConnCount {
			log.Infof("there is %d clients,is max num\n", dtu.clientNum)
			time.Sleep(5 * time.Second)
			continue
		}
		conn, err := dtu.lis.Accept()
		if err != nil {
			log.Errorf("listen err:[%v]\n", err)
		}
		dtu.mutex.Lock()
		dtu.connlist[idx] = conn
		idx++
		dtu.clientNum++
		dtu.mutex.Unlock()
		go dtu.ClientConnect(conn, idx-1)
	}
}

func (dtu *DtuServer) Init() error {
	dtu.baseBus.Init()
	if !dtu.loop {
		addr := fmt.Sprintf("0.0.0.0:%d", dtu.Port)
		log.Info("start ", addr)
		var err error = nil
		dtu.lis, err = net.Listen("tcp", addr)
		if err != nil {
			log.Errorf("err! addr %s open faild err:[%v]\n", addr, err)
			return err
		}
		dtu.mutex.Lock()
		defer dtu.mutex.Unlock()

		dtu.loop = true
		go dtu.StartServer()
	}
	return nil
}

func (dtu *DtuServer) Uninit() {
	dtu.mutex.Lock()
	defer dtu.mutex.Unlock()
	if dtu.loop {
		dtu.loop = false
		for _, c := range dtu.connlist {
			c.Close()
		}
		dtu.lis.Close()
		dtu.connlist = make(map[int]net.Conn)
	}
}

func (dtu *DtuServer) OpenChannel(chn interface{}, router []chan interface{}) IChannel {
	if chn == nil || reflect.TypeOf(chn).Kind() != reflect.String {
		panic("Open channel should be a uniq string")
	}
	ic := dtu.baseBus.OpenChannel(chn, router)
	return ic
}

func (dtu *DtuServer) CloseChannel(chn IChannel) error {
	dtu.baseBus.CloseChannel(chn)
	if dtu.loop {
		//idx := chn.(*BusChannel).conn.(int)
		for _, connID := range chn.(*BusChannel).conIdList {
			if conn, has := dtu.connlist[connID]; has {
				dtu.mutex.Lock()
				delete(dtu.connlist, connID)
				dtu.mutex.Unlock()
				conn.Close()
			}
		}
	}
	return nil
}

func (dtu *DtuServer) Send(ichn IChannel, buff interface{}) (int, error) {
	if ichn != nil {
		chn := ichn.(*BusChannel)
		for _, connID := range chn.conIdList {
			if connID >= 0 {
				if conn, has := dtu.connlist[connID]; has {
					conn.Write(buff.([]byte))
				}
			}
		}
	}
	return 0, nil
}

func GetFtpServerConfig(param []interface{}) int {
	Port := 0
	if param != nil && len(param) > 0 && reflect.TypeOf(param[0]).Kind() == reflect.Int {
		Port = param[0].(int)
	}

	if Port == 0 {
		Port = config.GetSysConfig().GetValue("Bus/DtuServer/Port", DEF_FTP_PORT).(int)
	}
	return Port
}

func NewDtuServer(param []interface{}) IBus {
	Port := GetFtpServerConfig(param)
	busid := GenDtuServerId(moduleName, param)
	b := &DtuServer{
		baseBus: baseBus{
			BusId: busid,
			mutex: &sync.Mutex{},
		},
		connlist:  make(map[int]net.Conn),
		clientNum: 0,
		Port:      Port,
		loop:      false,
	}
	return b
}

func GenDtuServerId(name string, param []interface{}) string {
	Port := GetFtpServerConfig(param)
	return name + ":" + strconv.Itoa(Port)
}

func init() {
	BusReg[moduleName] = NewDtuServer
	BusGetID[moduleName] = GenDtuServerId
}
