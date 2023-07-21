package game_login

import (
	"sync"

	"github.com/cwloo/server/src/global"
)

var (
	wg sync.WaitGroup
)

func Run(id int, name string) {
	global.Router = &GRouter{}
	// global.RpcServer = &RPCServer{}
	global.Router.Init(id, name)
	// global.RpcServer.Init(id, name)
	wg.Add(2)
	go global.Router.Run(id, name)
	// go global.RpcServer.Run(id, name)
	wg.Wait()
}
