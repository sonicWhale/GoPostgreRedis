package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"strconv"
	"sync"
)

var Client *redis.Client
var wg sync.WaitGroup

func waiter(c chan int) {
	wg.Wait()
	fmt.Println("exit....")
	c <- 1
}

func main() {

	//Подключаемся к Redis
	Client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	out := make(chan DevicesStruct)
	GetAllDevicesFromDB(out)
	waitme := make(chan int)
	go waiter(waitme)

	ex := 0
	for ex == 0 {
		select {
		case x := <-out:
			go func(x DevicesStruct) {
				allMetrics := CreateMetric(x)
				go func(m DevicesMetricStruct) {
					checkMetrics(m)
					wg.Done()
				}(allMetrics)
			}(x)
		case <-waitme:
			close(out)
			close(waitme)
			ex = 1
		}
	}
}

//Установить значения по ключу
func setValues(key int, value string) {
	keyToStr := strconv.Itoa(key)
	err := Client.Set(keyToStr, value, 0).Err()
	if err != nil {
		panic(err)
	}
}

//Получить значения по ключу
func getValues(key int) string {
	keyToStr := strconv.Itoa(key)
	val, err := Client.Get(keyToStr).Result()
	if err != nil {
		//panic(err)
	}
	return val
}
