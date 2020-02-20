package Api

import (
	. "github.com/duolabmeng6/efun/efun"
	. "github.com/duolabmeng6/efun/src/utils"
	"github.com/gogf/gf-demos/library/response"
	"github.com/gogf/gf/net/ghttp"
	"laravel-go/app/Http/Service"
)

// 用户API管理对象
type Controller struct{}

var Queue = Service.NewApiQueue("queue_api")

//获取队列中的任务
func (c *Controller) GetQueue(r *ghttp.Request) {
	data, flag := Queue.Pop()
	if flag == false {
		//E调试输出("没有任务呢")
		response.JsonExit(r, 201, "没有任务")

	} else {
		response.JsonExit(r, 200, "获取任务", data)
	}

}

//提交处理后的任务数据
func (c *Controller) PutQueue(r *ghttp.Request) {
	taskData := Service.TaskData{}
	taskData.Fun = E到文本(r.Get("fun"))
	taskData.Result = E到文本(r.Get("result"))
	taskData.Channel = E到文本(r.Get("channel"))
	E调试输出格式化("任务完成推送结果 PutQueue  fun:%s result:%s \r\n", taskData.Fun, taskData.Result)

	Queue.Callfun(&taskData)

	response.JsonExit(r, 200, "ok")
}

//获取队列中的状态信息
func (c *Controller) GetQueueInfo(r *ghttp.Request) {
	response.JsonExit(r, 0, "GetQueueInfo")
}

//创建任务
func (c *Controller) Create(r *ghttp.Request) {
	parameter := E到文本(r.Get("data"))

	time := New时间统计类()
	time.E开始()

	uuid := Euuidv4()
	E调试输出("生成任务id", uuid)
	data, flag := Queue.PushWait(uuid, parameter, 5)
	if flag == false {
		E调试输出("失败了", data)
		response.JsonExit(r, 0, "失败"+data)
		return
	}
	E调试输出格式化("收到任务结果  耗时 %s \r\n完成任务数据 %s \r\n", time.E取毫秒(), data)
	json := NewJson()
	json.LoadFromJsonString(data)
	response.JsonExit(r, 200, "成功", json.Data())
}
