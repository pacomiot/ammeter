package protocol

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/yuguorong/go/log"
)

type HJ212 struct {
	Name string
}

/*
[2021-10-06 03:05:26]rev data from 221.216.136.161:1114 msg(171):[23233031353953543d32323b434e3d323031313b50573d3132333435363b4d4e3d383838383838383832363436323238353b43503d26264461746154696d653d32303231313030363033303532373b54454d502d5274643d32322e303b48554d492d5274643d34332e353b504d32352d5274643d322e303b434f322d5274643d3430352e303b4348324f2d5274643d302e3030383b564f432d5274643d302e3035323b2626386566370d0a]
&{0xc000010050 [35 35 48 49 53 57 83 84 61 50 50 59 67 78 61 50 48 49 49 59 80 87 61 49 50 51 52 53 54 59 77 78 61 56 56 56 56 56 56 56 56 50 54 52 54 50 50 56 53 59 67 80 61 38 38 68 97 116 97 84 105 109 101 61 50 48 50 49 49 48 48 54 48 51 48 53 50 55 59 84 69 77 80 45 82 116 100 61 50 50 46 48 59 72 85 77 73 45 82 116 100 61 52 51 46 53 59 80 77 50 53 45 82 116 100 61 50 46 48 59 67 79 50 45 82 116 100 61 52 48 53 46 48 59 67 72 50 79 45 82 116 100 61 48 46 48 48 56 59 86 79 67 45 82 116 100 61 48 46 48 53 50 59 38 38 56 101 102 55 13 10]}
##0159ST=22;CN=2011;PW=123456;MN=8888888826462285;CP=&&DataTime=20211006030527;TEMP-Rtd=22.0;HUMI-Rtd=43.5;PM25-Rtd=2.0;CO2-Rtd=405.0;CH2O-Rtd=0.008;VOC-Rtd=0.052;&&8ef7
61748803
8888888826462285
8888888861748803
##
0159
ST=22;CN=2011;PW=123456;MN=8888888826462285;CP=
&&
DataTime=20211006030527;TEMP-Rtd=22.0;HUMI-Rtd=43.5;PM25-Rtd=2.0;CO2-Rtd=405.0;CH2O-Rtd=0.008;VOC-Rtd=0.052;
&&
8ef7
*/
//232330313539
//53543d32323b434e3d323031313b50573d3132333435363b4d4e3d383838383838383836313734383830333b43503d26264461746154696d653d32303231313030363031353434353b54454d502d5274643d32302e363b48554d492d5274643d34312e343b504d32352d5274643d352e303b434f322d5274643d3339332e303b4348324f2d5274643d302e3030373b564f432d5274643d302e3030383b2626393330380d0a
func (hj *HJ212) Init(name string) error {
	return nil
}

func (hj *HJ212) Uninit() {
}

func (hj *HJ212) ChannelDispatch(b []byte) int {
	if b[0] == 0x23 && b[1] == 0x23 {
		return 1
	}
	return 0
}

func parseField(sfileds string, sinfo *map[string]interface{}, bintv bool, skip ...string) error {
	sTelemtry := strings.Split(sfileds, ";")
	for _, v := range sTelemtry {
		kv := strings.Split(v, "=")
		if len(kv) == 2 && len(kv[1]) > 0 {
			bskiped := false
			for _, s := range skip {
				if s == kv[0] {
					bskiped = true
				}
			}
			if !bskiped {
				v := kv[1]
				if bintv {
					fv, err := strconv.ParseFloat(v, 64)
					if err == nil {
						skey := strings.Split(kv[0], "-")
						(*sinfo)[skey[0]] = fv
						continue
					}
				}
				(*sinfo)[kv[0]] = kv[1]
			}
		}
	}
	return nil
}

func (hj *HJ212) ParsePacket(buff []byte, param ...interface{}) interface{} {
	if buff[0] != 0x23 || buff[1] != 0x23 {
		return nil
	}
	blen := buff[2:6]
	plen, err := strconv.Atoi(string(blen))
	if err != nil {
		return nil
	}
	log.Info(plen)

	data := buff[6:]
	szPayload := string(data)
	sFields := strings.Split(szPayload, "&&")
	if len(sFields) < 1 {
		return nil
	}

	sinfo := make(map[string]interface{})
	parseField(sFields[0], &sinfo, false, "PW")

	if v, ok := sinfo["ST"]; !ok || v != "22" { //ST=22 ==> air qulity
		return nil
	}
	if v, ok := sinfo["CN"]; !ok || v != "2011" { //CN=2011 ==> air 实时分钟数据上传
		return nil
	}
	delete(sinfo, "ST")
	delete(sinfo, "CN")

	parseField(sFields[1], &sinfo, true, "DataTime")
	if len(param) > 0 && reflect.TypeOf(param[0]).Kind() == reflect.Ptr {
		if mn, has := sinfo["MN"]; has {
			*param[0].(*string) = mn.(string)
			delete(sinfo, "MN")
		}
	}
	return sinfo
}

func (hj *HJ212) PackageCmd(cmd string, param ...interface{}) interface{} {
	return nil
}

func NewHJ212() IProtocol {
	p := &HJ212{}
	p.Init("HJ212")
	return p
}

func init() {
	ProtReg["HJ212"] = NewHJ212
}
