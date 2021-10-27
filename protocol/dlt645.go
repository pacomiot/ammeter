package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strconv"

	"github.com/yuguorong/go/log"
)

/*
		"9521003872": {"2108", "05420005", "05420003", "05420002"}, //6f
		"9521003712": {"2108", "05420001"},                         //1F
		"9521003534": {"2108", "05420006"},                         //-1F
		"9521003697": {"2108", "05420004"},                         //5c

//fefefefe68040042050821681104333333332516
//fefefefe68 04 00 42 05 08 21 68910833333333b4bc34338016

data [{"accessToken":"AZM7QSPRqsV4CUxIGldn",
"id":"1ec23004b88a020bf52b38d8579115b",
"mqttServer":"test-sbuilding.pacom.cn:1885",
"name":"FGGW10-202110020001"}]
reqParam {"gwDevId":"1ec23004b88a020bf52b38d8579115b","pageNum":1,"pageSize":100} https://test-admin.pacom.cn/platform/dev/get-ammeter-list 2021-10-03 23:10:26.175516015 +0800 CST m=+7.127250782
data [
	{"address":"05420005",
	"code":"2108",
	"devName":"AM10-9521003872210805",
	"dutSn":"9521003872",
	"gwDevId":"1ec23004b88a020bf52b38d8579115b",
	"id":"1ec2300601517d0bf52b38d8579115b",
	"protocol":"DLT645-2007"},
	{"address":"05420001","code":"2108","devName":"AM10-9521003712210805","dutSn":"9521003712","gwDevId":"1ec23004b88a020bf52b38d8579115b","id":"1ec230092bb8de0bf52b38d8579115b","protocol":"DLT645-2007"},{"address":"05420006","code":"2108","devName":"AM10-9521003534210805","dutSn":"9521003534","gwDevId":"1ec23004b88a020bf52b38d8579115b","id":"1ec23009ed76090bf52b38d8579115b","protocol":"DLT645-2007"},{"address":"05420004","code":"2108","devName":"AM10-9521003697210805","dutSn":"9521003697","gwDevId":"1ec23004b88a020bf52b38d8579115b","id":"1ec2300abbc2890bf52b38d8579115b","protocol":"DLT645-2007"}]


FE FE FE FE
0	68H	帧起始符
1	A0	地址域
2	A1
3	A2
4	A3
5	A4
6	A5
7	68H	帧起始符
8	C	控制码
9	L	数据域长度
10	DATA	数据域
11	CS	校验码
*/

