package main

import (
	"fmt"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

var content string

func main() {
	unstop := make(chan int)
	c, _, err := zk.Connect([]string{"10.96.90.6"}, time.Second) //*10)
	if err != nil {
		panic(err)
	}
	bytes, _, ch, err := c.GetW("/didi")
	if err != nil {
		panic(err)
	}
	content = string(bytes)
	go func() {
		for {
			<-ch
			bytes, _, ch, err = c.GetW("/didi")
			if err != nil {
				panic(err)
			}
			content = string(bytes)
		}
	}()
	go func() {
		for {
			fmt.Printf("%v\n", content)
			time.Sleep(2 * time.Second)
		}
	}()

	<-unstop
}
