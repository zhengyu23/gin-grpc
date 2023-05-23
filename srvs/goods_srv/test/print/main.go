package main

import (
	"fmt"
	"log"
)

type s struct {
	name string
}

type test struct {
	X s
	Y int
}

func main() {
	x := s{name: "我是"}
	a := test{X: x, Y: 5}
	fmt.Println(a)
	log.Println(a)
}
