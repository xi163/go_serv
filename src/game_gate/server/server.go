package game_gate

import (
	"encoding/binary"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/xi123/libgo/core/net/conn"
	"github.com/xi123/libgo/core/net/tcp/tcpserver"
	"github.com/xi123/libgo/core/net/transmit"
	"github.com/xi123/libgo/logs"
	"github.com/xi123/libgo/utils"
	"github.com/xi123/libgo/utils/cmd"
	"github.com/xi123/libgo/utils/codec"
	"github.com/xi123/libgo/utils/conv"
	"github.com/xi123/libgo/utils/packet"
	"github.com/xi123/libgo/utils/random"
	"github.com/xi123/libgo/utils/user_context"
	"github.com/xi123/libgo/utils/user_session"
	"github.com/cwloo/grpc-etcdv3/getcdv3"
	pb_Gamecomm "github.com/cwloo/server/proto/game.comm"
	pb_Gamehall "github.com/cwloo/server/proto/game.hall"
	"github.com/cwloo/server/src/config"
)

// <summary>
// Server
// <summary>
type Server struct {
	proto   string
	addr    string
	port    int
	path    string
	n       int32
	maxConn int
	server  tcpserver.TCPServer
	users   *user_session.UserToPlatforms
}

func (s *Server) Server() tcpserver.TCPServer {
	return s.server
}

func (s *Server) Init(id int, name string) {
	s.maxConn = config.Config.GameGate.MaxConn
	switch cmd.PatternArg("server") {
	case "":
		if id >= len(config.Config.GameGate.Port) {
			logs.Fatalf("error id=%v GameGate.Port.size=%v", id, len(config.Config.GameGate.Port))
		}
		s.proto = config.Config.GameGate.Proto
		s.addr = config.Config.GameGate.Ip
		s.port = config.Config.GameGate.Port[id]
		s.path = config.Config.GameGate.Path.Handshake
	default:
		addr := conn.ParseAddress(cmd.PatternArg("server"))
		switch addr {
		case nil:
			logs.Fatalf("error")
		default:
			s.proto = addr.Proto
			s.addr = addr.Ip
			s.port = conv.StrToInt(addr.Port)
			s.path = addr.Path

			config.Config.GameGate.Proto = s.proto
			config.Config.GameGate.Ip = s.addr
			config.Config.GameGate.Port[id] = s.port
			config.Config.GameGate.Path.Handshake = s.path
		}
	}
	s.users = user_session.NewUserToPlatforms()
	addr := strings.Join([]string{s.proto, "://", net.JoinHostPort(s.addr, strconv.Itoa(s.port)), s.path}, "")
	s.server = tcpserver.NewTCPServer(name, addr)
	s.server.SetProtocolCallback(s.onProtocol)
	s.server.SetHandshakeCallback(s.onHandshake)
	s.server.SetConnectedCallback(s.onConnected)
	s.server.SetClosedCallback(s.onClosed)
	s.server.SetMessageCallback(s.onMessage)
	s.server.SetHandshakeTimeout(time.Duration(config.Config.GameGate.HandshakeTimeout) * time.Second)
	s.server.SetIdleTimeout(time.Duration(config.Config.GameGate.IdleTimeout) * time.Second)
	s.server.SetReadBufferSize(config.Config.GameGate.ReadBufferSize)
	s.server.SetHoldType(conn.KHoldNone)
	logs.Tracef("%v %v", s.server.Name(), s.server.ListenAddr().Addr)
}

func (s *Server) NumOfLoads() int {
	return s.users.Len()
}

func (s *Server) Run(id int, name string) {
	// s.server.ListenTCP(strings.Join([]string{s.proto, "://", s.server.ListenAddr().Addr, s.path}, ""))
	s.server.ListenTCP()
}

func (s *Server) onProtocol(proto string) transmit.Channel {
	switch proto {
	case "tcp":
		logs.Fatalf("tcp Channel undefine")
	case "ws", "wss":
		return packet.NewWSChannel()
	}
	panic("no proto setup")
}

func (s *Server) onHandshake(w http.ResponseWriter, r *http.Request) bool {
	if atomic.LoadInt32(&s.n) >= int32(s.maxConn) {
		logs.Errorf("maxConn=%v", s.maxConn)
		return false
	}
	// query := r.URL.Query()
	// logs.Infof("%v %v %v %#v", r.Method, r.URL.String(), query, r.Header)
	return true
}

