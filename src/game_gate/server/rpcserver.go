package game_gate

import (
	"context"
	"net"
	"os"
	"strconv"
	"strings"

	config "github.com/cwloo/server/src/config"
	"github.com/cwloo/server/src/game_gate/handler"
	"github.com/cwloo/server/src/global"

	"github.com/cwloo/gonet/core/net/conn"
	"github.com/cwloo/gonet/logs"
	"github.com/cwloo/gonet/utils/cmd"
	"github.com/cwloo/gonet/utils/conv"
	getcdv3 "github.com/cwloo/grpc-etcdv3/getcdv3"
	pb_getcdv3 "github.com/cwloo/grpc-etcdv3/getcdv3/proto"
	pb_gamegate "github.com/cwloo/server/proto/game.gate"
	pb_public "github.com/cwloo/server/proto/public"
	"google.golang.org/grpc"
	//promePkg "github.com/cwloo/server/src/global/pkg/common/prometheus"
	//grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
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
		if id >= len(config.Config.Rpc.GameGate.Port) {
			logs.Fatalf("error id=%v Rpc.GameGate.Port.size=%v", id, len(config.Config.Rpc.GameGate.Port))
		}
		s.addr = config.Config.Rpc.Ip
		s.port = config.Config.Rpc.GameGate.Port[id]
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
	s.node = config.Config.Rpc.GameGate.Node
	s.etcdSchema = config.Config.Etcd.Schema
	s.etcdAddr = config.Config.Etcd.Addr
	listener, err := net.Listen("tcp", net.JoinHostPort(s.addr, strconv.Itoa(s.port)))
	if err != nil {
		logs.Fatalf(err.Error())
	}
	defer listener.Close()
	var opts []grpc.ServerOption
	if config.Config.Prometheus.Enable {
		// promePkg.NewMsgRecvTotalCounter()
		// promePkg.NewGetNewestSeqTotalCounter()
		// promePkg.NewPullMsgBySeqListTotalCounter()
		// promePkg.NewMsgOnlinePushSuccessCounter()
		// promePkg.NewOnlineUserGauges()
		// promePkg.NewSingleChatMsgRecvSuccessCounter()
		// promePkg.NewGroupChatMsgRecvSuccessCounter()
		// promePkg.NewWorkSuperGroupChatMsgRecvSuccessCounter()
		// promePkg.NewGrpcRequestCounter()
		// promePkg.NewGrpcRequestFailedCounter()
		// promePkg.NewGrpcRequestSuccessCounter()
		opts = append(opts, []grpc.ServerOption{
			// grpc.UnaryInterceptor(promePkg.UnaryServerInterceptorProme),
			// grpc.StreamInterceptor(grpcPrometheus.StreamServerInterceptor),
			// grpc.UnaryInterceptor(grpcPrometheus.UnaryServerInterceptor),
		}...)
	}
	server := grpc.NewServer(opts...)
	defer server.GracefulStop()
	pb_getcdv3.RegisterPeerServer(server, s)
	pb_public.RegisterPeerServer(server, s)
	pb_gamegate.RegisterGameGateServer(server, s)
	// logs.Warnf("%v:%v etcd%v %v %v:%v:%v", name, id, s.etcdAddr, s.etcdSchema, s.node, s.addr, s.port)
	logs.Warnf("%v %v etcd%v %v %v", name, net.JoinHostPort(s.addr, conv.IntToStr(s.port)), s.etcdAddr, s.etcdSchema, s.node)
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
		config.Config.GameGate.Ip, config.Config.GameGate.Port[cmd.Id()],
		config.Config.Rpc.Ip, config.Config.Rpc.GameGate.Port[cmd.Id()],
		global.Uploaders.Len(),
		req)
	return &pb_public.RouterResp{}, nil
}

