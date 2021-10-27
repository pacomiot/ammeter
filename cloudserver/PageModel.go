package cloudserver

import (
	"encoding/json"
	"errors"

	"github.com/yuguorong/go/log"
)

type PageModel struct {
	Rows  []interface{} `json:"rows"`
	Total int           `json:"total"`
}

func (p *PageModel) ChangeData(resPoint interface{}) error {
	if len(p.Rows) == 0 {
		return nil
	}
	b, e := json.Marshal(p.Rows)
	log.Info("data", string(b))
	if e != nil {
		return e
	}
	e = json.Unmarshal(b, resPoint)
	return e
}

type CommonResp struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

func (c *CommonResp) ChangeData(resPoint interface{}) error {
	if c.Data == nil {
		return errors.New("data is empty")
	}

	b, e := json.Marshal(c.Data)
	if e != nil {
		return e
	}
	e = json.Unmarshal(b, resPoint)
	return e
}
