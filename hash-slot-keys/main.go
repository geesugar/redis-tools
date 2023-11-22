package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("args error: slotcount  ip:port")
		return
	}

	// 创建 Redis 客户端
	cli := redis.NewClient(&redis.Options{
		Addr: os.Args[1],
	})

	ctx := context.Background()
	length := 16384
	array := make([]int, length)
	remaining := length

	rand.Seed(time.Now().UnixNano())

	for remaining > 0 {
		// Generate a random key
		key := randStringBytes(10) // Generate a random string of length 10

		res := cli.ClusterKeySlot(ctx, key)
		if res.Err() != nil {
			fmt.Println(res.Err())
			return
		}

		index := res.Val()

		// If this is the first time the slot is incremented, decrease the remaining count
		if array[index] == 0 {
			remaining--

			fmt.Printf("slot-%d-{%s}\n", index, key)
		}

		// Increment the corresponding array value
		array[index]++
	}
}

// randStringBytes generates a random string of n length
func randStringBytes(n int) string {
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
