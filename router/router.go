package router

import (
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"laravel-go/app/Http/Controllers/Api"
)

func init() {
	s := g.Server()

	// 某些浏览器直接请求favicon.ico文件，特别是产生404时
	//s.SetRewrite("/favicon.ico", "/resource/image/favicon.ico")
	//
	//s.BindHandler("/status/:status", func(r *ghttp.Request) {
	//	r.Response.Write("woops, status ", r.Get("status"), " found")
	//})
	//s.BindStatusHandler(404, func(r *ghttp.Request) {
	//	r.Response.RedirectTo("/status/404")
	//})
	//s.BindHookHandler("/*", ghttp.HOOK_BEFORE_SERVE, beforeServeHook1)

	// 分组路由注册方式
	//s.Group("/", func(group *ghttp.RouterGroup) {
	//	ctlChat := new(chat.Controller)
	//	ctlUser := new(user.Controller)
	//	group.Middleware(middleware.CORS)
	//	group.ALL("/chat", ctlChat)
	//	group.ALL("/user", ctlUser)
	//	group.ALL("/curd/:table", new(curd.Controller))
	//	group.Group("/", func(group *ghttp.RouterGroup) {
	//		group.Middleware(middleware.Auth)
	//		group.ALL("/user/profile", ctlUser, "Profile")
	//	})
	//})
	//s.Group("/", func(group *ghttp.RouterGroup) {
	//	HomeController := new(Home.Controller)
	//	UserController := new(User.Controller)
	//	group.ALL("/", HomeController)
	//	group.ALL("/aaa", HomeController.Home)
	//
	//	group.ALL("/users", UserController)
	//
	//	group.Group("/users", func(group *ghttp.RouterGroup) {
	//		group.Middleware(Middleware.Auth)
	//		group.ALL("/auth", HomeController.Auth)
	//	})
	//})

	s.Group("/", func(group *ghttp.RouterGroup) {
		ApiController := new(Api.Controller)
		group.ALL("/api", ApiController)

	})
}
