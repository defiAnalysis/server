package util

import (
	"encoding/json"
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/onrik/ethrpc"
)

type BlockAnalysis struct {
	Redis
	MQ
	lastheight  int
	blockheight int
	block       chan int
	transaction chan ethrpc.Transaction
	hash        chan string
	client      chan *ethrpc.EthRPC
}

func (ba *BlockAnalysis) initClient() {
	rpcs := ReadSliceConcif("chain.rpc")
	for {
		index := rand.Intn(len(rpcs))

		client := ethrpc.New(rpcs[index])
		if _, err := client.EthBlockNumber(); err != nil {
			log.Println("节点客户端初始化异常：", err)
			time.Sleep(3 * time.Second)

			continue
		} else {
			ba.client <- client
			break
		}
	}
}

//初始化交易信道
func (ba *BlockAnalysis) Run() {
	ba.client = make(chan *ethrpc.EthRPC, 100)
	ba.InitAddr()
	ba.initClient()
	ba.InitHeight()

	ba.transaction = make(chan ethrpc.Transaction, 100)
	ba.block = make(chan int, 100)
	ba.hash = make(chan string, 10)

	go ba.monitorBlock()
	go ba.analysisBlock()
	go ba.analysisTransaction()
	//定时初始化以太坊客户端
	go func() {
		for {
			ba.initClient()
			time.Sleep(3 * time.Second)
		}
	}()
	select {}
}

//初始化监控地址
func (ba *BlockAnalysis) InitAddr() {
	addrs := ReadSliceConcif("chain.addr")

	for _,addr := range addrs {
		addr = strings.ToLower(addr)
		ba.RedisSet(addr,1,-1)
	}
}

//初始化块高
func (ba *BlockAnalysis) InitHeight() {
	ba.lastheight, _ = strconv.Atoi(ba.RedisGet("height_bsctest"))
	newclient := <-ba.client
	ba.blockheight, _ = newclient.EthBlockNumber()
	if ba.lastheight == 0 {
		ba.lastheight = ba.blockheight
	}
}

//扫描区块
func (ba *BlockAnalysis) monitorBlock() {
	for {
		log.Println("ba.lastheight:", ba.lastheight)
		if ba.lastheight < ba.blockheight {
			for ba.lastheight < ba.blockheight {
				ba.block <- ba.lastheight
				ba.lastheight++
				time.Sleep(1 * time.Millisecond*500)
			}

			go ba.RedisSet("height_bsctest", ba.lastheight, -1)
		}

		newclient := <-ba.client
		ba.blockheight, _ = newclient.EthBlockNumber()
		time.Sleep(2 * time.Second)
	}
}

//解析区块
func (ba *BlockAnalysis) analysisBlock() {
	for {
		if runtime.NumGoroutine() < 1000 {
			blockheight := <-ba.block
			go func() {
			A:
				for {
					newclient := getClient()
					if block, e := newclient.EthGetBlockByNumber(blockheight, true); e != nil {
						log.Println("区块解析失败，打回：", blockheight)
						time.Sleep(3 * time.Second)
						continue A
					} else {
						if block != nil {
							log.Println("Height: ", "[", blockheight, "]")
							for _, tran := range block.Transactions {
								tran.Gas = block.Timestamp
								ba.transaction <- tran
							}
							break A
						} else {
							continue A
						}
					}
				}
			}()
		}
	}
}

//解析交易
func (ba *BlockAnalysis) analysisTransaction() {
	for {
		tx := <-ba.transaction
		go func(tran ethrpc.Transaction) {
			to := strings.ToLower(tran.To)
			if len(tran.Input) >= 10 && ba.RedisExists(to) > 0 {

				go ba.validTransaction(tx.Hash)
			}
		}(tx)
	}
}

//截取字符串-将字符串前面的0去掉
func SubByZero(str string) string {
	arr := []byte(str)
	for i, str1 := range arr {
		if str1 != 48 {
			str = string(arr[i:])
			break
		}
	}
	return str
}

func findAddr(adds []string, addr string) bool {
	flag := false
	for i := 0; i < len(adds); i++ {
		if strings.ToLower(adds[i]) == strings.ToLower(addr) {
			flag = true
			break
		}
	}

	return flag
}

//校验交易
func (ba *BlockAnalysis) validTransaction(hash string) {
	for {
		client := <-ba.client
		if receipt, e := client.EthGetTransactionReceipt(hash); e != nil {
			time.Sleep(2 * time.Second)
			log.Println("e:", e)
			continue
		} else {
			if receipt.Status == "0x1" && len(receipt.Logs) > 0 {
				bytes, _ := json.Marshal(receipt.Logs)
				//推送到消息队列
				if err := ba.publish(string(bytes)); err != nil {
					time.Sleep(2 * time.Second)
					log.Println("err:", err)
					continue
				} else {
					log.Println("hash:", hash)
					log.Println("=======string(bytes):", string(bytes))
					return
				}

			} else {
				return
			}
		}
	}
}

func getClient() *ethrpc.EthRPC {
	rpcs := ReadSliceConcif("chain.rpc")

	index := rand.Intn(len(rpcs))

	return ethrpc.New(rpcs[index])
}
