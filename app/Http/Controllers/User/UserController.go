package User

import (
	"database/sql"
	"github.com/gogf/gf-demos/library/response"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/util/gconv"
	"laravel-go/app/Http/Models/post"
	"laravel-go/app/Http/Service/user"
)

// 用户API管理对象
type Controller struct{}

// @Summary hello接口
// @Description hello接口
// @Tags hello
// @Success 200 {string} string	"ok"
// @router / [get]
func (c *Controller) Home(r *ghttp.Request) {
	response.JsonExit(r, 0, "hello")
}
func (c *Controller) List(r *ghttp.Request) {
	response.JsonExit(r, 0, "List")
}
func (c *Controller) Info(r *ghttp.Request) {
	response.JsonExit(r, 0, "Info")
}

// 注册请求参数，用于前后端交互参数格式约定
type SignUpRequest struct {
	user.SignUpInput
}

// @summary 用户注册接口
// @tags    用户服务
// @produce json
// @param   passport  formData string  true "用户账号名称"
// @param   password  formData string  true "用户密码"
// @param   password2 formData string  true "确认密码"
// @param   nickname  formData string false "用户昵称"
// @router  /user/signup [POST]
// @success 200 {object} response.JsonResponse "执行结果"
func (c *Controller) SignUp(r *ghttp.Request) {
	var data *SignUpRequest
	// 这里没有使用Parse而是仅用GetStruct获取对象，
	// 数据校验交给后续的service层统一处理
	if err := r.GetStruct(&data); err != nil {
		response.JsonExit(r, 1, err.Error())
	}

	if err := user.SignUp(&data.SignUpInput); err != nil {
		response.JsonExit(r, 1, err.Error())
	}

	response.JsonExit(r, 0, "ok")

}

func (c *Controller) Post(r *ghttp.Request) {
	var entity post.Entity
	entity.Title = "aaaa"
	entity.Content = "bbbb"
	entity.Time = gconv.Int(gtime.Now().Timestamp())
	var rr sql.Result
	var err error
	if rr, err = post.Save(entity); err != nil {
		response.JsonExit(r, 0, "no")
	}
	id, _ := rr.LastInsertId()

	data := gconv.Map(entity)
	data["bbb"] = "bbb"
	data["id"] = gconv.Int(id)

	response.JsonExit(r, 0, "ok", data)
}

func (c *Controller) Posts(r *ghttp.Request) {
	count, _ := post.Model.Count()
	PageBarNum := 10
	CurrentPage := gconv.Int(r.Get("page"))

	all, _ := post.Model.Limit((CurrentPage-1)*PageBarNum, PageBarNum).FindAll()

	data := g.Map{}
	data["bbb"] = "bbb"
	data["总页数"] = gconv.Int(count / PageBarNum)
	data["当前页"] = gconv.Int(r.Get("page"))
	data["总数"] = gconv.Int(count)
	data["当前页条数"] = gconv.Int(len(all))

	data["data"] = all

	response.JsonExit(r, 0, "ok", data)
}