const (
	C_CODE_MASK = 0x1F //功能码标志位

	//功能码
	C_2007_CODE_BRC = 0x08 //广播校时
	C_2007_CODE_RD  = 0X11 //读数据
	C_2007_CODE_RDM = 0x12 //读后续数据
	C_2007_CODE_RDA = 0x13 //读通信地址
	C_2007_CODE_WR  = 0x14 //写数据
	C_2007_CODE_WRA = 0x15 //写通信地址
	C_2007_CODE_DJ  = 0x16 //冻结
	C_2007_CODE_BR  = 0x17 //更改通信速率
	C_2007_CODE_PD  = 0x18 //修改密码
	C_2007_CODE_XL  = 0x19 //最大需量清零
	C_2007_CODE_DB  = 0x1a //电表清零
	C_2007_CODE_MSG = 0x1b //事件清零

	ERR_2007_RATE  = 0x40 //费率数超
	ERR_2007_DAY   = 0x20 //日时段数超
	ERR_2007_YEAR  = 0x10 //年时区数超
	ERR_2007_BR    = 0x08 //通信速率不能更改
	ERR_2007_PD    = 0x04 //密码错误/未授权
	ERR_2007_DATA  = 0x02 //无请求数据
	ERR_2007_OTHER = 0x01 //其他错误

	/*DLT 645 2007数据标识*/
	DLT_GROUP_ACTIVE_POWER_TOTAL         = 0x0       //组合有功总电能
	DLT_PRIC1_GROUP_ACTIVE_POWER_TOTAL   = 0x100     //组合有功费率1电能
	DLT_PRIC2_GROUP_ACTIVE_POWER_TOTAL   = 0x200     //组合有功费率2电能
	DLT_PRIC3_GROUP_ACTIVE_POWER_TOTAL   = 0x300     //组合有功费率3电能
	DLT_PRIC4_GROUP_ACTIVE_POWER_TOTAL   = 0x400     //组合有功费率4电能
	DLT_GROUP_FORTH_POWER_TOTAL          = 0x10000   //正向有功总电能
	DLT_PRIC1_GROUP_FORTH_POWER_TOTAL    = 0x10100   //正向有功费率1电能
	DLT_PRIC2_GROUP_FORTH_POWER_TOTAL    = 0x10200   //正向有功费率2电能
	DLT_PRIC3_GROUP_FORTH_POWER_TOTAL    = 0x10300   //正向有功费率3电能
	DLT_PRIC4_GROUP_FORTH_POWER_TOTAL    = 0x10400   //正向有功费率4电能
	DLT_GROUP_BACK_POWER_TOTAL           = 0x20000   //反向有功总电能
	DLT_PRIC1_GROUP_BACK_POWER_TOTAL     = 0x20100   //反向有功费率1电能
	DLT_PRIC2_GROUP_BACK_POWER_TOTAL     = 0x20200   //反向有功费率2电能
	DLT_PRIC3_GROUP_BACK_POWER_TOTAL     = 0x20300   //反向有功费率3电能
	DLT_PRIC4_GROUP_BACK_POWER_TOTAL     = 0x20400   //反向有功费率4电能
	DLT_GROUP_NONE1_POWER_TOTAL          = 0x30000   //组合无功1总电能
	DLT_GROUP_NONE2_POWER_TOTAL          = 0x40000   //组合无功2总电能
	DLT_GROUP_QUAD1_NONE_POWER_TOTAL     = 0x50000   //第一象限无功电能
	DLT_GROUP_QUAD2_NONE_POWER_TOTAL     = 0x60000   //第二象限无功电能
	DLT_GROUP_QUAD3_NONE_POWER_TOTAL     = 0x70000   //第三象限无功电能
	DLT_GROUP_QUAD4_NONE_POWER_TOTAL     = 0x80000   //第四象限无功电能
	DLT_GROUP_FORTH_APPARENT_POWER_TOTAL = 0x90000   //正向视在总电能
	DLT_PHASE_A_VOLTAGE                  = 0x2010100 //A相电压
	DLT_PHASE_B_VOLTAGE                  = 0x2010200 //B相电压
	DLT_PHASE_C_VOLTAGE                  = 0x2010300 //C相电压
	DLT_PHASE_A_CURENT                   = 0x2020100 //A相电流
	DLT_PHASE_B_CURENT                   = 0x2020200 //B相电流
	DLT_PHASE_C_CURENT                   = 0x2020300 //C相电流
	DIC_2030000                          = 0x2030000 //总有功功率
	DIC_2030100                          = 0x2030100 //A相有功功率
	DIC_2030200                          = 0x2030200 //B相有功功率
	DIC_2030300                          = 0x2030300 //C相有功功率
	DIC_2040000                          = 0x2040000 //总无功功率
	DIC_2040100                          = 0x2040100 //A相无功功率
	DIC_2040200                          = 0x2040200 //B相无功功率
	DIC_2040300                          = 0x2040300 //C相无功功率
	DIC_2050000                          = 0x2050000 //总视在功率
	DIC_2050100                          = 0x2050100 //A相视在功率
	DIC_2050200                          = 0x2050200 //B相视在功率
	DIC_2050300                          = 0x2050300 //C相视在功率
	DIC_2060000                          = 0x2060000 //总功率因素
	DIC_2060100                          = 0x2060100 //A相功率因素
	DIC_2060200                          = 0x2060200 //B相功率因素
	DIC_2060300                          = 0x2060300 //C相功率因素
	DIC_20C0100                          = 0x20C0100 //AB线电压
	DIC_20C0200                          = 0x20C0200 //BC线电压
	DIC_20C0300                          = 0x20C0300 //CA线电压
	DIC_2800002                          = 0x2800002 //频率
	DIC_4000101                          = 0x4000101 //年月日星期
	DIC_4000102                          = 0x4000102 //时分秒
	DIC_5060001                          = 0x5060001 //上一次日冻结时间
	DIC_5060101                          = 0x5060101 //上一次日冻结正向有功电能
	DIC_30C0000                          = 0x30C0000 //过流总次数，总时间
	DIC_30C0101                          = 0x30C0101 //上一次A相过流记录
	DIC_3300100                          = 0x3300100 //电表清零总次数
	DIC_3300101                          = 0x3300101 //电表清零记录
	DIC_4000501                          = 0x4000501
	DIC_4000502                          = 0x4000502
	DIC_4000503                          = 0x4000503
	DIC_4000504                          = 0x4000504
	DIC_4000505                          = 0x4000505
	DIC_4000506                          = 0x4000506
	DIC_4000507                          = 0x4000507
	DIC_4000403                          = 0x4000403 //资产管理编码
	DIC_4000701                          = 0x4000701 //信号强度
	DIC_4000702                          = 0x4000702 //版本号
	DIC_7000001                          = 0x7000001 //master_api_key
	DIC_7000002                          = 0x7000002 //device_id
)

