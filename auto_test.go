package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/nats-io/stan.go"
)

const TEST_DATA = `{
	"order_uid": "testUid",
	"track_number": "WBILMTESTTRACK",
	"entry": "WBIL",
	"delivery": {
	  "name": "Test Testov",
	  "phone": "+9720000000",
	  "zip": "2639809",
	  "city": "Kiryat Mozkin",
	  "address": "Ploshad Mira 15",
	  "region": "Kraiot",
	  "email": "test@gmail.com"
	},
	"payment": {
	  "transaction": "b563feb7b2b84b6test",
	  "request_id": "",
	  "currency": "USD",
	  "provider": "wbpay",
	  "amount": 1817,
	  "payment_dt": 1637907727,
	  "bank": "alpha",
	  "delivery_cost": 1500,
	  "goods_total": 317,
	  "custom_fee": 0
	},
	"items": [
	  {
		"chrt_id": 9934930,
		"track_number": "WBILMTESTTRACK",
		"price": 453,
		"rid": "testRidItem1",
		"name": "Mascaras",
		"sale": 30,
		"size": "0",
		"total_price": 317,
		"nm_id": 2389212,
		"brand": "Vivienne Sabo",
		"status": 202
	  },
	  {
		"chrt_id": 9934930,
		"track_number": "WBILMTESTTRACK",
		"price": 453,
		"rid": "testRidItem2",
		"name": "Mascaras",
		"sale": 30,
		"size": "0",
		"total_price": 317,
		"nm_id": 2389212,
		"brand": "Vivienne Sabo",
		"status": 202
	  }
	],
	"locale": "en",
	"internal_signature": "",
	"customer_id": "test",
	"delivery_service": "meest",
	"shardkey": "9",
	"sm_id": 99,
	"date_created": "2021-11-26T06:22:19Z",
	"oof_shard": "1"
  }`

func TestModel(t *testing.T) {
	var testObj = order{}
	err := json.Unmarshal([]byte(TEST_DATA), &testObj)
	if err != nil {
		t.Error("Error unmarshal json to object: ", err)
	} else if testObj.Order_uid != "testUid" && testObj.Payment.Transaction != "b563feb7b2b84b6test" && testObj.Delivery.Zip != "2639809" && len(testObj.Items) != 2 && testObj.Items[0].Rid != "testRidItem1" && testObj.Items[1].Rid != "testRidItem2" {
		t.Error("Error create valid struct from json: ")
	}
}

func TestDB(t *testing.T) {
	db, errDb := sql.Open("postgres", CONN_STR)
	if errDb != nil {
		t.Error("Error connect to database: ", errDb)
		return
	}
	defer db.Close()
	tx, errTx := db.Begin()
	if errTx != nil {
		t.Error("Error begin transaction: ", errTx)
	}
	rows, errRow := tx.Query("SELECT * FROM orders LIMIT 1")
	if errRow != nil {
		tx.Rollback()
		t.Error("Error get rows from database: ", errRow)
		return
	}
	if !rows.Next() {
		t.Error("Error parse rows")
	}
	errTxCommit := tx.Commit()
	if errTxCommit != nil {
		t.Error("Error commit transaction: ", errTxCommit)
	}

}

func TestReciever(t *testing.T) {
	sc, errStan := stan.Connect(STAN_CLUSTER_ID, "TestClient")
	if errStan != nil {
		t.Error("Error connect to nats-streaming: ", errStan)
		return
	}
	defer sc.Close()
	ch, err := sc.Subscribe("Channel", func(m *stan.Msg) {
	}, stan.DurableName("TestDurable"))
	if err != nil {
		t.Error("Error subscribe on channel: ", err)
		return
	}
	defer ch.Unsubscribe()
}

func TestTemplate(t *testing.T) {
	testObj := order{}
	funcMap := template.FuncMap{
		"add": func(i int) int {
			return i + 1
		},
	}
	ts, errTmpl := template.New("mainpage").Funcs(funcMap).ParseFiles("./html/mainpage.tmpl")
	if errTmpl != nil {
		t.Error("Error create template: ", errTmpl)
		return
	}
	err := ts.Execute(os.Stdout, testObj)
	if err != nil {
		t.Error("Error execute template: ", err)
	}
}
