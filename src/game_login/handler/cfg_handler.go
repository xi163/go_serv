package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/cwloo/gonet/logs"
	"github.com/cwloo/server/src/global"
	"github.com/cwloo/server/src/global/httpsrv"
)

func handlerUpdateCfgJsonReq(body []byte) (*global.UpdateCfgResp, bool) {
	if len(body) == 0 {
		return &global.UpdateCfgResp{ErrCode: 3, ErrMsg: "no body"}, false
	}
	logs.Warnf("%v", string(body))
	req := global.UpdateCfgReq{}
	err := json.Unmarshal(body, &req)
	if err != nil {
		logs.Errorf(err.Error())
		return &global.UpdateCfgResp{ErrCode: 4, ErrMsg: "parse body error"}, false
	}
	logs.Debugf("%#v", req)
	return UpdateCfg(&req)
}

func handlerUpdateCfgQuery(query url.Values) (*global.UpdateCfgResp, bool) {
	req := &global.UpdateCfgReq{}
	if query.Has("interval") && len(query["interval"]) > 0 {
		req.Interval = query["interval"][0]
	}
	if query.Has("log_timezone") && len(query["log_timezone"]) > 0 {
		req.LogTimezone = query["log_timezone"][0]
	}
	if query.Has("log_mode") && len(query["log_mode"]) > 0 {
		req.LogMode = query["log_mode"][0]
	}
	if query.Has("log_style") && len(query["log_style"]) > 0 {
		req.LogStyle = query["log_style"][0]
	}
	if query.Has("log_level") && len(query["log_level"]) > 0 {
		req.LogLevel = query["log_level"][0]
	}
	if query.Has("maxMemory") && len(query["maxMemory"]) > 0 {
		req.MaxMemory = query["maxMemory"][0]
	}
	if query.Has("maxSegmentSize") && len(query["maxSegmentSize"]) > 0 {
		req.MaxSegmentSize = query["maxSegmentSize"][0]
	}
	if query.Has("maxSingleSize") && len(query["maxSingleSize"]) > 0 {
		req.MaxSingleSize = query["maxSingleSize"][0]
	}
	if query.Has("maxTotalSize") && len(query["maxTotalSize"]) > 0 {
		req.MaxTotalSize = query["maxTotalSize"][0]
	}
	if query.Has("pendingTimeout") && len(query["pendingTimeout"]) > 0 {
		req.PendingTimeout = query["pendingTimeout"][0]
	}
	if query.Has("fileExpiredTimeout") && len(query["fileExpiredTimeout"]) > 0 {
		req.FileExpiredTimeout = query["fileExpiredTimeout"][0]
	}
	if query.Has("checkMd5") && len(query["checkMd5"]) > 0 {
		req.CheckMd5 = query["checkMd5"][0]
	}
	if query.Has("writeFile") && len(query["writeFile"]) > 0 {
		req.WriteFile = query["writeFile"][0]
	}
	if query.Has("useTgBot") && len(query["useTgBot"]) > 0 {
		req.UseTgBot = query["useTgBot"][0]
	}
	if query.Has("tg_chatId") && len(query["tg_chatId"]) > 0 {
		req.TgBotChatId = query["tg_chatId"][0]
	}
	if query.Has("tg_token") && len(query["tg_token"]) > 0 {
		req.TgBotToken = query["tg_token"][0]
	}
	logs.Debugf("%#v", req)
	return UpdateCfg(req)
}

func UpdateCfgReq(w http.ResponseWriter, r *http.Request) {
	logs.Infof("%v %v %#v", r.Method, r.URL.String(), r.Header)
	switch strings.ToUpper(r.Method) {
	case http.MethodPost:
		switch r.Header.Get("Content-Type") {
		case "application/json":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				logs.Errorf(err.Error())
				resp := &global.UpdateCfgResp{ErrCode: 2, ErrMsg: "read body error"}
				httpsrv.WriteResponse(w, r, resp)
				return
			}
			resp, _ := handlerUpdateCfgJsonReq(body)
			httpsrv.WriteResponse(w, r, resp)
		default:
			resp, _ := handlerUpdateCfgQuery(r.URL.Query())
			httpsrv.WriteResponse(w, r, resp)
		}
	case http.MethodGet:
		switch r.Header.Get("Content-Type") {
		case "application/json":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				logs.Errorf(err.Error())
				resp := &global.UpdateCfgResp{ErrCode: 2, ErrMsg: "read body error"}
				httpsrv.WriteResponse(w, r, resp)
				return
			}
			resp, _ := handlerUpdateCfgJsonReq(body)
			httpsrv.WriteResponse(w, r, resp)
		default:
			resp, _ := handlerUpdateCfgQuery(r.URL.Query())
			httpsrv.WriteResponse(w, r, resp)
		}
	case http.MethodOptions:
		switch r.Header.Get("Content-Type") {
		case "application/json":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				logs.Errorf(err.Error())
				resp := &global.UpdateCfgResp{ErrCode: 2, ErrMsg: "read body error"}
				httpsrv.WriteResponse(w, r, resp)
				return
			}
			resp, _ := handlerUpdateCfgJsonReq(body)
			httpsrv.WriteResponse(w, r, resp)
		default:
			resp, _ := handlerUpdateCfgQuery(r.URL.Query())
			httpsrv.WriteResponse(w, r, resp)
		}
	}
}