func (s *Server) onConnected(peer conn.Session, v ...any) {
	if peer.Connected() {
		num := atomic.AddInt32(&s.n, 1)
		logs.Infof("%d [%v] <- [%v]", num, peer.LocalAddr(), peer.RemoteAddr())
		r := v[1].(*http.Request)
		// query := r.URL.Query()
		logs.Infof("%v", r.URL.String())
		ctx := user_context.NewCtx()
		// ctx.SetUserId(query["sendID"][0])
		// ctx.SetPlatformId(conv.StrToInt(query["platformID"][0]))
		// ctx.SetToken(query["token"][0])
		ctx.SetSession(random.CreateGUID())
		peer.SetContext("ctx", ctx)
		// s.addUserConn(peer)
	} else {
		panic("error")
	}
}

func (s *Server) onClosed(peer conn.Session, reason conn.Reason) {
	if peer.Connected() {
		logs.Fatalf("error")
	} else {
		num := atomic.AddInt32(&s.n, -1)
		logs.Infof(" %d [%v] <- [%v] %v", num, peer.LocalAddr(), peer.RemoteAddr(), reason.Msg)
		// s.delUserConn(peer)
		peer.SetContext("ctx", nil)
	}
}

func (s *Server) onMessage(peer conn.Session, msg any, recvTime utils.Timestamp) {
	switch b := msg.(type) {
	case []byte:
		cmd, data, err := packet.Unpack(b, binary.LittleEndian)
		if err != nil {
			logs.Errorf(err.Error())
			peer.Close()
			return
		}
		s.onData(peer, cmd, data)
	}
}

func (s *Server) onData(peer conn.Session, cmd uint32, msg any) {
	mainID, subID := packet.Deword(int(cmd))
	switch mainID {
	case int(pb_Gamecomm.MAINID_MAIN_MESSAGE_CLIENT_TO_PROXY):
	case int(pb_Gamecomm.MAINID_MAIN_MESSAGE_CLIENT_TO_HALL):
		switch subID {
		case int(pb_Gamecomm.MESSAGE_CLIENT_TO_HALL_SUBID_CLIENT_TO_HALL_LOGIN_MESSAGE_REQ):
			rpcConn, err := getcdv3.GetBalanceConn(config.Config.Etcd.Schema, config.Config.Rpc.GameHall.Node)
			if err != nil {
				rpcConn.Free()
				req := pb_Gamehall.LoginMessage{}
				err := codec.Decode(msg.([]byte), &req)
				if err != nil {
					panic(err.Error())
				}
				rsp := pb_Gamehall.LoginMessageResponse{}
				rsp.Header.Sign = req.Header.Sign
				rsp.RetCode = 1
				rsp.ErrorMsg = "No available game_hall found"
				return
			}
		}
	case int(pb_Gamecomm.MAINID_MAIN_MESSAGE_CLIENT_TO_GAME_SERVER):
		fallthrough
	case int(pb_Gamecomm.MAINID_MAIN_MESSAGE_CLIENT_TO_GAME_LOGIC):
	default:
		panic("error")
	}
	// switch cmd {
	// case uint32(packet.Enword(int(pb_Gamecomm.MAINID_MAIN_MESSAGE_CLIENT_TO_HALL),
	// 	int(pb_Gamecomm.MESSAGE_CLIENT_TO_HALL_SUBID_CLIENT_TO_HALL_LOGIN_MESSAGE_REQ))):
	// 	// req := pb_Gamehall.LoginMessage{}
	// 	// err := codec.Decode(msg.([]byte), &req)
	// 	// if err != nil {
	// 	// 	panic(err.Error())
	// 	// }
	// 	// logs.Infof("---------%v", req)
	// 	// getcdv3.Get
	// default:
	// 	panic("error")
	// }
}

func (s *Server) addUserConn(peer conn.Session) {
	ctx := peer.GetContext("ctx").(user_context.Ctx)
	go s.MultiLoginCheck(peer)
	s.KickUserCheck(peer)
	s.users.AddUserConn(ctx.GetUserId(), ctx.GetPlatformId(), ctx.GetSession(), peer)

	// promePkg.PromeGaugeInc(promePkg.OnlineUserGauge)

	// s.onUserOnline(peer)
}

func (s *Server) delUserConn(peer conn.Session) {
	ctx := peer.GetContext("ctx").(user_context.Ctx)
	s.users.DelUserConn(ctx.GetUserId(), ctx.GetPlatformId(), ctx.GetSession())

	// promePkg.PromeGaugeDec(promePkg.OnlineUserGauge)

	// s.onUserOffline(peer, operationId)
}

