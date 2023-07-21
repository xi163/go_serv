package reqstruct

type Request struct {
	Key   string `json:"key" form:"key"`
	Param string `json:"param" form:"param"`
}

// 注册
type RegisterReq struct {
	Account string `json:"account" form:"account"`
	Phone   string `json:"phone" form:"phone"`
	Email   string `json:"email" form:"email"`
	Passwd  string `json:"passwd" form:"passwd"`
}

type RegisterRsp struct {
}

// 登陆类型：0-游客 1-账号密码 2-手机号 3-第三方(微信/支付宝等) 4-邮箱
type LoginReq struct {
	Account   string `json:"account,omitempty"`
	Passwd    string `json:"passwd,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Email     string `json:"email,omitempty"`
	Type      int    `json:"type" form:"type"`
	Timestamp int64  `json:"timestamp" form:"timestamp"`
}

type LoginRsp struct {
	Account string       `json:"account" form:"account"`
	Userid  int64        `json:"userid" form:"userid"`
	Data    []ServerLoad `json:"data" form:"data"`
}

type ServerLoad struct {
	Host       string `json:"host" form:"host"`
	Domain     string `json:"domain" form:"domain"`
	NumOfLoads int    `json:"numOfLoads" form:"numOfLoads"`
}
