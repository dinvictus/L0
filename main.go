package main

import (
	"log"
	"os"
	"sync"
)

var MapMutex = sync.RWMutex{}
var Cash map[string]order
var CashItems map[string]*item
var InfoLog *log.Logger = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
var ErrorLog *log.Logger = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

func main() {
	initDb()
	initReciever()
	startHttpServer()
}