func (s *Server) MultiLoginCheck(peer conn.Session) {
	// ctx := peer.GetContext("ctx").(user_context.Ctx)
	// rpcConns := getcdv3.GetConns(config.Config.Etcd.Schema, config.Config.Rpc.GameGate.Node)
	// for _, v := range rpcConns {
	// 	if v.Conn().Target() == global.RpcServer.Target() {
	// 		logs.Debugf("Filter self=%v out", global.RpcServer.Target())
	// 		continue
	// 	}
	// 	client := pb_gamegate.NewGameGateClient(v.Conn())
	// 	req := &pb_gamegate.MultiLoginCheckReq{
	// 		PlatformID: int32(ctx.GetPlatformId()),
	// 		UserID:     ctx.GetUserId(),
	// 		SessionID:  ctx.GetSession(),
	// 		Token:      ctx.GetToken()}
	// 	rsp, err := client.MultiLoginCheck(context.Background(), req)
	// 	if err != nil {
	// 		logs.Errorf("%v", err.Error())
	// 		continue
	// 	}
	// 	if rsp.ErrCode != 0 {
	// 		logs.Errorf("%v %v", rsp.ErrCode, rsp.ErrMsg)
	// 		continue
	// 	}
	// 	logs.Debugf("%v", rsp.String())
	// }
}

func (s *Server) KickUserCheck(self conn.Session) {
	ctx := self.GetContext("ctx").(user_context.Ctx)
	userId := ctx.GetUserId()
	// platformId := ctx.GetPlatformId()
	sessionId := ctx.GetSession()
	// token := ctx.GetToken()
	if platforms := s.users.Get(userId); platforms != nil {
		platforms.Range(func(platformId int, sessions *user_session.SessionToConn) {
			sessions.Range(func(sid string, peer conn.Session) {
				if sid != sessionId { //not self
					if self == peer {
						logs.Fatalf("peer is self!")
					}
					s.sendKickMsg(peer)
					peer.CloseAfter(time.Second)
				} else {
					logs.Warnf("can't kick self!")
				}
			})
			//key UID_PID_TOKEN_STATUS:userid:platform field:token value:constant.KickedToken
			// m, err := db.DB.GetTokenMapByUidPid(userId, constant.PlatformIDToName(platformId))
			// if err != nil && err != go_redis.Nil {
			// 	return
			// }
			// if m == nil {
			// 	return
			// }
			// for k := range m {
			// 	if k != token {
			// 		m[k] = constant.KickedToken
			// 	}
			// }
			// err = db.DB.SetTokenMapByUidPid(userId, platformId, m)
			// if err != nil {
			// 	return
			// }
		})
	}
}

func (s *Server) KickUserCheckBy(userId string, platformId int, sessionId string, token string) {
	if platforms := s.users.Get(userId); platforms != nil {
		platforms.Range(func(platformId int, sessions *user_session.SessionToConn) {
			sessions.Range(func(sid string, peer conn.Session) {
				if sid != sessionId { //not self
					s.sendKickMsg(peer)
					peer.CloseAfter(time.Second)
				} else {
					logs.Warnf("can't kick self!")
				}
			})
			//key UID_PID_TOKEN_STATUS:userid:platform field:token value:constant.KickedToken
			// m, err := db.DB.GetTokenMapByUidPid(userId, constant.PlatformIDToName(platformId))
			// if err != nil && err != go_redis.Nil {
			// 	return
			// }
			// if m == nil {
			// 	return
			// }
			// for k := range m {
			// 	if k != token {
			// 		m[k] = constant.KickedToken
			// 	}
			// }
			// err = db.DB.SetTokenMapByUidPid(userId, platformId, m)
			// if err != nil {
			// 	return
			// }
		})
	}
}

func (s *Server) sendKickMsg(peer conn.Session) {
	// mReply := Resp{
	// 	ReqIdentifier: constant.WSKickOnlineMsg,
	// 	ErrCode:       constant.ErrTokenInvalid.ErrCode,
	// 	ErrMsg:        constant.ErrTokenInvalid.ErrMsg,
	// }
	// var b bytes.Buffer
	// enc := gob.NewEncoder(&b)
	// err := enc.Encode(mReply)
	// if err != nil {
	// 	logs.LogError("%v", err)
	// 	return
	// }
	// peer.Write(b.Bytes())
}

func (s *Server) GetUserConn(userId string, platformId int, sessionId string) conn.Session {
	return s.users.GetUserConn(userId, platformId, sessionId)
}

func (s *Server) GetUserPlatformConns(userId string, platformId int) *user_session.SessionToConn {
	return s.users.GetUserPlatformConns(userId, platformId)
}

func (s *Server) GetUserConns(userId string) *user_session.PlatformToSessions {
	return s.users.GetUserConns(userId)
}
