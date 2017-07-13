package main

import (
	"github.com/zanecloud/apiserver/cli"
	_ "net/http/pprof"
)

func main() {

	cli.Run()
}
