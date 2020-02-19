package Service

import (
	"fmt"
	"github.com/CHH/eventemitter"
	. "github.com/duolabmeng6/efun/efun"
	"github.com/gogf/gf/database/gredis"
	"github.com/gogf/gf/frame/g"
)

type ApiQueue struct {
	redisConn *gredis.Conn
	emitter   *eventemitter.EventEmitter
}

//初始化消息队列
func NewApiQuue() *ApiQueue {
	api := new(ApiQueue)
	api.Init()

	api.emitter = eventemitter.New()

	go func() {
		fmt.Printf("开始订阅数据")
		conn := g.Redis().Conn()
		defer conn.Close()

		_, err := conn.Do("SUBSCRIBE", "channel_return")
		if err != nil {
			panic(err)
		}
		for {
			reply, err := conn.ReceiveVar()
			if err != nil {
				panic(err)
			}
			message := reply.Vars()[0].String()
			channel := reply.Vars()[1].String()
			value := reply.Vars()[2].String()

			fmt.Printf("%s channel:%s value:%s \r\n", message, channel, value)
			json := NewJson()
			ok := json.LoadFromJsonString(value)

			fun := json.GetString("fun")
			data := json.GetString("data")
			fmt.Printf("%s fun:%s data:%s \r\n", ok, fun, data)

			api.emitter.Emit("call", data)
			//这里如何回调到 PushWait 的函数中
		}
	}()

	return api
}

//初始化消息队列
func (this *ApiQueue) Init() *ApiQueue {
	this.redisConn = g.Redis().Conn()
	return this
}

//初始化消息队列
func (this *ApiQueue) PushWait(key string, senddata string) string {
	fmt.Printf("收到任务 id:%s 任务数据:%s \r\n", key, senddata)

	data := g.Map{}
	data["fun"] = key
	data["data"] = senddata

	ret, _ := this.redisConn.Do("lpush", "queue_test", data)
	fmt.Printf("lpush结果%v \r\n", ret)
	this.emitter.On("call", func(value string) {
		fmt.Printf("收到返回的结果了 %s", value)
	})

	//我要在这里拿到 后端处理好的结果 然后返回
	return "怎么返回啊"
}

//取出任务
func (this *ApiQueue) Pop() {

}

//回到函数
func (this *ApiQueue) Callfun() {

}
