package main

import "github.com/s-frick/exchange/server"

func main() {
	go server.Start()

	select {}
}
