package main

import (
	"html/template"
	"net/http"
)

func startHttpServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", getOrderInfo)
	fileServer := http.FileServer(http.Dir("./html/"))
	mux.Handle("/files/", http.StripPrefix("/files", fileServer))
	InfoLog.Println("Starting server in http://localhost:8080/")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		ErrorLog.Println("Error run http server: ", err)
	}

}

func getOrderInfo(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	r.ParseForm()
	if r.Form.Get("uid") != "" {
		InfoLog.Println("Request order info, uid: ", r.Form.Get("uid"))
	}
	funcMap := template.FuncMap{
		"add": func(i int) int {
			return i + 1
		},
	}
	ts, errTmpl := template.New("mainpage").Funcs(funcMap).ParseFiles("./html/mainpage.tmpl")
	if errTmpl != nil {
		ErrorLog.Println("Error create template: ", errTmpl.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}
	MapMutex.RLock()
	findObj := Cash[r.Form.Get("uid")]
	MapMutex.RUnlock()
	err := ts.Execute(w, findObj)
	if err != nil {
		ErrorLog.Println("Error execute template: ", err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}
