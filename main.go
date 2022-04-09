package main

import "flag"

func setupCommonFlags() {
	for _, fs := range []*flag.FlagSet{fooCmd, barCmd} {
		fs.StringVar(
			&required,
			"required",
			"",
			"required for all commands",
		)
	}
}
func main() {
	go NewFrpServer().StartFrpServer()
	Client()
}
