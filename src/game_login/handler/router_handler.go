package handler

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/xi123/libgo/logs"
	"github.com/xi123/libgo/utils/cmd"
	"github.com/xi123/libgo/utils/safe"
	"github.com/xi123/grpc-etcdv3/getcdv3"
	"github.com/xi123/grpc-etcdv3/getcdv3/gRPCs"
	pb_file "github.com/xi123/server/proto/file"
	pb_public "github.com/xi123/server/proto/public"
	"github.com/xi123/server/src/config"
	"github.com/xi123/server/src/global"
	"github.com/xi123/server/src/global/httpsrv"
)

func GetNodeInfo() (*pb_public.NodeInfoResp, error) {
	return &pb_public.NodeInfoResp{
		Node: &pb_public.NodeInfo{
			Pid:        int32(os.Getpid()),
			Name:       global.Name,
			Id:         int32(cmd.Id()) + 1,
			NumOfPends: int32(PendingNum()),
			NumOfFiles: int32(FinishedNum()),
			NumOfLoads: int32(global.Uploaders.Len()),
			Ip:         config.Config.HttpGate.Ip,
			Port:       int32(config.Config.HttpGate.Port[cmd.Id()]),
			Rpc: &pb_public.NodeInfo_Rpc{
				Ip:   config.Config.Rpc.Ip,
				Port: int32(config.Config.Rpc.HttpGate.Port[cmd.Id()]),
			},
		},
		ErrCode: 0,
		ErrMsg:  "ok"}, nil
}

func QueryRouter(md5 string) (*global.RouterResp, bool) {
	defer safe.Catch()
	rpcConns := getcdv3.GetConns(config.Config.Etcd.Schema, config.Config.Rpc.File.Node)
	// logs.Infof("%v rpcConns.size=%v", md5, len(rpcConns))
	NumOfLoads := map[string]*pb_public.RouterResp{}
	for _, v := range rpcConns {
		client := pb_file.NewFileClient(v.Conn())
		switch client {
		case nil:
			continue
		}
		req := &pb_public.RouterReq{
			Md5: md5,
		}
		resp, err := client.GetRouter(context.Background(), req)
		if err != nil {
			logs.Errorf(err.Error())
			gRPCs.Conns().RemoveBy(err)
			v.Close()
			continue
		}
		switch resp.ErrCode {
		default:
			logs.Errorf("%v %v [%v:%v %v:%v rpc:%v:%v NumOfLoads:%v]", v.Conn().Target(),
				resp.Node.Pid,
				resp.Node.Name,
				resp.Node.Id,
				resp.Node.Ip, resp.Node.Port,
				resp.Node.Rpc.Ip, resp.Node.Rpc.Port,
				resp.Node.NumOfLoads)
			NumOfLoads[resp.Node.Domain] = resp
			v.Free()
			continue
		case 0:
			logs.Infof("%v %v [%v:%v %v:%v rpc:%v:%v NumOfLoads:%v]", v.Conn().Target(),
				resp.Node.Pid,
				resp.Node.Name,
				resp.Node.Id,
				resp.Node.Ip, resp.Node.Port,
				resp.Node.Rpc.Ip, resp.Node.Rpc.Port,
				resp.Node.NumOfLoads)
			v.Free()
			return &global.RouterResp{
				Node: &pb_public.NodeInfo{
					Pid:        int32(resp.Node.Pid),
					Name:       resp.Node.Name,
					Id:         int32(resp.Node.Id),
					NumOfPends: int32(PendingNum()),
					NumOfFiles: int32(FinishedNum()),
					NumOfLoads: int32(resp.Node.NumOfLoads),
					Ip:         resp.Node.Ip,
					Port:       int32(resp.Node.Port),
					Domain:     resp.Node.Domain,
					Rpc: &pb_public.NodeInfo_Rpc{
						Ip:   resp.Node.Rpc.Ip,
						Port: int32(resp.Node.Rpc.Port),
					},
				},
				Md5:     md5,
				ErrCode: 0,
				ErrMsg:  "ok"}, true
		}
	}
	var minRouter *pb_public.RouterResp
	minLoads := -1
	for _, v := range NumOfLoads {
		switch minLoads {
		case -1:
			minRouter = v
			minLoads = int(v.Node.GetNumOfLoads())
		default:
			switch int(v.Node.GetNumOfLoads()) < minLoads {
			case true:
				minRouter = v
				minLoads = int(v.Node.GetNumOfLoads())
			}
		}
	}
	switch minRouter {
	case nil:
		return &global.RouterResp{
			Md5:     md5,
			ErrCode: 6,
			ErrMsg:  "no file_server"}, false
	default:
		return &global.RouterResp{
			Node: &pb_public.NodeInfo{
				Pid:        int32(minRouter.Node.Pid),
				Name:       minRouter.Node.Name,
				Id:         int32(minRouter.Node.Id),
				NumOfPends: int32(PendingNum()),
				NumOfFiles: int32(FinishedNum()),
				NumOfLoads: int32(minRouter.Node.NumOfLoads),
				Ip:         minRouter.Node.Ip,
				Port:       int32(minRouter.Node.Port),
				Domain:     minRouter.Node.Domain,
				Rpc: &pb_public.NodeInfo_Rpc{
					Ip:   minRouter.Node.Rpc.Ip,
					Port: int32(minRouter.Node.Rpc.Port),
				},
			},
			Md5:     md5,
			ErrCode: 0,
			ErrMsg:  "ok"}, true
	}
}

