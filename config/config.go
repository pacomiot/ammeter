package config

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/yuguorong/go/log"
)

type Configs struct {
	fileds map[string]interface{}
	mutex  sync.Locker
}

func (c *Configs) LoadConfig() {
	c.LoadFromFS()
}

func (c *Configs) Password(passwd string) string {
	data := []byte(passwd)
	has := md5.Sum(data)
	md5str1 := fmt.Sprintf("%x", has) //将[]byte转成16进制
	//log.Info(md5str1)
	return md5str1
}

func (c *Configs) genUUID() {
	f, _ := os.OpenFile("/dev/urandom", os.O_RDONLY, 0)
	b := make([]byte, 16)
	f.Read(b)
	f.Close()
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	log.Info(uuid)
}

//Load Get system debug configuration
func (c *Configs) LoadFromFS() {
	if file, err := os.Open("conf.json"); err == nil {
		defer file.Close()
		var tmp = make([]byte, 1024)
		n, err := file.Read(tmp)
		if err == nil {
			confv := make(map[string]interface{})
			err = json.Unmarshal(tmp[:n], &confv)
			if err == nil {
				c.fileds = confv
			} else {
				log.Error(err)
			}
		}
	}
}

//Update Save sys config
func (c *Configs) Save() {
	f, _ := os.OpenFile("conf.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer f.Close()
	jsonStu, _ := json.Marshal(c.fileds)
	f.WriteString(string(jsonStu))
	log.Info("SysConfig saved!")
}

func (c *Configs) LoadFromCloud() {

}

func (c *Configs) GetValue(fieldPath string, defVal interface{}) interface{} {
	spath := strings.Split(fieldPath, "/")
	isSimpleT := true
	if defVal != nil {
		vt := reflect.TypeOf(defVal)
		if vt.Kind() >= reflect.Array && vt.Kind() != reflect.String {
			isSimpleT = false
		}
	}

	lens := len(spath)
	if lens > 0 {
		lpath := c.fileds
		c.mutex.Lock()
		for i := 0; i < lens-1; i++ {
			gwc, has := (lpath)[spath[i]]
			if !has {
				newp := make(map[string]interface{})
				lpath[spath[i]] = newp
				lpath = newp
			} else {
				lpath = gwc.(map[string]interface{})
			}
		}
		c.mutex.Unlock()

		value, has := lpath[spath[lens-1]]
		if has {
			if isSimpleT {
				return value
			} else {
				json.Unmarshal([]byte(value.(string)), defVal)
				return defVal
			}
		}

	}
	if isSimpleT {
		return defVal
	} else {
		return nil
	}
}

func (c *Configs) SetValue(fieldPath string, value interface{}) bool {
	spath := strings.Split(fieldPath, "/")
	lens := len(spath)
	if lens > 0 {
		c.mutex.Lock()
		defer c.mutex.Unlock()
		lpath := &c.fileds
		for i := 0; i < lens-1; i++ {
			gwc, has := (*lpath)[spath[i]]
			if !has {
				newp := make(map[string]interface{})
				(*lpath)[spath[i]] = newp
				lpath = &newp
			} else {
				v := gwc.(map[string]interface{})
				lpath = &v
			}
		}
		vt := reflect.TypeOf(value)
		if vt.Kind() < reflect.Array || vt.Kind() == reflect.String {
			if v, ok := (*lpath)[spath[lens-1]]; ok {
				if reflect.TypeOf(value) == reflect.TypeOf(v) && reflect.DeepEqual(v, value) {
					return false
				}
			}
			(*lpath)[spath[lens-1]] = value
		} else {
			sval, _ := json.Marshal(value)
			if v, ok := (*lpath)[spath[lens-1]]; ok {
				if v.(string) == string(sval) {
					return false
				}
			}
			(*lpath)[spath[lens-1]] = string(sval)
		}
	}
	return true
}

func (c *Configs) SetProfile(name string, prof interface{}) bool {
	path := "profile/" + name
	if c.SetValue(path, prof) {
		c.Save()
		return true
	}
	return false
}

func (c *Configs) GetProfile(name string, defV interface{}) interface{} {
	path := "profile/" + name
	prof := c.GetValue(path, defV)
	if prof == nil {
		prof = defV
		c.SetValue(path, defV)
		c.Save()
	}

	return prof
}

var sysConfig = &Configs{
	fileds: make(map[string]interface{}),
	mutex:  &sync.Mutex{},
}

func GetSysConfig() *Configs {
	return sysConfig
}

func init() {
	sysConfig.LoadConfig()
}
