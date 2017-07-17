package main

import (
	"github.com/zanecloud/apiserver/cli"
	_ "net/http/pprof"
	"runtime"
)

func main() {

	runtime.GOMAXPROCS(4)
	cli.Run()
}