func handlerGetCfgJsonReq(body []byte) (*global.GetCfgResp, bool) {
	if len(body) == 0 {
		return &global.GetCfgResp{ErrCode: 3, ErrMsg: "no body"}, false
	}
	logs.Warnf("%v", string(body))
	req := global.GetCfgReq{}
	err := json.Unmarshal(body, &req)
	if err != nil {
		logs.Errorf(err.Error())
		return &global.GetCfgResp{ErrCode: 4, ErrMsg: "parse body error"}, false
	}
	logs.Debugf("%#v", req)
	return GetCfg(&req)
}

func handlerGetCfgQuery(query url.Values) (*global.GetCfgResp, bool) {
	req := &global.GetCfgReq{}
	// if query.Has("interval") && len(query["interval"]) > 0 {
	// 	req.Interval = query["interval"][0]
	// }
	// if query.Has("maxMemory") && len(query["maxMemory"]) > 0 {
	// 	req.MaxMemory = query["maxMemory"][0]
	// }
	// if query.Has("maxSegmentSize") && len(query["maxSegmentSize"]) > 0 {
	// 	req.MaxSegmentSize = query["maxSegmentSize"][0]
	// }
	// if query.Has("maxSingleSize") && len(query["maxSingleSize"]) > 0 {
	// 	req.MaxSingleSize = query["maxSingleSize"][0]
	// }
	// if query.Has("maxTotalSize") && len(query["maxTotalSize"]) > 0 {
	// 	req.MaxTotalSize = query["maxTotalSize"][0]
	// }
	// if query.Has("pendingTimeout") && len(query["pendingTimeout"]) > 0 {
	// 	req.PendingTimeout = query["pendingTimeout"][0]
	// }
	// if query.Has("fileExpiredTimeout") && len(query["fileExpiredTimeout"]) > 0 {
	// 	req.FileExpiredTimeout = query["fileExpiredTimeout"][0]
	// }
	// if query.Has("checkMd5") && len(query["checkMd5"]) > 0 {
	// 	req.CheckMd5 = query["checkMd5"][0]
	// }
	// if query.Has("writeFile") && len(query["writeFile"]) > 0 {
	// 	req.WriteFile = query["writeFile"][0]
	// }
	logs.Debugf("%#v", req)
	return GetCfg(req)
}

func GetCfgReq(w http.ResponseWriter, r *http.Request) {
	logs.Infof("%v %v %#v", r.Method, r.URL.String(), r.Header)
	switch strings.ToUpper(r.Method) {
	case http.MethodPost:
		switch r.Header.Get("Content-Type") {
		case "application/json":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				logs.Errorf(err.Error())
				resp := &global.GetCfgResp{ErrCode: 2, ErrMsg: "read body error"}
				httpsrv.WriteResponse(w, r, resp)
				return
			}
			resp, _ := handlerGetCfgJsonReq(body)
			httpsrv.WriteResponse(w, r, resp)
		default:
			resp, _ := handlerGetCfgQuery(r.URL.Query())
			httpsrv.WriteResponse(w, r, resp)
		}
	case http.MethodGet:
		switch r.Header.Get("Content-Type") {
		case "application/json":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				logs.Errorf(err.Error())
				resp := &global.GetCfgResp{ErrCode: 2, ErrMsg: "read body error"}
				httpsrv.WriteResponse(w, r, resp)
				return
			}
			resp, _ := handlerGetCfgJsonReq(body)
			httpsrv.WriteResponse(w, r, resp)
		default:
			resp, _ := handlerGetCfgQuery(r.URL.Query())
			httpsrv.WriteResponse(w, r, resp)
		}
	case http.MethodOptions:
		switch r.Header.Get("Content-Type") {
		case "application/json":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				logs.Errorf(err.Error())
				resp := &global.GetCfgResp{ErrCode: 2, ErrMsg: "read body error"}
				httpsrv.WriteResponse(w, r, resp)
				return
			}
			resp, _ := handlerGetCfgJsonReq(body)
			httpsrv.WriteResponse(w, r, resp)
		default:
			resp, _ := handlerGetCfgQuery(r.URL.Query())
			httpsrv.WriteResponse(w, r, resp)
		}
	}
}
