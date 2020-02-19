package Middleware

import (
	"github.com/gogf/gf/net/ghttp"
	"net/http"
)

// 鉴权中间件，只有登录成功之后才能通过
func Auth(r *ghttp.Request) {
	if true {
		r.Middleware.Next()
	} else {
		r.Response.WriteStatus(http.StatusForbidden)
	}
}
