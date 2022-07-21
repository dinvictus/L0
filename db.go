package main

import (
	"database/sql"
	"sync"

	_ "github.com/lib/pq"
)

const CONN_STR = "user=Dmitry password=1711 dbname=Service sslmode=disable"
const QUER_ORDER = "INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard, delivery_uid, payment_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)"
const QUER_DELIVERY = "INSERT INTO delivery (name, phone, zip, city, address, region, email) VALUES ($1, $2, $3, $4, $5, $6, $7) returning delivery_uid"
const QUER_PAYMENNT = "INSERT INTO payments (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) returning payment_id"
const QUER_ORDERS_ITEMS = "INSERT INTO orders_items (order_uid, item_id) VALUES ($1, $2)"
const QUER_ITEM = "INSERT INTO items (chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) ON CONFLICT (rid) DO NOTHING"

const QUER_GET_DATA = "SELECT ord.order_uid, ord.track_number, entry, locale, internal_signature, customer_id, delivery_service, sm_id, shardkey, date_created, oof_shard, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee, del.name, phone, city, address, region, email, zip, chrt_id, it.track_number, price, rid, it.name, sale, size, total_price, nm_id, brand, status FROM orders ord, payments pay, delivery del, items it, orders_items ord_it WHERE ord.delivery_uid = del.delivery_uid AND ord.payment_id = pay.payment_id AND ord.order_uid = ord_it.order_uid AND it.rid = ord_it.item_id"

var db *sql.DB

func initDb() {
	var err error
	db, err = sql.Open("postgres", CONN_STR)
	if err != nil {
		ErrorLog.Println("Error connect to database: ", err)
	}
	InfoLog.Println("Connected to database")
	restoreCash()
}

func restoreCash() {
	Cash = make(map[string]order)
	CashItems = make(map[string]*item)
	rows, err := db.Query(QUER_GET_DATA)
	if err != nil {
		ErrorLog.Println("Error restore cash: ", err)
		return
	}
	for rows.Next() {
		tempObj := order{}
		tempObjItem := new(item)
		err := rows.Scan(&tempObj.Order_uid, &tempObj.Track_number, &tempObj.Entry, &tempObj.Locale, &tempObj.Internal_signature, &tempObj.Customer_id, &tempObj.Delivery_service, &tempObj.Sm_id, &tempObj.Shardkey, &tempObj.Date_created, &tempObj.Oof_shard, &tempObj.Payment.Transaction, &tempObj.Payment.Request_id, &tempObj.Payment.Currency, &tempObj.Payment.Provider, &tempObj.Payment.Amount, &tempObj.Payment.Payment_dt, &tempObj.Payment.Bank, &tempObj.Payment.Delivery_cost, &tempObj.Payment.Goods_total, &tempObj.Payment.Custom_fee, &tempObj.Delivery.Name, &tempObj.Delivery.Phone, &tempObj.Delivery.City, &tempObj.Delivery.Address, &tempObj.Delivery.Region, &tempObj.Delivery.Email, &tempObj.Delivery.Zip, &tempObjItem.Chrt_id, &tempObjItem.Track_number, &tempObjItem.Price, &tempObjItem.Rid, &tempObjItem.Name, &tempObjItem.Sale, &tempObjItem.Size, &tempObjItem.Total_price, &tempObjItem.Nm_id, &tempObjItem.Brand, &tempObjItem.Status)
		if item, ok := CashItems[tempObjItem.Rid]; ok {
			tempObjItem = item
		} else {
			CashItems[tempObjItem.Rid] = tempObjItem
		}
		tempObj.Items = append(tempObj.Items, tempObjItem)
		if obj, ok := Cash[tempObj.Order_uid]; ok {
			obj.Items = append(obj.Items, tempObjItem)
			Cash[tempObj.Order_uid] = obj
		} else {
			Cash[tempObj.Order_uid] = tempObj
		}
		if err != nil {
			ErrorLog.Println("Error scan row: ", err)
			continue
		}
	}
	InfoLog.Println("Cash restored")
}

func writeToDb(obj order) {
	wg := new(sync.WaitGroup)
	chanDelivery := make(chan int64)
	chanPayment := make(chan int64)
	tx, err := db.Begin()
	if err != nil {
		ErrorLog.Println("Error to begin transaction: ", err)
		return
	}
	wg.Add(2)
	go writePaymentAndDelivery(obj.Payment, obj.Delivery, chanPayment, chanDelivery, tx, wg)
	go writeOrder(obj, chanDelivery, chanPayment, tx, wg)
	wg.Wait()
	errTx := tx.Commit()
	if errTx != nil {
		ErrorLog.Println("Error add data to database: ", err)
		return
	}
	InfoLog.Println("Data addded to database")

}

func writeOrder(obj order, chDel chan int64, chPay chan int64, tx *sql.Tx, wg *sync.WaitGroup) {
	var payId int64 = <-chPay
	var delId int64 = <-chDel
	writeItems(obj.Items, tx)
	_, err := tx.Exec(QUER_ORDER, obj.Order_uid, obj.Track_number, obj.Entry, obj.Locale, obj.Internal_signature, obj.Customer_id, obj.Delivery_service, obj.Shardkey, obj.Sm_id, obj.Date_created, obj.Oof_shard, delId, payId)
	if err != nil {
		ErrorLog.Println("Error write order to database: ", err)
		tx.Rollback()
	} else {
		writeOrderItems(obj, tx)
	}
	close(chDel)
	close(chPay)
	wg.Done()
}

func writePaymentAndDelivery(objPay payment, objDel delivery, chPay chan int64, chDel chan int64, tx *sql.Tx, wg *sync.WaitGroup) {
	var idPay int64
	errPay := tx.QueryRow(QUER_PAYMENNT, objPay.Transaction, objPay.Request_id, objPay.Currency, objPay.Provider, objPay.Amount, objPay.Payment_dt, objPay.Bank, objPay.Delivery_cost, objPay.Goods_total, objPay.Custom_fee).Scan(&idPay)
	if errPay != nil {
		ErrorLog.Println("Error write payment to database: ", errPay)
		tx.Rollback()
	} else {
		chPay <- idPay
	}
	var idDel int64
	errDel := tx.QueryRow(QUER_DELIVERY, objDel.Name, objDel.Phone, objDel.Zip, objDel.City, objDel.Address, objDel.Region, objDel.Email).Scan(&idDel)
	if errDel != nil {
		ErrorLog.Println("Error write delivery to database: ", errDel)
		tx.Rollback()
	} else {
		chDel <- idDel
	}
	wg.Done()
}

func writeItems(arrObj []*item, tx *sql.Tx) {
	for _, item := range arrObj {
		_, err := tx.Exec(QUER_ITEM, item.Chrt_id, item.Track_number, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.Total_price, item.Nm_id, item.Brand, item.Status)
		if err != nil {
			ErrorLog.Println("Error write item to database: ", err)
			tx.Rollback()
		}
	}
}

func writeOrderItems(obj order, tx *sql.Tx) {
	for _, item := range obj.Items {
		_, err := tx.Exec(QUER_ORDERS_ITEMS, obj.Order_uid, item.Rid)
		if err != nil {
			ErrorLog.Println("Error write orders_items info to database: ", err)
			tx.Rollback()
		}
	}
}
