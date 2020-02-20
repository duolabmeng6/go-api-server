package Service

import (
	. "github.com/duolabmeng6/efun/efun"
	"testing"
	"time"
)

var Queue = NewApiQueue("queue_api")

func AutoHandle() {
	//获取任务
	go func() {

		t := time.NewTicker(time.Second * 3)
		for {
			select {
			case <-t.C:
				E调试输出格式化("检查是否有任务需要处理 \r\n")
				data, flag := Queue.Pop()
				if flag == false {
					E调试输出格式化("没有任务呢")
					break
				}
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
	data, flag := Queue.PushWait(uuid, "123456", 10)
	if flag == false {
		E调试输出("失败了", data)
	}
	E调试输出格式化("收到任务结果  耗时 %s \r\n 完成任务数据 %s \r\n", time.E取毫秒(), data)

}
