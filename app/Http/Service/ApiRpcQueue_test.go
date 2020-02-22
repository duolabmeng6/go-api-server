package Service

import (
	. "github.com/duolabmeng6/efun/efun"
	"github.com/gomodule/redigo/redis"
	"testing"
	"time"
)

var Queue = NewApiRpcQueue("queue_api")

func AutoHandle() {
	//获取任务
	go func() {

		t := time.NewTicker(time.Second * 3)
		for {
			select {
			case <-t.C:
				E调试输出格式化("检查是否有任务需要处理 \r\n")
				data, flag := Queue.Pop()
				if flag == 0 {
					E调试输出格式化("没有任务呢")
					break
				}
				E调试输出("正在处理 \r\n", data)

				//t.Stop()
				rejson := NewJson()
				rejson.Set("newname", "your name"+data.Data)
				rejson.Set("name", "hello")
				rejson.Set("age", "10")

				data.Result = rejson.ToJson(false)

				Queue.Callfun(data)
			}
		}
	}()
}
func TestAutoHandle(t *testing.T) {
	AutoHandle()
}

func TestNewUserInfo(t *testing.T) {
	time := New时间统计类()
	time.E开始()

	uuid := Euuidv4()
	E调试输出("生成任务id", uuid)
	data, flag := Queue.PushWait(uuid, "123456", 10, 0)
	if flag == false {
		E调试输出("失败了", data)
	}
	E调试输出格式化("收到任务结果  耗时 %s \r\n 完成任务数据 %s \r\n", time.E取毫秒(), data)

}

func testRedis(t *testing.T) {
	c, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		t.Log("Connect to redis error", err)
		return
	}
	defer c.Close()

	_, err = c.Do("SET", "mykey", "superWang")
	if err != nil {
		t.Log("redis set failed:", err)
	}
	//username, err := redis.String(c.Do("GET", "mykey"))
	//if err != nil {
	//	t.Log("redis get failed:", err)
	//} else {
	//	t.Log("Get mykey:", username)
	//}

	//queue_api_1
	//
	//username, err := redis.Strings(c.Do("brpop", "queue_api_1", 10))
	//if err != nil {
	//	t.Log("redis get failed:", err)
	//} else {
	//	t.Log("Get mykey:", username)
	//}

	psc := redis.PubSubConn{Conn: c}
	psc.Subscribe("example")
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			E调试输出格式化("%s: message: %s\n", v.Channel, v.Data)
		case redis.Subscription:
			E调试输出格式化("%s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			E调试输出("error", v)

			//return v
		}
	}
}
