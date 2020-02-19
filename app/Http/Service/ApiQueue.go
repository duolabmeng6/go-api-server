package Service

import (
	"fmt"
	"github.com/CHH/eventemitter"
	. "github.com/duolabmeng6/efun/efun"
	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/container/gvar"
	"github.com/gogf/gf/database/gredis"
	"github.com/gogf/gf/frame/g"
	"time"
)

type ApiQueue struct {
	redisConn *gredis.Conn
	emitter   *eventemitter.EventEmitter
	Hash      *gmap.Map
	keychan   map[string]chan string
}

//初始化消息队列
func NewApiQuue() *ApiQueue {
	api := new(ApiQueue)
	api.Init()

	api.emitter = eventemitter.New()
	api.Hash = gmap.New()

	api.keychan = map[string]chan string{}

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

			//api.emitter.Emit("call", data)
			//这里如何回调到 PushWait 的函数中

			//api.Hash.Set("call", data)

			api.keychan[fun] <- data

		}
	}()

	//获取任务
	go func() {
		conn := g.Redis().Conn()
		defer conn.Close()

		t := time.NewTicker(time.Second * 3)
		for {
			select {
			case <-t.C:
				fmt.Println("定时检查队列中的任务 去完成它")
				Popdata := api.Pop().String()
				if Popdata == "" {
					fmt.Println("没有任务呢")
					break
				}
				json := NewJson()
				json.LoadFromJsonString(Popdata)

				fun := json.GetString("fun")
				data := json.GetString("data")

				fmt.Printf("Pop fun: %s data: %s  \r\n", fun, data)

				if data == "" {
					break
				}
				//t.Stop()
				rejson := NewJson()
				rejson.Set("fun", fun)
				//rejson.Set("data", "ok reload"+data)
				rejson.Set("data.name", "hello")
				rejson.Set("data.age", "10")
				fmt.Println("完成任务", rejson.ToJson(false))

				_, err := conn.Do("publish", "channel_return", rejson.ToJson(false))
				if err != nil {
					panic(err)
				}

			}
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
func (this *ApiQueue) PushWait(key string, senddata string, timeOut int) string {
	data := NewJson()
	data.Set("fun", key)
	data.Set("data", senddata)
	data.Set("timeOut", timeOut)
	data.Set("time", E取时间戳())

	fmt.Printf("将任务推入队列 %s \r\n", data.ToJson(true))

	ret, _ := this.redisConn.Do("lpush", "queue_test", data.ToJson(false))
	fmt.Printf("lpush结果%v \r\n", ret)
	//this.emitter.On("call", func(value string) {
	//	fmt.Printf("收到返回的结果了 %s", value)
	//})
	//var value string
	//for value == "" {
	//	value = this.Hash.GetVar("call").String()
	//	fmt.Printf("call结果%v \r\n", value)
	//	E延时(100)
	//}
	//this.Hash.Remove("call")

	//注册监听通道

	this.keychan[key] = make(chan string)
	var value string
	for {
		breakFlag := false
		select {
		case data := <-this.keychan[key]:
			fmt.Println("所有完成 done", data)
			value = data
			breakFlag = true
		case <-time.After(time.Duration(timeOut) * time.Second):
			fmt.Println("超时")
			breakFlag = true
		}
		if breakFlag {
			delete(this.keychan, key)
			break
		}
	}

	//我要在这里拿到 后端处理好的结果 然后返回
	return value
}

//取出任务
func (this *ApiQueue) Pop() *gvar.Var {

	ret, _ := this.redisConn.DoVar("lpop", "queue_test")

	fmt.Printf("Pop :%s \r\n", ret)

	return ret
}

//回到函数
func (this *ApiQueue) Callfun() {

}