const (
	DLT_CHAR_HEADER      = 0xFE
	DLT_CHAR_TAIL        = 0x16
	DLT_CHAR_FRAME_START = 0x68
	DLT_DEF_CHAR_OPCODE  = C_2007_CODE_RD
	DLT_ADDR_POS         = 1
	DLT_FILED_POS        = 7
	DLT_OPCODE_POS       = 8
	DLT_DICCODE_POS      = 9
	DLT_DICCODE_LEN      = 4
	DLT_PARAM_POS        = 9
	DLT_ADDR_LEN         = 6
	DLT_LEAD_LEN         = 4
	DLT_CMD_DIR_REVERT   = 0x80
)

/*FE FE FE FE
0	68H	帧起始符
1	A0	地址域
2	A1
3	A2
4	A3
5	A4
6	A5
7	68H	帧起始符
8	C	控制码
9	L	数据域长度
10	DATA	数据域
11	CS	校验码
*/
type dlt645 struct {
	name string
}

func bcd2int(bcd []byte) int64 {
	vi := int64(0)
	for i := 0; i < len(bcd); i++ {
		vi = (vi * 10) + int64((bcd[i]>>4)&0xF)
		vi = (vi * 10) + int64(bcd[i]&0xF)
	}
	return vi
}

func (p *dlt645) setDltAddr(buff []byte, amaddr string) error {
	addr, err := hex.DecodeString(amaddr)
	if err == nil {
		len := len(addr)
		for i := 0; i < len/2; i++ {
			j := len - i - 1
			addr[i], addr[j] = addr[j], addr[i]
		}
		copy(buff, addr)
	}
	return err
}

func (p *dlt645) setDltOpcode(buff []byte, opcode string) error {
	opc, err := hex.DecodeString(opcode)
	if err == nil {
		for i := 0; i < len(opc); i++ {
			buff[i+1] = opc[i] + 0x33
		}
		buff[0] = byte(len(opc))
	}
	return err
}

func (p *dlt645) setDltCheckSum(buff []byte) {
	sum := byte(0)
	pos := 0
	len := len(buff) - 1
	for ; pos < len && buff[pos] == 0xFE; pos++ {
	}
	for ; pos < len; pos++ {
		sum += buff[pos]
	}
	buff[len] = sum
}

func (p *dlt645) PackageCmd(opcode string, param ...interface{}) interface{} {
	amaddr := param[0].(string)
	buff := []byte{0xFE, 0xFE, 0xFE, 0xFE, DLT_CHAR_FRAME_START, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, DLT_CHAR_FRAME_START, DLT_DEF_CHAR_OPCODE, 0x00, 0x00, 0x00, 0x00, 0x00, 0X00}
	pkPos := DLT_LEAD_LEN + DLT_ADDR_POS
	err := p.setDltAddr(buff[pkPos:pkPos+DLT_ADDR_LEN], amaddr)
	if err != nil {
		log.Error("error in set meter addr ", amaddr)
		return nil
	}
	parampos := pkPos + DLT_ADDR_LEN + 2
	err = p.setDltOpcode(buff[parampos:parampos+5], opcode)
	if err != nil {
		log.Error("error in set op code ", opcode)
		return nil
	}
	p.setDltCheckSum(buff)
	buff = append(buff, 0x16)
	log.Infof("Send cmd [%s] to dtu[%s]\n", hex.EncodeToString(buff), amaddr)
	return buff
}

