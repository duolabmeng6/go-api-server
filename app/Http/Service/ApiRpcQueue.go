package Service

import (
	"encoding/json"
	. "github.com/duolabmeng6/efun/efun"
	. "github.com/duolabmeng6/efun/src/utils"
	"github.com/gogf/gf/database/gredis"
	"github.com/gogf/gf/frame/g"
	"sync"
	"time"
)

//适用于 api接口 rpc调用 消息队列模式

// PushWait 推入任务并且等待处理结果
// Pop 获取任务
// Callfun 回调处理结果
// Info 获取任务信息

type ApiRpcQueue struct {
	//redis客户端
	redisConn *gredis.Redis
	//等待消息回调的通道
	keychan map[string]chan string
	//redis订阅频道名称
	channel string
	//redis队列名称
	queueName string
	//读写锁用于keychan的
	lock sync.RWMutex
}

//调用任务的结构
type TaskData struct {
	//任务id 回调函数id
	Fun string `json:"fun"`
	//任务数据
	Data string `json:"data"`
	//加入任务时间
	StartTime int `json:"start_time"`
	//超时时间
	TimeOut int `json:"timeout"`
	//执行完成结果
	Result string `json:"result"`
	//完成时间
	CompleteTime int `json:"complete_time"`
	//发布频道
	Channel string `json:"channel"`
	//队列的名称
	Queue string `json:"queue"`
}

//初始化消息队列
func NewApiRpcQueue(queue_name string) *ApiRpcQueue {
	this := new(ApiRpcQueue)
	this.Init()

	this.keychan = map[string]chan string{}
	//随机生成一个频道的用于订阅回调数据
	this.channel = "channel_" + Euuidv4()
	this.queueName = queue_name

	//用于监听数据,任务完成以后将数据实时回调
	go func() {
		//E调试输出格式化("开始订阅数据 channel %s", this.channel)
		//新建一个 redis连接
		conn := g.Redis().Conn()

		_, err := conn.Do("SUBSCRIBE", this.channel)
		if err != nil {
			panic(err)
		}
		for {
			reply, err := conn.ReceiveVar()
			if err != nil {
				panic(err)
			}
			//message := reply.Vars()[0].String()
			//channel := reply.Vars()[1].String()
			value := reply.Vars()[2].String()

			//E调试输出格式化("%s channel:%s value:%s \r\n", message, channel, value)

			taskData := &TaskData{}
			json.Unmarshal([]byte(value), &taskData)
			data := taskData.Result
			fun := taskData.Fun

			//通过go的chan回调数据

			this.lock.RLock()
			funchan, ok := this.keychan[fun]
			this.lock.RUnlock()
			if ok {
				funchan <- data
			} else {
				//E调试输出格式化("fun not find %s", fun)

			}

		}
	}()

	return this
}

//初始化消息队列
func (this *ApiRpcQueue) Init() *ApiRpcQueue {
	this.redisConn = g.Redis()
	return this
}

//推入任务并且等待结果
func (this *ApiRpcQueue) PushWait(key string, senddata string, timeOut int, priority int) (string, bool) {
	taskData := TaskData{}
	//任务id
	taskData.Fun = key
	//任务数据
	taskData.Data = senddata
	//超时时间 1.pop 取出任务超时了 就放弃掉 2.任务在规定时间内未完成 超时 退出
	taskData.TimeOut = timeOut
	//任务加入时间
	taskData.StartTime = int(E取时间戳())
	//任务完成以后回调的频道名称
	taskData.Channel = this.channel

	taskData.Queue = this.queueName + "_" + E到文本(priority)

	jsondata, _ := json.Marshal(taskData)
	//E调试输出格式化("加入任务 %s \r\n", jsondata)

	//加入任务
	if this.push(string(jsondata), taskData.Queue, priority) == false {
		return "push error", false
	}
	value, flag := this.waitResult(key, timeOut)
	//E调试输出格式化("waitResult %v %s \r\n", flag, value)
	return value, flag
}

