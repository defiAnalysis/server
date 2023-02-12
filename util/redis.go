package util

import (
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

//var redisclient *redis.ClusterClient
var redisclient *redis.Client

type Redis struct {
}

//func init() {
//	redisclient = redis.NewClusterClient(&redis.ClusterOptions{
//		Addrs: strings.Split(Get("redis.url"),","),
//		Password:Get("redis.pass"),
//		MinIdleConns:100,
//		ReadTimeout:3*time.Second,
//		PoolSize:1000,
//	})
//	err := redisclient.Ping().Err()
//	if err != nil {
//		log.Fatal("Redis cluster wrong ",err)
//	}
//}

//
func init() {
	db,_ := strconv.Atoi(ReadStringConcif("redis.db"))
	redisclient = redis.NewClient(&redis.Options{
		Addr:         ReadStringConcif("redis.addr"),
		Password:     ReadStringConcif("redis.password"),
		ReadTimeout:  time.Second * time.Duration(3),
		WriteTimeout: time.Second * time.Duration(3),
		IdleTimeout:  time.Second * time.Duration(60),
		PoolSize:     64,
		MinIdleConns: 16,
		DB:           db,
	})
	err := redisclient.Ping().Err()
	if err != nil {
		log.Fatal("Redis client wrong ", err)
	}
}

func (m *Redis) RedisSet(key string, val interface{}, exp time.Duration) (string, error) {
	for {
		if err := redisclient.Ping().Err(); err == nil {
			return redisclient.Set(key, val, exp).Result()
		} else {
			log.Println("添加redis ["+key+"] 异常：", err)
			continue
		}
	}
}

func (m *Redis) RedisGet(key string) string {
	for {
		if err := redisclient.Ping().Err(); err == nil {
			s, _ := redisclient.Get(key).Result()
			return s
		} else {
			log.Println("读取redis ["+key+"] 异常：", err)
			continue
		}
	}
}

func (m *Redis) RedisExists(key string) (int64) {
	for {
		if err := redisclient.Ping().Err(); err == nil {
			s,_ := redisclient.Exists(key).Result()
			return s
		} else {
			log.Println("读取redis ["+key+"] 异常：", err)
			continue
		}
	}
}
