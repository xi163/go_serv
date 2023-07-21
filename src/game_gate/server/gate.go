package game_gate

import (
	"sync"

	"github.com/cwloo/gonet/logs"
	"github.com/cwloo/server/src/config"
	"github.com/cwloo/server/src/global"
	//promePkg "github.com/cwloo/server/src/global/pkg/common/prometheus"
)

var (
	wg sync.WaitGroup
)

func RunPromethues(id int) {
	promethuesPort := func(id int) (promethuesPort int) {
		if config.Config.Prometheus.Enable {
			if len(config.Config.Prometheus.GameGate.Port) > 0 {
				if id >= len(config.Config.Prometheus.GameGate.Port) {
					logs.Fatalf("error id=%v Prometheus.GameGate.Port.size=%v", id, len(config.Config.Prometheus.GameGate.Port))
				}
				promethuesPort = config.Config.Prometheus.GameGate.Port[id]
			}
		}
		return
	}(id)
	if promethuesPort > 0 {
		go func() {
			// err := promePkg.StartPromeSrv(promethuesPort)
			// if err != nil {
			// 	panic(err)
			// }
		}()
	}
}

func Run(id int, name string) {
	// rwLock = new(sync.RWMutex)
	// validate = validator.New()
	// statistics.NewStatistics(&sendMsgAllCount, config.Config.ModuleName.LongConnSvrName, fmt.Sprintf("%d second recv to msg_gateway sendMsgCount", constant.StatisticsTimeInterval), constant.StatisticsTimeInterval)
	// statistics.NewStatistics(&userCount, config.Config.ModuleName.LongConnSvrName, fmt.Sprintf("%d second add user conn", constant.StatisticsTimeInterval), constant.StatisticsTimeInterval)
	// global.Router = &Router{}
	global.Server = &Server{}
	global.RpcServer = &RPCServer{}
	// global.Router.Init(id, name)
	global.Server.Init(id, name)
	global.RpcServer.Init(id, name)
	wg.Add(2)
	RunPromethues(id)
	// go global.Router.Run(id, name)
	go global.Server.Run(id, name)
	go global.RpcServer.Run(id, name)
	wg.Wait()
}
