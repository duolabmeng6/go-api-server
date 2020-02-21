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

type ApiQueue struct {
	redisConn *gredis.Redis
	keychan   map[string]chan string
	channel   string
	queueName string
	lock      sync.RWMutex
}
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
	CompleteTime int `json:"completeTime"`
	//发布频道
	Channel string `json:"channel"`
}

//初始化消息队列
func NewApiQueue(queue_name string) *ApiQueue {
	this := new(ApiQueue)
	this.Init()

	this.keychan = map[string]chan string{}
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
func (this *ApiQueue) Init() *ApiQueue {
	this.redisConn = g.Redis()
	return this
}

//初始化消息队列
func (this *ApiQueue) PushWait(key string, senddata string, timeOut int, priority int) (string, bool) {
	taskData := TaskData{}
	////任务id
	taskData.Fun = key
	////任务数据
	taskData.Data = senddata
	////超时时间 1.pop 取出任务超时了 就放弃掉 2.任务在规定时间内未完成 超时 退出
	taskData.TimeOut = timeOut
	////任务加入时间
	taskData.StartTime = int(E取时间戳())
	taskData.Channel = this.channel

	jsondata, _ := json.Marshal(taskData)
	//E调试输出格式化("加入任务 %s \r\n", jsondata)

	//加入任务
	if this.Push(string(jsondata), priority) == false {
		return "push error", false
	}
	value, flag := this.waitResult(key, timeOut)
	//E调试输出格式化("waitResult %v %s \r\n", flag, value)
	return value, flag
}

//加入任务
func (this *ApiQueue) waitResult(key string, timeOut int) (string, bool) {
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
func (this *ApiQueue) Push(data string, priority int) bool {
	_, err := this.redisConn.Do("lpush", this.queueName+"_"+E到文本(priority), data)
	if err != nil {
		E调试输出("Push Error", err.Error())
	}
	return err == nil
}

//取出任务
//返回值 任务数据 TaskData 状态值 0 没有任务 1获取成功 2任务是超时的删除无需执行
func (this *ApiQueue) Pop() (*TaskData, int) {
	taskData := &TaskData{}

	//ret, _ := this.redisConn.DoVar("lpop", this.queueName)
	ret, _ := this.redisConn.DoVar("blpop", this.queueName+"_2", this.queueName+"_1", this.queueName+"_0", 10)
	if ret.String() == "" {
		return taskData, 0
	}
	//E调试输出格式化("Pop %s \r\n", ret.Strings()[1])
	//这一句代码节省了一堆代码...
	json.Unmarshal([]byte(ret.Strings()[1]), &taskData)

	//E调试输出格式化("Pop %s \r\n", taskData)

	if taskData.StartTime+taskData.TimeOut < int(E取时间戳()) {
		E调试输出格式化("任务超时抛弃 %s \r\n", taskData)
		return taskData, 2
	}

	return taskData, 1

}

//回到函数
func (this *ApiQueue) Callfun(taskData *TaskData) {

	//rejson := NewJson()
	//rejson.Set("fun", taskData.Fun)
	//rejson.Set("result", taskData.Result)
	//rejson.Set("start_time", taskData.StartTime)
	//rejson.Set("timeout", taskData.TimeOut)
	//rejson.Set("complete_time", E取时间戳())
	//////E调试输出("通知任务完成", rejson.ToJson(false))
	taskData.CompleteTime = int(E取时间戳())

	jsondata, _ := json.Marshal(taskData)
	////E调试输出("\r\n 通知任务完成", string(jsondata))
	////E调试输出("\r\n Channel", taskData.Channel)

	_, err := this.redisConn.Do("publish", taskData.Channel, jsondata)
	if err != nil {
		panic(err)
	}
}

//队列中相关信息
func (this *ApiQueue) Info() string {
	llen, _ := this.redisConn.DoVar("llen", this.queueName)
	list, _ := this.redisConn.DoVar("lrange", this.queueName, 0, 10)

	jsonData := NewJson()
	jsonData.Set("count", llen.Int())

	for _, data := range list.Strings() {
		//E调试输出(data)

		TaskData := TaskData{}
		json.Unmarshal([]byte(data), &TaskData)

		jsonData.SetArray("list", TaskData)
	}

	//E调试输出(jsonData.ToJson(true))

	return jsonData.ToJson(false)
}
