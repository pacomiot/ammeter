module github.com/ammeter/cmd

go 1.14

replace github.com/ammeter/platform => ../platform

replace github.com/ammeter/cloudserver => ../cloudserver

replace github.com/ammeter/config => ../config

replace github.com/ammeter/drivers => ../drivers

replace github.com/ammeter/protocol => ../protocol

replace github.com/ammeter/Bus => ../Bus

require (
	github.com/ammeter/cloudserver v0.0.0-00010101000000-000000000000
	github.com/ammeter/config v0.0.0-00010101000000-000000000000
	github.com/ammeter/platform v0.0.0-00010101000000-000000000000
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/yuguorong/go v0.0.0-20180604090527-bdc77568d726
)
