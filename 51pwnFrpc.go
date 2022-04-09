package main

import (
	_ "github.com/fatedier/frp/assets/frpc"
	"github.com/fatedier/frp/cmd/frpc/sub"
)

func Client() {
	sub.Execute()
}

// func main() {
// 	client()
// }
