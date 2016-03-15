package common

import (
	"github.com/codegangsta/cli"
)

var (
	FlAddr = cli.StringFlag{
		Name:  "addr",
		Usage: "<ip>:<port> to listen on",
		Value: "127.0.0.1:8101",
	}

	FlZkAddr = cli.StringFlag{
		Name:  "zk-addr",
		Usage: "<ip>[:<port>] to bind to",
		Value: "127.0.0.1:2181",
	}
)
