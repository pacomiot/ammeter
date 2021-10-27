package cloudserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	"github.com/yuguorong/go/log"
)

type CloudServer struct {
	conf       *CloudServerConf
	httpClient *http.Client
	Name       string
}

func GetCloudServer(profile interface{}) *CloudServer {
	pfofileCS := profile.(*CloudServerConf)
	res := &CloudServer{conf: pfofileCS}
	res.httpClient = &http.Client{Timeout: 5 * time.Second}
	return res
}

func (c *CloudServer) GetClientData(infourl string, pData interface{}, param *map[string]string) error {
	pNum, psize := 1, 100

	reqParams := make(map[string]interface{})
	reqParams["pageSize"] = psize
	if param != nil {
		for k, v := range *param {
			reqParams[k] = v
		}
	}

	pin := reflect.ValueOf(pData).Elem()
	tin := reflect.TypeOf(pData).Elem()
	sliceret := reflect.MakeSlice(tin, 0, 0)

	itemCount := 0
	for {
		slicetmp := reflect.MakeSlice(tin, 0, 0)
		pin.Set(slicetmp)
		reqParams["pageNum"] = pNum
		p, e := c.requestPage(infourl, reqParams)
		pNum += 1
		if e != nil {
			log.Info("query ammeter error", e)
			break
		}

		if len(p.Rows) == 0 {
			break
		}

		e = p.ChangeData(pData)
		if e != nil {
			log.Info("change data error", e)
			continue
		}

		vin := reflect.ValueOf(pData).Elem()
		sliceret = reflect.AppendSlice(sliceret, vin)

		itemCount += vin.Len()
		if itemCount >= p.Total {
			vin.Set(sliceret)
			break
		}
	}

	return nil
}

func (c *CloudServer) requestPage(url string, data interface{}) (p *PageModel, erro error) {
	url = c.conf.ServerUrl + url
	httpClient := c.httpClient

	payload, erro := json.Marshal(data)
	if erro != nil {
		return
	}
	log.Info("reqParam", string(payload), url, time.Now())
	reqBody := bytes.NewReader(payload)
	req, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		erro = err
		return
	}

	req.Header.Add("appId", c.conf.AppId)
	req.Header.Add("signature", c.conf.GenerateSignature())
	req.Header.Add("Content-Type", "application/json")

	resp, err := httpClient.Do(req)

	if err != nil {
		erro = err
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		erro = fmt.Errorf("respose statuscode is not equal 200")
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		erro = err
		return
	}

	var jbody CommonResp
	err = json.Unmarshal(body, &jbody)
	if err != nil {
		erro = err
		return
	}
	if jbody.Code != 0 {
		erro = fmt.Errorf("response error %s", jbody.Msg)
		return
	}

	pageModel := &PageModel{}
	erro = jbody.ChangeData(pageModel)
	p = pageModel
	return
}
