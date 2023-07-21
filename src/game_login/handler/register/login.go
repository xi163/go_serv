package register

import (
	"context"
	"fmt"
	"time"

	"github.com/xi123/libgo/logs"
	"github.com/xi123/libgo/utils/codec/base64"
	URI "github.com/xi123/libgo/utils/codec/uri"
	"github.com/xi123/libgo/utils/conv"
	"github.com/xi123/libgo/utils/crypto/aes"
	"github.com/xi123/libgo/utils/crypto/md5"
	"github.com/xi123/libgo/utils/json"
	"github.com/xi123/libgo/utils/response"
	"github.com/xi123/libgo/utils/sign"
	"github.com/cwloo/grpc-etcdv3/getcdv3"
	pb_gamegate "github.com/cwloo/server/proto/game.gate"
	"github.com/cwloo/server/src/common/mongoop"
	"github.com/cwloo/server/src/common/redisop"
	"github.com/cwloo/server/src/common/redisop/redisKeys"
	"github.com/cwloo/server/src/common/utils"
	"github.com/cwloo/server/src/config"
	"github.com/cwloo/server/src/game_login/reqstruct"
	"github.com/cwloo/server/src/global/Err"
	"github.com/gin-gonic/gin"
	redis "github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
)

var Md5Key = "334270F58E3E9DEC"
var AesKey = "111362EE140F157D"

func CreateGuestUser(seq int64, account string) (model mongoop.GameUser) {
	model.UserId = seq
	if account == "" {
		model.Account = "guest_" + utils.IntToStr(model.UserId)
	} else {
		model.Account = account
	}
	model.Agentid = 10001
	model.Linecode = "10001_1"
	model.Nickname = model.Account
	model.Headindex = 0
	model.Registertime = time.Now()
	// model.Regip = c.RemoteIP()
	model.Lastlogintime = time.Now()
	// model.Lastloginip = c.RemoteIP()
	model.Activedays = 0
	model.Keeplogindays = 0
	model.Alladdscore = 0
	model.Allsubscore = 0
	model.Addscoretimes = 0
	model.Subscoretimes = 0
	model.Gamerevenue = 0
	model.WinLosescore = 0
	model.Score = 0
	model.Status = 0
	model.Onlinestatus = 0
	model.Gender = 0
	model.Integralvalue = 0
	return
}

func GetGateServerList() (servList []reqstruct.ServerLoad) {
	rpcConns := getcdv3.GetConns(config.Config.Etcd.Schema, config.Config.Rpc.GameGate.Node)
	for _, v := range rpcConns {
		client := pb_gamegate.NewGameGateClient(v.Conn())
		switch client {
		case nil:
			continue
		}
		req := &pb_gamegate.GameGateReq{}
		rsp, err := client.GetGameGate(context.Background(), req)
		if err != nil {
			logs.Errorf(err.Error())
			// gRPCs.Conns().RemoveBy(err)
			// v.Close()
			v.Free()
			continue
		}
		servList = append(servList, reqstruct.ServerLoad{Host: rsp.Host, Domain: rsp.Domain, NumOfLoads: int(rsp.NumOfLoads)})
		v.Free()
	}
	return
}

func Login(c *gin.Context) {
	r := c.Request
	req := reqstruct.Request{}
	if err := c.BindJSON(&req); err != nil {
		logs.Infof("%v %v %#v", r.Method, r.URL.String(), r.Header)
		response.BadRequest(c)
		return
	}
	logs.Infof("%v %v %v\n%#v", r.Method, r.URL.String(), json.String(&req), r.Header)
	strBase64 := URI.URLDecode(req.Param)
	encrypt := base64.URLDecode(strBase64)
	param := aes.ECBDecryptPKCS7(encrypt, []byte(AesKey), []byte(AesKey))
	// model := map[string]any{}
	// URI.ParseQuery(string(param), &model)
	vParam := &reqstruct.LoginReq{}
	// json.MapToStruct(&model, &vParam)
	json.Parse(param, &vParam)
	logs.Debugf(string(param))
	key := md5.Md5(fmt.Sprintf("%v%v%v%v", vParam.Account, vParam.Type, vParam.Timestamp, Md5Key), true)
	if key != req.Key {
		response.BadRequest(c)
		return
	}
	DoLogin(c, vParam)
}