//等待任务结果
func (this *ApiRpcQueue) waitResult(key string, timeOut int) (string, bool) {
	//注册监听通道
	this.lock.Lock()
	this.keychan[key] = make(chan string)
	this.lock.Unlock()

	var value string

	breakFlag := false
	timeOutFlag := false
	for {
		select {
		case data := <-this.keychan[key]:
			//收到结果放进去
			value = data
			breakFlag = true
		case <-time.After(time.Duration(timeOut) * time.Second):
			//超时跳出并且删除
			breakFlag = true
			timeOutFlag = true
		}
		if breakFlag {
			break
		}
	}
	//将通道的key删除
	this.lock.Lock()
	delete(this.keychan, key)
	this.lock.Unlock()

	if timeOutFlag {
		return "time out", false
	}
	return value, true
}

//加入任务
func (this *ApiRpcQueue) push(data string, queueName string, priority int) bool {
	_, err := this.redisConn.Do("lpush", queueName, data)
	if err != nil {
		E调试输出("Push Error", err.Error())
	}
	return err == nil
}

//取出任务
//brpop 先进先出
//返回值 任务数据 TaskData 状态值 0 没有任务 1获取成功 2任务是超时的删除无需执行
func (this *ApiRpcQueue) Pop() (*TaskData, int) {
	taskData := &TaskData{}
	//这不支持优先级
	//ret, _ := this.redisConn.DoVar("lpop", this.queueName)
	//支持优先级处理
	ret, _ := this.redisConn.DoVar("brpop", this.queueName+"_2", this.queueName+"_1", this.queueName+"_0", 10)
	if ret.String() == "" {
		return taskData, 0
	}
	//E调试输出格式化("Pop %s \r\n", ret.Strings()[1])
	json.Unmarshal([]byte(ret.Strings()[1]), &taskData)
	//E调试输出格式化("Pop %s \r\n", taskData)
	if taskData.StartTime+taskData.TimeOut < int(E取时间戳()) {
		//E调试输出格式化("任务超时抛弃 %s %s \r\n", ret.Strings()[0], taskData)
		this.redisConn.Do("incr", ret.Strings()[0]+"_timeout_count")

		return taskData, 2
	}
	this.redisConn.Do("incr", ret.Strings()[0]+"_pop_count")

	return taskData, 1
}

//回到函数
func (this *ApiRpcQueue) Callfun(taskData *TaskData) {
	taskData.CompleteTime = int(E取时间戳())

	jsondata, _ := json.Marshal(taskData)
	//E调试输出("\r\n 通知任务完成", string(jsondata))
	//E调试输出("\r\n Channel", taskData.Channel)

	this.redisConn.Do("incr", taskData.Queue+"_complete_count")

	_, err := this.redisConn.Do("publish", taskData.Channel, jsondata)
	if err != nil {
		panic(err)
	}
}

//队列中相关信息
func (this *ApiRpcQueue) Info() interface{} {
	//this.queueName + "_2", this.queueName + "_1", this.queueName + "_0"

	strarr := New文本型数组()
	for i := 0; i <= 2; i++ {
		strarr.E加入成员(this.queueName + "_" + E到文本(i))
	}

	jsonData := NewJson()

	for i := 0; i < strarr.E取数组成员数(); i++ {
		llen, _ := this.redisConn.DoVar("llen", strarr.Get(i))
		list, _ := this.redisConn.DoVar("lrange", strarr.Get(i), 0, 10)
		timeout_count, _ := this.redisConn.DoVar("get", strarr.Get(i)+"_timeout_count")
		pop_count, _ := this.redisConn.DoVar("get", strarr.Get(i)+"_pop_count")
		complete_count, _ := this.redisConn.DoVar("get", strarr.Get(i)+"_complete_count")

		jsonData.Set(strarr.Get(i)+".count", llen.Int())
		jsonData.Set(strarr.Get(i)+".timeout_count", timeout_count)
		jsonData.Set(strarr.Get(i)+".pop_count", pop_count)
		jsonData.Set(strarr.Get(i)+".complete_count", complete_count)

		for _, data := range list.Strings() {
			//E调试输出(data)

			TaskData := TaskData{}
			json.Unmarshal([]byte(data), &TaskData)

			jsonData.SetArray(strarr.Get(i)+".list", TaskData)
		}
	}

	//E调试输出(jsonData.ToJson(true))

	return jsonData.Data()
}
