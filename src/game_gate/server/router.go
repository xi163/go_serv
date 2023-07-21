package game_gate

import (
	"net/http"

	"github.com/xi123/libgo/core/net/conn"
	"github.com/xi123/libgo/logs"
	"github.com/xi123/libgo/utils/cmd"
	"github.com/xi123/libgo/utils/conv"
	"github.com/cwloo/server/src/config"
	"github.com/cwloo/server/src/game_gate/handler"
	"github.com/cwloo/server/src/global/httpsrv"
)

// <summary>
// Router
// <summary>
type Router struct {
	server httpsrv.HttpServer
}

func (s *Router) Server() httpsrv.HttpServer {
	return s.server
}

func (s *Router) Init(id int, name string) {
}

func (s *Router) Run(id int, name string) {
	switch cmd.PatternArg("router") {
	case "":
		if id >= len(config.Config.GameGate.Http.Port) {
			logs.Fatalf("error id=%v GameGate.Http.Port.size=%v", id, len(config.Config.GameGate.Http.Port))
		}
		s.server = httpsrv.NewHttpServer(
			config.Config.GameGate.Http.Ip,
			config.Config.GameGate.Http.Port[id],
			config.Config.GameGate.Http.IdleTimeout)
	default:
		addr := conn.ParseAddress(cmd.PatternArg("router"))
		switch addr {
		case nil:
			logs.Fatalf("error")
		default:
			s.server = httpsrv.NewHttpServer(
				addr.Ip,
				conv.StrToInt(addr.Port),
				config.Config.GameGate.Http.IdleTimeout)
		}
	}
	// s.server.Router(config.Config.Path.UpdateCfg, s.UpdateConfigReq)
	// s.server.Router(config.Config.Path.GetCfg, s.GetConfigReq)
	// s.server.Router(config.Config.GameGate.Http.Path.Router, s.RouterReq)
	s.server.Run(id, name)
}

func (s *Router) UpdateConfigReq(w http.ResponseWriter, r *http.Request) {
	handler.UpdateCfgReq(w, r)
}

func (s *Router) GetConfigReq(w http.ResponseWriter, r *http.Request) {
	handler.GetCfgReq(w, r)
}

func (s *Router) RouterReq(w http.ResponseWriter, r *http.Request) {
	handler.RouterReq(w, r)
}
