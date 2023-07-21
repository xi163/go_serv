package game_login

import (
	"context"
	"net"
	"os"
	"strconv"
	"strings"

	config "github.com/cwloo/server/src/config"
	"github.com/cwloo/server/src/game_login/handler"
	"github.com/cwloo/server/src/global"

	"github.com/xi123/libgo/core/net/conn"
	"github.com/xi123/libgo/logs"
	"github.com/xi123/libgo/utils/cmd"
	"github.com/xi123/libgo/utils/conv"
	getcdv3 "github.com/cwloo/grpc-etcdv3/getcdv3"
	pb_getcdv3 "github.com/cwloo/grpc-etcdv3/getcdv3/proto"
	pb_httpgate "github.com/cwloo/server/proto/gate.http"
	pb_public "github.com/cwloo/server/proto/public"
	"google.golang.org/grpc"
)

// <summary>
// RPCServer
// <summary>
type RPCServer struct {
	addr       string
	port       int
	node       string
	etcdSchema string
	etcdAddr   []string
	target     string
}

func (s *RPCServer) Addr() string {
	return s.addr
}

func (s *RPCServer) Port() int {
	return s.port
}

func (s *RPCServer) Node() string {
	return s.node
}

func (s *RPCServer) Schema() string {
	return s.etcdSchema
}

func (s *RPCServer) Target() string {
	return s.target
}

func (s *RPCServer) Init(id int, name string) {
}

func (s *RPCServer) Run(id int, name string) {
	switch cmd.PatternArg("rpc") {
	case "":
		if id >= len(config.Config.Rpc.HttpGate.Port) {
			logs.Fatalf("error id=%v Rpc.HttpGate.Port.size=%v", id, len(config.Config.Rpc.HttpGate.Port))
		}
		s.addr = config.Config.Rpc.Ip
		s.port = config.Config.Rpc.HttpGate.Port[id]
	default:
		addr := conn.ParseAddress(cmd.PatternArg("rpc"))
		switch addr {
		case nil:
			logs.Fatalf("error")
		default:
			s.addr = addr.Ip
			s.port = conv.StrToInt(addr.Port)
		}
	}
	s.node = config.Config.Rpc.HttpGate.Node
	s.etcdSchema = config.Config.Etcd.Schema
	s.etcdAddr = config.Config.Etcd.Addr
	listener, err := net.Listen("tcp", strings.Join([]string{s.addr, strconv.Itoa(s.port)}, ":"))
	if err != nil {
		logs.Fatalf(err.Error())
	}
	defer listener.Close()
	var opts []grpc.ServerOption
	server := grpc.NewServer(opts...)
	defer server.GracefulStop()
	pb_getcdv3.RegisterPeerServer(server, s)
	pb_public.RegisterPeerServer(server, s)
	pb_httpgate.RegisterHttpGateServer(server, s)
	logs.Warnf("%v:%v etcd%v %v %v:%v:%v", name, id, s.etcdAddr, s.etcdSchema, s.node, s.addr, s.port)
	err = getcdv3.RegisterEtcd(s.etcdSchema, s.node, s.addr, s.port, config.Config.Etcd.Timeout.Keepalive)
	if err != nil {
		errMsg := strings.Join([]string{s.etcdSchema, strings.Join(s.etcdAddr, ","), net.JoinHostPort(s.addr, strconv.Itoa(s.port)), s.node, err.Error()}, " ")
		logs.Fatalf(errMsg)
	}
	s.target = getcdv3.GetUniqueTarget(s.etcdSchema, s.node, s.addr, s.port)
	// logs.Warnf("target=%v", s.target)
	err = server.Serve(listener)
	if err != nil {
		logs.Fatalf(err.Error())
		return
	}
}

func (r *RPCServer) GetRouter(_ context.Context, req *pb_public.RouterReq) (*pb_public.RouterResp, error) {
	logs.Debugf("%v [%v:%v %v:%v rpc:%v:%v NumOfLoads:%v] %+v",
		os.Getpid(),
		global.Name,
		cmd.Id()+1,
		config.Config.HttpGate.Ip, config.Config.HttpGate.Port[cmd.Id()],
		config.Config.Rpc.Ip, config.Config.Rpc.HttpGate.Port[cmd.Id()],
		global.Uploaders.Len(),
		req)
	return &pb_public.RouterResp{}, nil
}

func (r *RPCServer) GetNodeInfo(_ context.Context, req *pb_public.NodeInfoReq) (*pb_public.NodeInfoResp, error) {
	// logs.Debugf("%v [%v:%v %v:%v rpc:%v:%v NumOfLoads:%v] %+v",
	// 	os.Getpid(),
	// 	global.Name,
	// 	cmd.Id()+1,
	// 	config.Config.HttpGate.Ip, config.Config.HttpGate.Port[cmd.Id()],
	// 	config.Config.Rpc.Ip, config.Config.Rpc.HttpGate.Port[cmd.Id()],
	// 	global.Uploaders.Len(),
	// 	req)
	return handler.GetNodeInfo()
}

func (r *RPCServer) GetAddr(_ context.Context, req *pb_getcdv3.PeerReq) (*pb_getcdv3.PeerResp, error) {
	// logs.Debugf("%v [%v:%v %v:%v rpc:%v:%v NumOfLoads:%v] %+v",
	// 	os.Getpid(),
	// 	global.Name,
	// 	cmd.Id()+1,
	// 	config.Config.HttpGate.Ip, config.Config.HttpGate.Port[cmd.Id()],
	// 	config.Config.Rpc.Ip, config.Config.Rpc.HttpGate.Port[cmd.Id()],
	// 	global.Uploaders.Len(),
	// 	req)
	return &pb_getcdv3.PeerResp{Addr: strings.Join([]string{config.Config.Rpc.Ip, strconv.Itoa(config.Config.Rpc.HttpGate.Port[cmd.Id()])}, ":")}, nil
}
