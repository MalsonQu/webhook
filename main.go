package main

import (
	"github.com/MalsonQu/webhook/server"
	_ "github.com/MalsonQu/webhook/utils/base"
)

func main() {
	var _server server.Server
	_server.Start()
}
