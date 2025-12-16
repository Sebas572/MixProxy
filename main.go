package main

import (
	"fmt"
	"mixploy/src/proxy"
)

func main() {
	fmt.Println("Init Mixploy")

	proxy.Start()
}