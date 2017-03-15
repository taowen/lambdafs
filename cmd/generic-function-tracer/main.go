package main

import (
	"github.com/derekparker/delve/service/debugger"
	"github.com/taowen/function-tracer/infra"
	"os"
	"strconv"
	"errors"
	"fmt"
)


// int3 + ptrace
func main() {
	fmt.Println(main_())
}

func main_() error {
	if len(os.Args) < 2 {
		return errors.New("pid should be provied")
	}
	attachPid, err := strconv.Atoi(os.Args[1])
	if err != nil {
		infra.LogError("parse pid", "pid", os.Args[1], "err", err)
		return err
	}
	debugger_, err := debugger.New(&debugger.Config{
		AttachPid: attachPid,
	})
	if err != nil {
		infra.LogError("new debugger", "err", err)
		return err
	}
	fmt.Println(debugger_.Functions(""))
	return nil
}
