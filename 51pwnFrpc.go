package main

import (
	"github.com/hktalent/frp/cmd/frpc/sub"
)

type FrpClient struct{}

func (c *FrpClient) Client() {
	sub.Execute()
}

// func main() {
// 	client()
// }