func DoLogin(c *gin.Context, req *reqstruct.LoginReq) {
	//0-游客 1-账号密码 2-手机号 3-第三方(微信/支付宝等) 4-邮箱
	switch req.Type {
	case 0:
		if req.Account == "" {
			//查询网关节点
			servList := GetGateServerList()
			if len(servList) == 0 {
				Err.Result(Err.ErrGameGateNotExist, nil, c)
				return
			}
			//生成userid
			m := mongoop.FindOneAndUpdate(bson.M{"seq": 1}, bson.M{"$inc": bson.M{"seq": 1}}, bson.M{"_id": "userid"})
			if m == nil {
				Err.Result(Err.ErrCreateGameUser, nil, c)
				return
			}
			//创建并插入user表
			model := CreateGuestUser(m["seq"].(int64), req.Account)
			model.Regip = c.RemoteIP()
			model.Lastloginip = c.RemoteIP()
			_, err := mongoop.InsertOneGameUser(&model)
			if err != nil {
				Err.Result(Err.ErrCreateGameUser, nil, c)
				return
			}
			//token签名加密
			token := sign.Encode(reqstruct.LoginRsp{
				Account: model.Account,
				Userid:  model.UserId,
				Data:    servList,
			}, time.Now().Add(redisKeys.Expire_Token*time.Second), []byte(AesKey))
			//更新redis account->uid
			redisop.SetAccountUid(model.Account, conv.Int64ToStr(m["seq"].(int64)))
			//缓存token
			redisop.SetToken(token, m["seq"].(int64), model.Account)
			response.OkMsg("登陆成功", nil, gin.H{"token": token}, c)
			return
		}
		//先查redis
		v, err := redisop.GetAccountUid(req.Account)
		if err != nil && err != redis.Nil {
			Err.Result(Err.ErrCreateGameUser, nil, c)
			return
		}
		switch err {
		case redis.Nil:
			//再查mongo
			cols := mongoop.FindOneFromGameUSer(bson.M{"userid": 1}, bson.M{"account": req.Account})
			if cols == nil {
				//查询网关节点
				servList := GetGateServerList()
				if len(servList) == 0 {
					Err.Result(Err.ErrGameGateNotExist, nil, c)
					return
				}
				//生成userid
				m := mongoop.FindOneAndUpdate(bson.M{"seq": 1}, bson.M{"$inc": bson.M{"seq": 1}}, bson.M{"_id": "userid"})
				if m == nil {
					Err.Result(Err.ErrCreateGameUser, nil, c)
					return
				}
				//创建并插入user表
				model := CreateGuestUser(m["seq"].(int64), req.Account)
				model.Regip = c.RemoteIP()
				model.Lastloginip = c.RemoteIP()
				_, err := mongoop.InsertOneGameUser(&model)
				if err != nil {
					Err.Result(Err.ErrCreateGameUser, nil, c)
					return
				}
				//token签名加密
				token := sign.Encode(reqstruct.LoginRsp{
					Account: model.Account,
					Userid:  model.UserId,
					Data:    servList,
				}, time.Now().Add(redisKeys.Expire_Token*time.Second), []byte(AesKey))
				//更新redis account->uid
				redisop.SetAccountUid(model.Account, conv.Int64ToStr(m["seq"].(int64)))
				//缓存token
				redisop.SetToken(token, m["seq"].(int64), model.Account)
				response.OkMsg("登陆成功", nil, gin.H{"token": token}, c)
			} else {
				//查询mongo命中
				userid := cols["userid"].(int64)
				//查询网关节点
				servList := GetGateServerList()
				if len(servList) == 0 {
					Err.Result(Err.ErrGameGateNotExist, nil, c)
					return
				}
				//token签名加密
				token := sign.Encode(reqstruct.LoginRsp{
					Account: req.Account,
					Userid:  userid,
					Data:    servList,
				}, time.Now().Add(redisKeys.Expire_Token*time.Second), []byte(AesKey))
				//更新redis account->uid
				redisop.SetAccountUid(req.Account, conv.Int64ToStr(userid))
				//缓存token
				redisop.SetToken(token, userid, req.Account)
				response.OkMsg("登陆成功", nil, gin.H{"token": token}, c)
			}
		case nil:
			//查询redis命中
			userid := utils.StrToInt64(v)
			//查询网关节点
			servList := GetGateServerList()
			if len(servList) == 0 {
				Err.Result(Err.ErrGameGateNotExist, nil, c)
				return
			}
			//token签名加密
			token := sign.Encode(reqstruct.LoginRsp{
				Account: req.Account,
				Userid:  userid,
				Data:    servList,
			}, time.Now().Add(redisKeys.Expire_Token*time.Second), []byte(AesKey))
			//更新redis account->uid
			redisop.ExpireAccountUid(req.Account)
			//缓存token
			redisop.SetToken(token, userid, req.Account)
			response.OkMsg("登陆成功", nil, gin.H{"token": token}, c)
		default:
			logs.Fatalf("error")
		}
	case 1:
	case 2:
	case 3:
	}
}
