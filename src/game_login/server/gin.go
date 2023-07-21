package game_login

import (
	"net"
	"net/http"
	"strconv"

	"github.com/xi123/libgo/core/net/conn"
	"github.com/xi123/libgo/logs"
	"github.com/xi123/libgo/utils/cmd"
	"github.com/xi123/libgo/utils/conv"
	"github.com/cwloo/server/src/config"
	"github.com/cwloo/server/src/game_login/handler/register"
	"github.com/cwloo/server/src/global/httpsrv"
	"github.com/gin-gonic/gin"
)

// <summary>
// GRouter
// <summary>
type GRouter struct {
	server httpsrv.HttpServer
	r      *gin.Engine
	addr   string
	port   int
}

func (s *GRouter) Server() httpsrv.HttpServer {
	return s.server
}

func (s *GRouter) Init(id int, name string) {
	gin.SetMode(gin.ReleaseMode)
	s.r = gin.Default()
	s.r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar") // 跨域关键设置 让浏览器可以解析
		c.Header("Access-Control-Max-Age", "172800")                                                                                                                                                           // 缓存请求信息 单位为秒
		c.Header("Access-Control-Allow-Credentials", "false")                                                                                                                                                  //  跨域请求是否需要带cookie信息 默认设置为true
		c.Header("content-type", "application/json")                                                                                                                                                           // 设置返回格式是json
		//Release all option pre-requests
		if c.Request.Method == http.MethodOptions {
			c.JSON(http.StatusOK, "Options Request!")
		}
		c.Next()
	})
}

func (s *GRouter) Run(id int, name string) {
	switch cmd.PatternArg("server") {
	case "":
		if id >= len(config.Config.GameLogin.Port) {
			logs.Fatalf("error id=%v GameLogin.Port.size=%v", id, len(config.Config.GameLogin.Port))
		}
		s.addr = config.Config.GameLogin.Ip
		s.port = config.Config.GameLogin.Port[id]
	default:
		addr := conn.ParseAddress(cmd.PatternArg("server"))
		switch addr {
		case nil:
			logs.Fatalf("error")
		default:
			s.addr = addr.Ip
			s.port = conv.StrToInt(addr.Port)
		}
	}
	s.register()
	err := s.r.Run(net.JoinHostPort(s.addr, strconv.Itoa(s.port)))
	switch err {
	case nil:
	default:
		logs.Fatalf(err.Error())
	}
}

func (s *GRouter) assert() {
	if s.r == nil {
		logs.Fatalf("error")
	}
}

func (s *GRouter) register() {
	s.assert()
	s.r.POST("/register", register.Register)
	s.r.POST("/login", register.Login)
}
