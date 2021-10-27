package protocol

import (
	"time"
)

type ThingsBoards struct {
	Name string
}

/*
{
  "Device A": [
    {
      "ts": 1483228800000,
      "values": {
        "temperature": 42,
        "humidity": 80
      }
    },
    {
      "ts": 1483228801000,
      "values": {
        "temperature": 43,
        "humidity": 82
      }
    }
  ],
  "Device B": [
    {
      "ts": 1483228800000,
      "values": {
        "temperature": 42,
        "humidity": 80
      }
    }
  ]

  {
    "UEMIS61748803":[
        {
            "ts":1633471518102004,
            "values":{
                "CH2O-Rtd":"0.007",
                "CO2-Rtd":"399.0",
                "HUMI-Rtd":"38.8",
                "PM25-Rtd":"1.0",
                "TEMP-Rtd":"20.4",
                "VOC-Rtd":"0.022"
            }
        }
    ]
}
}
*/
func (tbs *ThingsBoards) Init(name string) error {
	return nil
}

func (tbs *ThingsBoards) Uninit() {
}

func (tbs *ThingsBoards) ChannelDispatch(b []byte) int {
	return 0
}

func (tbs *ThingsBoards) ParsePacket(buff []byte, param ...interface{}) interface{} {

	return nil
}

func (tbs *ThingsBoards) rptDevTelemetry(devName string, telemetry interface{}) interface{} {
	tbObj := make(map[string]interface{})
	devVal := make(map[string]interface{})
	devVal["ts"] = time.Now().UnixNano() / (1000 * 1000)
	devVal["values"] = telemetry
	devList := make([]interface{}, 1)
	devList[0] = devVal
	tbObj[devName] = devList
	return tbObj

}

func (tbs *ThingsBoards) PackageCmd(dev string, param ...interface{}) interface{} {
	v := param[0].(map[string]interface{})
	return tbs.rptDevTelemetry(dev, v)
}

func NewThingsboard() IProtocol {
	p := &ThingsBoards{}
	p.Init("ThingsBoards")
	return p
}

func init() {
	ProtReg["ThingsBoards"] = NewThingsboard
}
