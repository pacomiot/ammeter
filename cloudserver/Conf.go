package cloudserver

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"os"

	"github.com/yuguorong/go/log"
)

type CloudServerConf struct {
	ServerUrl string
	AppId     string
	AppSecret string
	Route     string
}

var defConf = &CloudServerConf{}

func (c *CloudServerConf) GenerateSignature() string {
	res := md5V(c.AppId + "-pacom-" + c.AppSecret)
	return res
}

func md5V(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func GetCloudConfig() interface{} {
	return defConf
}

func init() {
	if file, err := os.Open("cloudconf.json"); err == nil {
		defer file.Close()
		var tmp = make([]byte, 1024)
		n, err := file.Read(tmp)
		if err == nil {
			err = json.Unmarshal(tmp[:n], &defConf)
			if err != nil {
				log.Error(err)
			}
		}
	}
}
