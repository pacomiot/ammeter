package drviers

type IDriver interface {
	Name() string
	Probe(name string, param interface{})
	CreateDevice(model interface{}) (int, []IDevice)
	GetModel() interface{}
	GetDevice(id string) IDevice
	GetDeviceList() map[string]IDevice
	Uninstall()
}

type baseDriver struct {
	DrvName string
	DevList map[string]IDevice
}

func (drv *baseDriver) Name() string {
	return drv.DrvName
}

func (drv *baseDriver) Probe(name string, param interface{}) {
	drv.DrvName = name
	drv.DevList = make(map[string]IDevice)
}

func (drv *baseDriver) CreateDevice(model interface{}) (int, []IDevice) {
	dev := make([]IDevice, len(drv.DevList))
	i := 0
	for _, v := range drv.DevList {
		dev[i] = v
		i++
	}
	return i, dev
}

func (drv *baseDriver) GetModel() interface{} {
	return nil
}

func (drv *baseDriver) GetDevice(id string) IDevice {
	if dev, has := drv.DevList[id]; has {
		return dev
	}
	return nil
}

func (drv *baseDriver) GetDeviceList() map[string]IDevice {
	return drv.DevList
}

func (drv *baseDriver) Uninstall() {
	for _, dev := range drv.DevList {
		dev.Close()
	}
	drv.DevList = nil
}

type funcRegDriver func(interface{}) IDriver

var driverReg map[string]funcRegDriver
var driverList map[string]IDriver

func Shutdown() {
	for _, drv := range driverList {
		drv.Uninstall()
	}
}

func Install(name string, param interface{}) IDriver {
	if p, has := driverList[name]; has && p != nil {
		return p
	}
	if f, has := driverReg[name]; has && f != nil {
		drv := f(param)
		drv.Probe(name, param)
		return drv
	}
	return nil
}

func init() {
	driverList = make(map[string]IDriver)
	driverReg = make(map[string]funcRegDriver)
}
