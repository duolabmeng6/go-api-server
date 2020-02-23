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

var Queue = Service.NewApiRpcQueue("queue_ip_query")
var ApiLog = Service.NewApiRpcLog()

//获取队列中的任务
func (c *Controller) Get(r *ghttp.Request) {

	flag := 2
	var data *Service.TaskData
	for flag == 2 {
		data, flag = Queue.Pop()
	}

	if flag == 1 {
		ApiLog.SetStatus(data.Fun, 1)
		response.JsonExit(r, 200, "获取任务", data)
	} else if flag == 0 {
		response.JsonExit(r, 201, "没有任务")
	} else if flag == 2 {
		ApiLog.SetStatus(data.Fun, 2)

		response.JsonExit(r, 202, "任务超时无需执行")
	}

}

//提交处理后的任务数据composer require tymon/jwt-auth:1.0.0-rc.3
func (c *Controller) Put(r *ghttp.Request) {
	taskData := Service.TaskData{}
	taskData.Fun = E到文本(r.Get("fun"))
	taskData.Result = E到文本(r.Get("result"))
	taskData.Channel = E到文本(r.Get("channel"))
	taskData.Queue = E到文本(r.Get("queue"))
	//E调试输出格式化("任务完成推送结果 PutQueue  fun:%s result:%s \r\n", taskData.Fun, taskData.Result)

	Queue.Callfun(&taskData)

	response.JsonExit(r, 200, "ok")
}

//获取队列中的状态信息
func (c *Controller) Info(r *ghttp.Request) {
	data := Queue.Info()
	response.JsonExit(r, 200, "队列信息", data)
}

//创建任务
func (c *Controller) Create(r *ghttp.Request) {
	parameter := E到文本(r.Get("data"))

	uuid := Euuidv4()
	//E调试输出("生成任务id", uuid)
	//task, flag := Queue.PushTask(uuid, parameter, 5, E取随机数(0, 2))
	task, flag := Queue.PushTask(uuid, parameter, 5, 2)
	ApiLog.Put(task)

	task, flag = Queue.WaitResult(task)

	var taskLog *Service.TaskDataModel
	if flag == false {
		//E调试输出("失败了", data)
		taskLog = ApiLog.Put_complete(task, 2)

		response.JsonExit(r, 0, "失败", task.Result)
		return
	}
	taskLog = ApiLog.Put_complete(task, 3)

	//E调试输出格式化("收到任务结果  耗时 %s \r\n完成任务数据 %s \r\n", time.E取毫秒(), data)

	json := NewJson()
	json.LoadFromJsonString(task.Result)
	json.Set("time", taskLog.ProcessTime)
	json.Set("fun", taskLog.Fun)
	response.JsonExit(r, 200, "成功", json.Data())
}

//提交处理后的任务数据composer require tymon/jwt-auth:1.0.0-rc.3
func (c *Controller) Result(r *ghttp.Request) {
	Fun := E到文本(r.Get("fun"))

	task, flag := ApiLog.Find(Fun)

	if flag {
		json := NewJson()
		json.LoadFromJsonString(task.Result)
		json.Set("time", task.ProcessTime)
		json.Set("fun", task.Fun)

		response.JsonExit(r, 200, "成功", json.Data())
	} else {
		response.JsonExit(r, 201, "失败")
	}
}
