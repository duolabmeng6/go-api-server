package boot

import (
	"github.com/gogf/gf/frame/g"
)

// 用于应用初始化。
func init() {
	s := g.Server()
	//s.SetAccessLogEnabled(true)
	s.SetErrorLogEnabled(true)
}
