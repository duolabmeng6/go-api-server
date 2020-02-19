package Api

import (
	"github.com/duolabmeng6/efun/efun"
	"github.com/gogf/gf-demos/library/response"
	"github.com/gogf/gf/net/ghttp"
	"laravel-go/app/Http/Service"
)

// 用户API管理对象
type Controller struct{}

var Queue = Service.NewApiQuue()

//获取队列中的任务
func (c *Controller) GetQueue(r *ghttp.Request) {
	response.JsonExit(r, 0, "GetQueue")
}

//提交处理后的任务数据
func (c *Controller) PutQueue(r *ghttp.Request) {
	response.JsonExit(r, 0, "LiPutQueuest")
}

//获取队列中的状态信息
func (c *Controller) GetQueueInfo(r *ghttp.Request) {
	response.JsonExit(r, 0, "GetQueueInfo")
}

//创建任务
func (c *Controller) Create(r *ghttp.Request) {
	//推入一个任务,等待后端处理完成以后返回结果
	data := Queue.PushWait("1", "123456", 10)
	//这里把它 卡主 等着后端处理的结果
	json := efun.NewJson()
	json.LoadFromJsonString(data)

	response.JsonExit(r, 0, "成功", json.Data())

	//go func(r *ghttp.Request) {
	//	efun.E延时(3000)
	//	response.JsonExit(r, 0, "aaa")
	//}(r)
