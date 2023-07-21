package register

import (
	"github.com/cwloo/gonet/logs"
	"github.com/cwloo/gonet/utils/response"
	"github.com/cwloo/server/src/game_login/reqstruct"
	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	r := c.Request
	logs.Infof("%v %v %#v", r.Method, r.URL.String(), r.Header)
	req := reqstruct.RegisterReq{}
	if err := c.BindJSON(&req); err != nil {
		response.BadRequest(c)
		return
	}
	// token, domain, err := gamedb.GetToken(req.Account)
	// if err != nil {
	// 	response.Result(-2, err.Error(), req, gin.H{}, c)
	// 	return
	// }
	response.Ok(req, reqstruct.RegisterRsp{}, c)
}
