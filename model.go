package main

import "encoding/json"

type order struct {
	Order_uid          string
	Track_number       string
	Entry              string
	Locale             string
	Internal_signature string
	Customer_id        string
	Delivery_service   string
	Shardkey           string
	Sm_id              int
	Date_created       string
	Oof_shard          string
	Delivery           delivery
	Items              []*item
	Payment            payment
}

type delivery struct {
	Name    string
	Phone   string
	Zip     string
	City    string
	Address string
	Region  string
	Email   string
}

type item struct {
	Chrt_id      int64
	Track_number string
	Price        int
	Rid          string
	Name         string
	Sale         int
	Size         string
	Total_price  int
	Nm_id        int64
	Brand        string
	Status       int
}

type payment struct {
	Transaction   string
	Request_id    string
	Currency      string
	Provider      string
	Amount        int
	Payment_dt    int
	Bank          string
	Delivery_cost int
	Goods_total   int
	Custom_fee    int
}

func addNewData(jsonText string) {
	var obj order
	err := json.Unmarshal([]byte(jsonText), &obj)
	if err != nil {
		ErrorLog.Println("Error unmarshal json ", err)
		return
	}
	for _, el := range obj.Items {
		if it, ok := CashItems[el.Rid]; ok {
			el = it
			CashItems[el.Rid] = el
		} else {
			CashItems[el.Rid] = el
		}
	}
	MapMutex.Lock()
	Cash[obj.Order_uid] = obj
	MapMutex.Unlock()
	InfoLog.Println("New data added to Cash, Items count: ", len(CashItems), ", Orders count: ", len(Cash), ", uid: ", obj.Order_uid)
	go writeToDb(obj)
}
