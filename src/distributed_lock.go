package main

import (
	"github.com/go-redis/redis"
	"fmt"
	"strconv"
	"github.com/samuel/go-zookeeper/zk"
	"time"
	"strings"
)

var client *redis.Client

var zkHandler *zk.Conn

func init() {
	client = redis.NewClient(&redis.Options{
		Addr:     "10.96.90.6:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	zkHandler, _, _ = zk.Connect([]string{"10.96.90.6"}, time.Second) //*10)
}

func ReadAndIncrement(key string) {
	val, err := client.Get(key).Result()
	if err != nil {
		panic(err)
	}
	intVal, _ := strconv.Atoi(val)
	fmt.Printf("A:[key=%v,val=%v]\n", key, intVal)

	err = client.Set(key, fmt.Sprintf("%v", intVal+1), 0).Err()
	if err != nil {
		panic(err)
	}
}

func lock() string {
	n, _ := zkHandler.Create(fmt.Sprintf("%s/lock-", "/didi"), []byte(" "), zk.FlagEphemeral|zk.FlagSequence, zk.WorldACL(zk.PermAll))
	for {
		children, _, _ := zkHandler.Children(fmt.Sprintf("%s", "/didi"))
		tmp := strings.Split(n, "-")
		nNum, _ := strconv.Atoi("1" + tmp[1])
		isLowestNode := true
		for _, child := range children {
			tmp = strings.Split(child, "-")
			childNum, _ := strconv.Atoi("1" + tmp[1])
			if nNum > childNum {
				isLowestNode = false
				break
			}
		}
		if isLowestNode {
			return n
		}
		p := fmt.Sprintf("%v", nNum-1)[1:]
		existed, _, ch, _ := zkHandler.ExistsW(p)
		if existed {
			<-ch
		}
	}
}

func unlock(node string) {
	_, stat, _ := zkHandler.Get(node)
	zkHandler.Delete(node, stat.Version)
}

func main() {
	unstop := make(chan int)

	key := "my_key"
	for i := 0; i < 200; i++ {
		n := lock()
		ReadAndIncrement(key)
		unlock(n)
	}

	val, _ := client.Get(key).Result()
	fmt.Println("A:final val:", val)
	<-unstop
}
