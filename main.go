package main

import "server/util"

func main() {
	server := new(util.BlockAnalysis)
	server.Run()
}
