package main

import (
	"fmt"
	"net"
)

func main() {
	ip1 := net.ParseIP("aoeuaoeu")
	fmt.Println(ip1 == nil)
	fmt.Println(len(ip1))
}
