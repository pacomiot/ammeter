module github.com/ammeter/Bus

go 1.14

replace github.com/ammeter/platform => ../platform

replace github.com/ammeter/config => ../config

require (
	github.com/ammeter/config v0.0.0-00010101000000-000000000000
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/siddontang/go v0.0.0-20180604090527-bdc77568d726
	github.com/yuguorong/go v0.0.0-20180604090527-bdc77568d726
)
