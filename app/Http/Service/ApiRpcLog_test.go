package Service

import (
	. "github.com/duolabmeng6/efun/efun"
	"testing"
)

func TestDb(t *testing.T) {
	ApiLog := NewApiRpcLog()

	flag := ApiLog.Put(&TaskData{
		Fun:          Euuidv4(),
		Data:         "",
		StartTime:    0,
		TimeOut:      0,
		Result:       "",
		CompleteTime: 0,
		Channel:      "",
		Queue:        "",
	})
	E调试输出(flag)

	flag2 := ApiLog.Update(&TaskData{
		Fun:          "3c571325-8d65-4fb7-abb4-2d7fa4c620b1",
		Data:         "aaaaaaaaaaaaaa",
		StartTime:    0,
		TimeOut:      0,
		Result:       "",
		CompleteTime: 0,
		Channel:      "",
		Queue:        "",
	})
	E调试输出(flag2)

	data, flag3 := ApiLog.Find("3c571325-8d65-4fb7-abb4-2d7fa4c620b1")
	E调试输出("查询", flag3, data)

	data2, flag4 := ApiLog.Find("aaa")
	if flag4 {
		E调试输出P("查询2", flag4, data2)
	} else {
		E调试输出P("查询2 失败")
	}
	//db, err := gorm.Open("sqlite3", "test.db")
	//if err != nil {
	//	t.Error(err)
	//	panic("连接数据库失败")
	//}
	//defer db.Close()
	//
	////db.SingularTable(true)
	//
	//// 自动迁移模式
	//db.AutoMigrate(&TaskDataModel{})

	//// 创建
	//db.Create(&TaskDataModel{
	//	Fun:          efun.Euuidv4(),
	//	Data:         "",
	//	StartTime:    123,
	//	TimeOut:      3,
	//	Result:       "123",
	//	CompleteTime: 1233,
	//	Queue:        "1233",
	//})

	// 读取
	//var TaskDataModel TaskDataModel
	//db.First(&TaskDataModel, 1)                // 查询id为1的TaskDataModel
	//db.First(&TaskDataModel, "Fun = ?", "123") // 查询code为l1212的TaskDataModel
	//t.Log(&TaskDataModel)
	//
	//// 更新 - 更新TaskDataModel的price为2000
	//db.Model(&TaskDataModel).Update("Price", 2000)
	//
	//// 删除 - 删除TaskDataModel
	//db.Delete(&TaskDataModel)
}