func Assert(est bool, infoi int) {
	if !est {
		_, file, line, ok := runtime.Caller(1)
		var serror string
		if ok {
			serror = fmt.Sprintf("Assert at [%s:%d]: (%d)", file, line, infoi)
		} else {
			serror = "parsecmd error with info:" + "" + strconv.Itoa(infoi)
		}
		panic(errors.New(serror))
	}
}

func (p *dlt645) skipLeader(buf []byte) (int, int) {
	pos := 0
	lenb := len(buf)
	for ; pos < lenb && (buf[pos] == 0xFF || buf[pos] == DLT_CHAR_HEADER); pos++ {
	}
	return pos, lenb
}

//fefefefe68040042050821681104333333332516
//fefefefe6804004205082168910833333333b4bc34338016
//fefefefe68 060042050821 68 91 08 33 33 33 33 47 c5 33 33 1d |16
func (p *dlt645) ParsePacket(rxb []byte, params ...interface{}) (mapv interface{}) {
	if len(params) <= 0 || len(rxb) < 12 {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			log.Infof("%s\n", err)
		}

	}()

	sum := byte(0)
	pos, lenb := p.skipLeader(rxb)
	Assert(rxb[pos] == DLT_CHAR_FRAME_START, int(rxb[pos]))
	for i := pos; i < lenb-2; i++ {
		sum += rxb[i]
	}
	Assert(rxb[lenb-1-1] == sum, int(sum))

	//get addr
	addr := rxb[pos+1 : pos+1+DLT_ADDR_LEN]
	for i := 0; i < DLT_ADDR_LEN/2; i++ {
		addr[i], addr[DLT_ADDR_LEN-i-1] = addr[DLT_ADDR_LEN-i-1], addr[i]
	}
	saddr := hex.EncodeToString(addr)
	if len(params) > 0 && reflect.TypeOf(params[0]).Kind() == reflect.Ptr {
		*params[0].(*string) = saddr
	}

	Assert(rxb[pos+DLT_FILED_POS] == DLT_CHAR_FRAME_START, int(rxb[pos+DLT_FILED_POS]))
	//CMD. opcode
	cmd := rxb[pos+DLT_OPCODE_POS]
	Assert(cmd == (C_2007_CODE_RD|DLT_CMD_DIR_REVERT), int(rxb[pos+DLT_DICCODE_POS]))

	//PARAM = DIC_CODE + PARAM
	pos += DLT_DICCODE_POS
	lenParam := int(rxb[pos])
	pos++

	Assert(lenParam > DLT_DICCODE_LEN, lenParam)

	dicCode := 0
	dicBuff := bytes.NewBuffer(rxb[pos : pos+DLT_DICCODE_LEN])
	binary.Read(dicBuff, binary.LittleEndian, &dicCode)

	pos += DLT_DICCODE_LEN

	lenParam = lenParam - DLT_DICCODE_LEN
	Assert(lenParam+pos < lenb-1, lenParam)
	param := make([]byte, lenParam)
	for i := 0; i < lenParam; i++ {
		param[lenParam-i-1] = rxb[pos+i] - 0x33
		rxb[pos+i] = 0
	}
	pos += lenParam

	Assert(rxb[pos+1] == DLT_CHAR_TAIL, int(rxb[pos+1]))

	val := float64(bcd2int(param)) / 100.0
	Assert(val != 0, 0)
	for k, v := range Cmd2Code {
		if int(v) == dicCode {
			mapv := make(map[string]interface{})
			mapv[k] = val
			return mapv
		}
	}
	return nil
}

func (p *dlt645) VisableData(name string, opcode string, param interface{}) string {
	return ""
}

func (p *dlt645) Init(name string) error {
	p.name = name
	ProtList[p.name] = p
	return nil
}

func (p *dlt645) Uninit() {
}

func (p *dlt645) ChannelDispatch(rxb []byte) int {
	return 0
}

func NewDlt645() IProtocol {
	p := &dlt645{}
	p.Init("DLT645-2007")
	return p
}

var Cmd2Code map[string]uint32

func init() {
	ProtReg["DLT645-2007"] = NewDlt645
	Cmd2Code = map[string]uint32{
		"TotalActivePower": DLT_GROUP_ACTIVE_POWER_TOTAL,
	}
}
