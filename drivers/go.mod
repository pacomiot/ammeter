module github.com/ammeter/drviers

go 1.14

replace github.com/ammeter/Bus => ../Bus

replace github.com/ammeter/config => ../config

replace github.com/ammeter/protocol => ../protocol

require (
	github.com/ammeter/Bus v0.0.0-00010101000000-000000000000
	github.com/ammeter/config v0.0.0-00010101000000-000000000000
	github.com/ammeter/protocol v0.0.0-00010101000000-000000000000
	github.com/yuguorong/go v0.0.0-20180604090527-bdc77568d726
)