func handlerRouterJsonReq(body []byte) (*global.RouterResp, bool) {
	if len(body) == 0 {
		return &global.RouterResp{ErrCode: 3, ErrMsg: "no body"}, false
	}
	logs.Warnf("%v", string(body))
	req := global.RouterReq{}
	err := json.Unmarshal(body, &req)
	if err != nil {
		logs.Errorf(err.Error())
		return &global.RouterResp{ErrCode: 4, ErrMsg: "parse body error"}, false
	}
	if req.Md5 == "" && len(req.Md5) != 32 {
		return &global.RouterResp{Md5: req.Md5, ErrCode: 1, ErrMsg: "parse param error"}, false
	}
	return QueryRouter(req.Md5)
}

func handlerRouterQuery(query url.Values) (*global.RouterResp, bool) {
	var md5 string
	if query.Has("md5") && len(query["md5"]) > 0 {
		md5 = query["md5"][0]
	}
	if md5 == "" && len(md5) != 32 {
		return &global.RouterResp{Md5: md5, ErrCode: 1, ErrMsg: "parse param error"}, false
	}
	return QueryRouter(md5)
}

func RouterReq(w http.ResponseWriter, r *http.Request) {
	logs.Infof("%v %v %#v", r.Method, r.URL.String(), r.Header)
	switch strings.ToUpper(r.Method) {
	case http.MethodPost:
		switch r.Header.Get("Content-Type") {
		case "application/json":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				logs.Errorf(err.Error())
				resp := &global.RouterResp{ErrCode: 2, ErrMsg: "read body error"}
				httpsrv.WriteResponse(w, r, resp)
				return
			}
			resp, _ := handlerRouterJsonReq(body)
			httpsrv.WriteResponse(w, r, resp)
		default:
			resp, _ := handlerRouterQuery(r.URL.Query())
			httpsrv.WriteResponse(w, r, resp)
		}
	case http.MethodGet:
		switch r.Header.Get("Content-Type") {
		case "application/json":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				logs.Errorf(err.Error())
				resp := &global.RouterResp{ErrCode: 2, ErrMsg: "read body error"}
				httpsrv.WriteResponse(w, r, resp)
				return
			}
			resp, _ := handlerRouterJsonReq(body)
			httpsrv.WriteResponse(w, r, resp)
		default:
			resp, _ := handlerRouterQuery(r.URL.Query())
			httpsrv.WriteResponse(w, r, resp)
		}
	case http.MethodOptions:
		switch r.Header.Get("Content-Type") {
		case "application/json":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				logs.Errorf(err.Error())
				resp := &global.RouterResp{ErrCode: 2, ErrMsg: "read body error"}
				httpsrv.WriteResponse(w, r, resp)
				return
			}
			resp, _ := handlerRouterJsonReq(body)
			httpsrv.WriteResponse(w, r, resp)
		default:
			resp, _ := handlerRouterQuery(r.URL.Query())
			httpsrv.WriteResponse(w, r, resp)
		}
	}
}
