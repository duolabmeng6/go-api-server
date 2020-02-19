package main

import (
	"github.com/gogf/gf/frame/g"
	_ "laravel-go/boot"
	_ "laravel-go/router"
)

func main() {
	g.Server().Run()

}
