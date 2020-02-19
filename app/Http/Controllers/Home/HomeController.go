package Home

import (
	"github.com/gogf/gf-demos/library/response"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gtime"
)

// 用户API管理对象
type Controller struct{}

// @summary 用户注册接口
// @tags    用户服务
// @produce json
// @param   passport  formData string  true "用户账号名称"
// @param   password  formData string  true "用户密码"
// @param   password2 formData string  true "确认密码"
// @param   nickname  formData string false "用户昵称"
// @router  /user/signup [POST]
// @success 200 "执行结果"
func (c *Controller) Home(r *ghttp.Request) {
	datetime := r.Cookie.Get("datetime")
	r.Cookie.Set("datetime", gtime.Datetime())
	//r.Response.Write("datetime:", datetime)
	response.JsonExit(r, 0, "ok", g.Map{
		"datetime": datetime,
	})

	//_ = r.Response.WriteJson(g.Map{ "datetime": datetime, })

}
func (c *Controller) List(r *ghttp.Request) {
	response.JsonExit(r, 0, "List")
}
func (c *Controller) Artice(r *ghttp.Request) {
	response.JsonExit(r, 0, "Artice")
}
func (c *Controller) List2(r *ghttp.Request) {
	response.JsonExit(r, 0, "List")
}

func (c *Controller) Auth(r *ghttp.Request) {
	response.JsonExit(r, 0, "Auth")
}
