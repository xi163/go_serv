package main

import (
	"math/rand"
	"time"

	"github.com/xi123/libgo/logs"
	"github.com/xi123/libgo/utils/cmd"
	"github.com/xi123/server/src/config"
	gate "github.com/xi123/server/src/game_gate/server"
	"github.com/xi123/server/src/global"
)

func init() {
	cmd.InitArgs(func(arg *cmd.ARG) {
		arg.SetConf("config/conf.ini")
		arg.AppendPattern("server", "server", "srv", "svr", "s")
		arg.AppendPattern("router", "router", "httpserver")
		arg.AppendPattern("rpc", "rpc", "r")
	})
}

// game_gate.exe --dir-level=3 --conf-name=deploy\config\conf.ini --log-dir=E:\winshare\server/deploy/log --s=ws://192.168.0.105:7785/ws_7785 --r=192.168.0.105:3556
func main() {
	cmd.ParseArgs()
	config.InitGameGateConfig(cmd.Conf())
	logs.SetTimezone(logs.Timezone(config.Config.Log.GameGate.Timezone))
	logs.SetMode(logs.Mode(config.Config.Log.GameGate.Mode))
	logs.SetStyle(logs.Style(config.Config.Log.GameGate.Style))
	logs.SetLevel(logs.Level(config.Config.Log.GameGate.Level))
	logs.Init(config.Config.Log.GameGate.Dir, global.Exe, 100000000)
	rand.Seed(time.Now().UnixNano())
	// task.After(time.Duration(config.Config.Interval+(rand.Int()%50))*time.Second, cb.NewFunctor00(func() {
	// 	handler.ReadConfig()
	// }))
	gate.Run(cmd.Id(), config.ServiceName())
	logs.Close()
}
