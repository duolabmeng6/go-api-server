package Service

import (
	. "github.com/duolabmeng6/efun/efun"
	"github.com/gogf/gf/database/gredis"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/util/gconv"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"sync"
)

type TaskDataModel struct {
	gorm.Model

	//任务id 回调函数id
	Fun string `json:"fun" gorm:"type:varchar(100);unique_index"`
	//任务数据
	Data string `json:"data" gorm:"type:varchar(1000)"`
	//加入任务时间
	StartTime int64 `json:"start_time"`
	//超时时间
	TimeOut int `json:"timeout"`
	//执行完成结果
	Result string `json:"result" gorm:"type:varchar(1000)"`
	//完成时间
	CompleteTime int64 `json:"complete_time"`
	//发布频道
	Channel string `json:"channel"`
	//队列的名称
	Queue string `json:"queue"`

	//状态 0 任务加入 1任务弹出处理中 2任务超时不处理 3任务完成
	Status int `json:"status"`
	//过程耗时
	ProcessTime int64 `json:"process_time"`
}

func (TaskDataModel) TableName() string {
	return "task_data_log"
}

type ApiRpcLog struct {
	//redis客户端
	redisConn *gredis.Redis
	lock      sync.RWMutex
	db        *gorm.DB
}

func NewApiRpcLog() *ApiRpcLog {
	this := new(ApiRpcLog)
	this.Init()

	return this
}

//初始化消息队列
func (this *ApiRpcLog) Init() *ApiRpcLog {
	this.redisConn = g.Redis()
	var err error
	this.db, err = gorm.Open("sqlite3", "test.db?cache=shared")
	if err != nil {
		E调试输出(err)
		panic("连接数据库失败")
	}
	this.db.AutoMigrate(&TaskDataModel{})

	return this
}

func (this *ApiRpcLog) Put(data *TaskData) bool {
	taskDataModel := TaskDataModel{}

	err := gconv.Struct(data, &taskDataModel)
	if err != nil {
		panic(err)
		return false
	}

	// 创建
	this.lock.Lock()
	defer this.lock.Unlock()

	this.db.Create(&taskDataModel)

	return true
}

func (this *ApiRpcLog) Put_complete(data *TaskData, Status int) *TaskDataModel {
	taskDataModel := TaskDataModel{}

	err := gconv.Struct(data, &taskDataModel)
	if err != nil {
		panic(err)
		return &taskDataModel
	}
	//E调试输出P(taskDataModel)
	taskDataModel.Status = Status
	// 创建
	this.lock.Lock()
	defer this.lock.Unlock()

	taskDataModel.CompleteTime = E取毫秒()
	taskDataModel.ProcessTime = taskDataModel.CompleteTime - taskDataModel.StartTime

	this.db.Model(&taskDataModel).
		Where("fun = ?", taskDataModel.Fun).
		Update(taskDataModel)

	return &taskDataModel
}

func (this *ApiRpcLog) SetStatus(fun string, status int) bool {
	taskDataModel := TaskDataModel{}

	this.lock.Lock()
	defer this.lock.Unlock()
	this.db.Model(&taskDataModel).
		Where("fun = ?", fun).
		Update("status", status)
	return true
}

func (this *ApiRpcLog) Find(fun string) (*TaskDataModel, bool) {
	taskDataModel := TaskDataModel{}
	// 创建

	flag := this.db.
		Where("fun = ?", fun).
		First(&taskDataModel).RecordNotFound()

	return &taskDataModel, flag == false

}
