package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

func add(i *int, d, min int) int {
	*i += d
	if *i < min {
		*i = min
	}
	return *i
}

func main() {
	t := time.NewTicker(time.Second)
	var utime, stime, cpu int
	for range t.C {
		b, _ := json.Marshal(map[string]interface{}{
			"mem": map[string]interface{}{
				"Heap":  10000 + rand.Intn(2000),
				"Sys":   20000 + rand.Intn(1000),
				"Stack": 3000 + rand.Intn(500),
			},
			"cpu": map[string]interface{}{
				"UTime": add(&utime, 100+rand.Intn(100), 0),
				"STime": add(&stime, 100+rand.Intn(200), 0),
			},
			"Threads": add(&cpu, rand.Intn(10)-4, 1),
		})
		fmt.Printf("%s\n", b)
	}
}
