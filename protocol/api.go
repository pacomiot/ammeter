package protocol

type IProtocol interface {
	Init(name string) error
	Uninit()
	ChannelDispatch(rxb []byte) int
	ParsePacket(buff []byte, param ...interface{}) interface{}
	PackageCmd(cmd string, param ...interface{}) interface{}
}

type funcRegProt func() IProtocol

var ProtList map[string]IProtocol
var ProtReg map[string]funcRegProt

func LoadProtocol(name string) IProtocol {
	if p, has := ProtList[name]; has && p != nil {
		return p
	}
	if f, has := ProtReg[name]; has && f != nil {
		return f()
	}
	return nil
}

func init() {
	ProtList = make(map[string]IProtocol)
	ProtReg = make(map[string]funcRegProt)
}
