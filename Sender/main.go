package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/nats-io/stan.go"
)

const LETTERS = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
const STAN_CLUSTER_ID = "test-cluster"
const STAN_CLIENT_ID = "client"

var infoLog *log.Logger = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

func main() {
	sc, err := stan.Connect(STAN_CLUSTER_ID, STAN_CLIENT_ID)
	if err != nil {
		panic(err)
	}
	for {
		sc.Publish("Channel", getRandData())
		time.Sleep(time.Millisecond * 10)
	}
}

func getRandData() []byte {
	rand.Seed(time.Now().UnixNano())
	randObj := order{}
	randObj.Order_uid = getRandString(20)
	randObj.Track_number = getRandString(15)
	randObj.Entry = getRandString(4)
	randObj.Locale = getRandString(2)
	randObj.Internal_signature = getRandString(5)
	randObj.Customer_id = getRandString(10)
	randObj.Delivery_service = getRandString(5)
	randObj.Shardkey = strconv.Itoa(getRandInt(0, 10))
	randObj.Sm_id = getRandInt(10, 100)
	randObj.Date_created = time.Now().Format("2006.01.02 15:04:05")
	randObj.Oof_shard = strconv.Itoa(getRandInt(0, 10))
	randObj.Delivery.Name = getRandString(5) + " " + getRandString(10)
	randObj.Delivery.Phone = "+" + strconv.Itoa(getRandInt(100000000, 999999999))
	randObj.Delivery.Zip = strconv.Itoa(getRandInt(1000000, 9999999))
	randObj.Delivery.City = getRandString(8) + " " + getRandString(10)
	randObj.Delivery.Address = getRandString(10) + " " + getRandString(8) + strconv.Itoa(getRandInt(1, 1000))
	randObj.Delivery.Region = getRandString(15)
	randObj.Delivery.Email = getRandString(5) + "@" + getRandString(4) + ".com"
	randObj.Payment.Transaction = getRandString(20)
	randObj.Payment.Request_id = getRandString(5)
	randObj.Payment.Currency = getRandString(3)
	randObj.Payment.Provider = getRandString(5)
	randObj.Payment.Amount = getRandInt(1, 20000)
	randObj.Payment.Payment_dt = getRandInt(100000000, 999999999)
	randObj.Payment.Bank = getRandString(10)
	randObj.Payment.Delivery_cost = getRandInt(100, 100000)
	randObj.Payment.Goods_total = getRandInt(1, 500)
	randObj.Payment.Custom_fee = getRandInt(0, 10)
	var randIntItems = getRandInt(1, 30)
	for i := 0; i < randIntItems; i++ {
		randObjItem := item{}
		randObjItem.Chrt_id = int64(getRandInt(100000000, 999999999))
		randObjItem.Track_number = getRandString(15)
		randObjItem.Price = getRandInt(1, 50000)
		randObjItem.Rid = getRandString(30)
		randObjItem.Name = getRandString(15)
		randObjItem.Sale = getRandInt(1, 100)
		randObjItem.Size = strconv.Itoa(getRandInt(0, 10))
		randObjItem.Total_price = getRandInt(1, 50000)
		randObjItem.Nm_id = int64(getRandInt(1000000, 9999999))
		randObjItem.Brand = getRandString(10) + " " + getRandString(5)
		randObjItem.Status = getRandInt(100, 500)
		randObj.Items = append(randObj.Items, &randObjItem)
	}
	bytes, _ := json.Marshal(&randObj)
	infoLog.Println("Message send, uid: ", randObj.Order_uid)
	return bytes
}

func getRandString(n int) string {
	bytes := make([]byte, n)
	for i := range bytes {
		bytes[i] = LETTERS[rand.Intn(len(LETTERS))]
	}
	return string(bytes)
}

func getRandInt(min int, max int) int {
	return rand.Intn(max-min) + min
}