func (r *RPCServer) GetNodeInfo(_ context.Context, req *pb_public.NodeInfoReq) (*pb_public.NodeInfoResp, error) {
	logs.Debugf("%v [%v:%v %v:%v rpc:%v:%v NumOfLoads:%v] %+v",
		os.Getpid(),
		global.Name,
		cmd.Id()+1,
		config.Config.GameGate.Ip, config.Config.GameGate.Port[cmd.Id()],
		config.Config.Rpc.Ip, config.Config.Rpc.GameGate.Port[cmd.Id()],
		global.Uploaders.Len(),
		req)
	return handler.GetNodeInfo()
}

func (r *RPCServer) GetAddr(_ context.Context, req *pb_getcdv3.PeerReq) (*pb_getcdv3.PeerResp, error) {
	logs.Debugf("%v [%v:%v %v:%v rpc:%v:%v NumOfLoads:%v] %+v",
		os.Getpid(),
		global.Name,
		cmd.Id()+1,
		config.Config.GameGate.Ip, config.Config.GameGate.Port[cmd.Id()],
		config.Config.Rpc.Ip, config.Config.Rpc.GameGate.Port[cmd.Id()],
		global.Uploaders.Len(),
		req)
	return &pb_getcdv3.PeerResp{Addr: strings.Join([]string{config.Config.Rpc.Ip, strconv.Itoa(config.Config.Rpc.GameGate.Port[cmd.Id()])}, ":")}, nil
}

func (r *RPCServer) GetGameGate(_ context.Context, req *pb_gamegate.GameGateReq) (*pb_gamegate.GameGateRsp, error) {
	logs.Warnf(config.Config.GameGate.Path.Handshake)
	return &pb_gamegate.GameGateRsp{
		NumOfLoads: int32(global.Server.(*Server).NumOfLoads()),
		Host:       net.JoinHostPort(config.Config.GameGate.Ip, conv.IntToStr(config.Config.GameGate.Port[cmd.Id()])),
		Domain: strings.Join([]string{config.Config.GameGate.Proto, "://",
			net.JoinHostPort(config.Config.GameGate.Ip, strconv.Itoa(config.Config.GameGate.Port[cmd.Id()])),
			config.Config.GameGate.Path.Handshake}, ""),
	}, nil
}

// 用户多端或异地登陆检查
// func (r *RPCServer) MultiLoginCheck(ctx context.Context, req *pb_gamegate.MultiLoginCheckReq) (*pb_gamegate.MultiLoginCheckRsp, error) {
// 	logs.Debugf("userid=%v platformID=%d sessionID=%v", req.UserID, int(req.PlatformID), req.SessionID)
// 	global.Server.(*Server).KickUserCheckBy(req.UserID, int(req.PlatformID), req.SessionID, req.Token)
// 	return &pb_gamegate.MultiLoginCheckRsp{}, nil
// }

// 用户状态变更通知
// func (r *RPCServer) UserStatusChanged(_ context.Context, req *pb_gamegate.UserStatusChangedReq) (*pb_gamegate.UserStatusChangedRsp, error) {

// 	return nil, nil
// }

// 查询用户状态
// func (r *RPCServer) GetUsersStatus(_ context.Context, req *pb_gamegate.UsersStatusReq) (*pb_gamegate.UsersStatusRsp, error) {
// 	return nil, nil
// }

// 踢人下线
// func (r *RPCServer) KickUserOffline(_ context.Context, req *pb_gamegate.KickUserReq) (*pb_gamegate.KickUserRsp, error) {
// 	// logs.Debugf("")
// 	//
// 	//	for _, v := range req.KickUserIDList {
// 	//		SetTokenKicked(v, int(req.PlatformID), req.OperationID)
// 	//		if platforms := ws.GetUserConns(v); platforms != nil {
// 	//			platforms.Range(func(platformId int, sessions *ws_session.SessionToConn) {
// 	//				sessions.Range(func(sessionId string, peer conn.Session) {
// 	//					ws.sendKickMsg(peer)
// 	//					peer.CloseAfter(time.Second)
// 	//				})
// 	//			})
// 	//		}
// 	//	}
// 	//
// 	// return &pb_gamegate.KickUserOfflineResp{}, nil
// 	return nil, nil
// }
