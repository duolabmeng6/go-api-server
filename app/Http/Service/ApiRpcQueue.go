package Service

import (
	"encoding/json"
	. "github.com/duolabmeng6/efun/efun"
	. "github.com/duolabmeng6/efun/src/utils"
	"github.com/gomodule/redigo/redis"
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
	redisConn *redis.Pool
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
	StartTime int64 `json:"start_time"`
	//超时时间
	TimeOut int64 `json:"timeout"`
	//执行完成结果
	Result string `json:"result"`
	//完成时间
	CompleteTime int64 `json:"complete_time"`
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
		psc := redis.PubSubConn{Conn: this.redisConn.Get()}
		psc.Subscribe(this.channel)
		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				//E调试输出格式化("%s: message: %s\n", v.Channel, v.Data)
				value := v.Data
				taskData := &TaskData{}
				json.Unmarshal([]byte(value), &taskData)
				data := taskData.Result
				fun := taskData.Fun

				this.lock.RLock()
				funchan, ok := this.keychan[fun]
				this.lock.RUnlock()
				if ok {
					funchan <- data
				} else {
					//E调试输出格式化("fun not find %s", fun)
				}
			case redis.Subscription:
				E调试输出格式化("%s: %s %d\n", v.Channel, v.Kind, v.Count)
			case error:
				E调试输出("Subscribe error", v)
				//return v

				psc = redis.PubSubConn{Conn: this.redisConn.Get()}
				psc.Subscribe(this.channel)
			}

		}
	}()

	return this
}

//初始化消息队列
func (this *ApiRpcQueue) Init() *ApiRpcQueue {
	//this.redisConn.Get() = g.Redis()
	//p *Pool
	//this.redisConn.Get(), _ = redis.Dial("tcp", "127.0.0.1:6379")

	this.redisConn = &redis.Pool{
		MaxIdle:     1000,
		MaxActive:   0,
		IdleTimeout: 240 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			con, err := redis.Dial("tcp", "127.0.0.1:6379",
				//redis.DialPassword(conf["Password"].(string)),
				redis.DialDatabase(int(0)),
				redis.DialConnectTimeout(240*time.Second),
				redis.DialReadTimeout(240*time.Second),
				redis.DialWriteTimeout(240*time.Second))
			if err != nil {
				return nil, err
			}
			return con, nil
		},
	}
	return this
}

func (this *ApiRpcQueue) PushTask(key string, senddata string, timeOut int64, priority int) (*TaskData, bool) {
	conn := this.redisConn.Get()
	defer conn.Close()
	conn.Do("incr", "count")

	taskData := TaskData{}
	//任务id
	taskData.Fun = key
	//任务数据
	taskData.Data = senddata
	//超时时间 1.pop 取出任务超时了 就放弃掉 2.任务在规定时间内未完成 超时 退出
	taskData.TimeOut = timeOut
	//任务加入时间
	taskData.StartTime = E取毫秒()
	//任务完成以后回调的频道名称
	taskData.Channel = this.channel

	taskData.Queue = this.queueName + "_" + E到文本(priority)

	jsondata, _ := json.Marshal(taskData)
	//E调试输出格式化("加入任务 %s \r\n", jsondata)

	//加入任务
	if this.push(string(jsondata), taskData.Queue, priority) == false {
		return &taskData, false
	}
	return &taskData, true

}

func (this *ApiRpcQueue) WaitResult(taskData *TaskData) (*TaskData, bool) {
	value, flag := this.waitResult(taskData.Fun, taskData.TimeOut)
	taskData.Result = value
	return taskData, flag
}

//推入任务并且等待结果
func (this *ApiRpcQueue) PushWait(key string, senddata string, timeOut int64, priority int) (string, bool) {
	task, flag := this.PushTask(key, senddata, timeOut, priority)

	task, flag = this.WaitResult(task)

	return task.Result, flag
}

//等待任务结果
func (this *ApiRpcQueue) waitResult(key string, timeOut int64) (string, bool) {
	//注册监听通道
	this.lock.Lock()
	this.keychan[key] = make(chan string)
	mychan := this.keychan[key]
	this.lock.Unlock()

	var value string

	breakFlag := false
	timeOutFlag := false

	for {
		select {

		case data := <-mychan:
			//收到结果放进RUnlock()
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

	conn := this.redisConn.Get()
	defer conn.Close()
	_, err := conn.Do("lpush", queueName, data)
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
	//ret, _ := this.redisConn.Get().DoVar("lpop", this.queueName)
	//支持优先级处理
	conn := this.redisConn.Get()
	defer conn.Close()

	ret, _ := redis.Strings(conn.Do("brpop", this.queueName+"_2", this.queueName+"_1", this.queueName+"_0", 10))
	if len(ret) == 0 {
		return taskData, 0
	}
	//E调试输出格式化("Pop %s \r\n", ret.Strings()[1])
	json.Unmarshal([]byte(ret[1]), &taskData)
	//E调试输出格式化("Pop %s \r\n", taskData)
	//E调试输出格式化("Pop %s \r\n", taskData.StartTime, taskData.StartTime/1000, taskData.TimeOut, E取时间戳())

	if taskData.StartTime/1000+taskData.TimeOut < E取时间戳() {
		//E调试输出格式化("任务超时抛弃 %s %s \r\n", ret[0], taskData)
		//this.redisConn.Get().Do("incr", ret[0]+"_timeout_count")

		return taskData, 2
	}
	//this.redisConn.Get().Do("incr", ret[0]+"_pop_count")

	return taskData, 1
}

//回到函数
func (this *ApiRpcQueue) Callfun(taskData *TaskData) {
	//taskData.CompleteTime = int(E取时间戳())

	jsondata, _ := json.Marshal(taskData)
	//E调试输出("\r\n 通知任务完成", string(jsondata))
	//E调试输出("\r\n Channel", taskData.Channel)
	conn := this.redisConn.Get()
	defer conn.Close()
	_, err := conn.Do("publish", taskData.Channel, jsondata)
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
		llen, _ := this.redisConn.Get().Do("llen", strarr.Get(i))
		list, _ := redis.Strings(this.redisConn.Get().Do("lrange", strarr.Get(i), 0, 10))
		timeout_count, _ := this.redisConn.Get().Do("get", strarr.Get(i)+"_timeout_count")
		pop_count, _ := this.redisConn.Get().Do("get", strarr.Get(i)+"_pop_count")
		complete_count, _ := this.redisConn.Get().Do("get", strarr.Get(i)+"_complete_count")

		jsonData.Set(strarr.Get(i)+".count", llen)
		jsonData.Set(strarr.Get(i)+".timeout_count", timeout_count)
		jsonData.Set(strarr.Get(i)+".pop_count", pop_count)
		jsonData.Set(strarr.Get(i)+".complete_count", complete_count)

		for _, data := range list {
			//E调试输出(data)

			TaskData := TaskData{}
			json.Unmarshal([]byte(data), &TaskData)

			jsonData.SetArray(strarr.Get(i)+".list", TaskData)
		}
	}

	//E调试输出(jsonData.ToJson(true))

	return jsonData.Data()
}
