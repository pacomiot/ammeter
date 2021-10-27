package config

import (
	"flag"
	"net/http"

	"github.com/yuguorong/go/log"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8990", "http config address")

var upgrader = websocket.Upgrader{
	// 解决跨域问题
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} // use default options

func ws(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Error("read:", err)
			break
		}
		log.Infof("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Info("write:", err)
			break
		}
	}
}

func WebsocketEntry() {
	flag.Parse()
	http.HandleFunc("/ws", ws)
	log.Info(*addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
