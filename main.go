package main

import (
	"fmt"
	"os"
	cmd "todo-cli/cmd"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("please choose between web or cli")
		return
	}
	rCmd := os.Args[1]
	if rCmd == "" {
		fmt.Println("please choose between web or cli")
		return
	}
	cmds := map[string]func(){
		"cli": cmd.RunCli,
	}
	c, ok := cmds[rCmd]
	if !ok {
		fmt.Println("please choose between web or cli")
	} else {
		c()
	}
}
