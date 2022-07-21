package main

import (
	"github.com/nats-io/stan.go"
)

const STAN_CLUSTER_ID = "test-cluster"
const STAN_CLIENT_ID = "client2"
const STAN_DURABLE_NAME = "DurableSubscription"

func initReciever() {
	sc, err := stan.Connect(STAN_CLUSTER_ID, STAN_CLIENT_ID)
	if err != nil {
		ErrorLog.Println("Error connect to nats-streaming: ", err)
		return
	}
	InfoLog.Println("Reciever ready")
	recieveMessage(sc)
}

func recieveMessage(sc stan.Conn) {
	_, err := sc.Subscribe("Channel", func(m *stan.Msg) {
		InfoLog.Println("Received a message")
		addNewData(string(m.Data))
	}, stan.DurableName(STAN_DURABLE_NAME))

	if err != nil {
		ErrorLog.Println("Error subscribe on channel: ", err)
	}
}
