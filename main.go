package main

import (
	"log"
	"os"
	"runtime"

	"./app"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var app = app.App{}
	f, err := os.OpenFile("/tmp/go_Service.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	if err != nil {
		log.Fatal("err")
	}
	app.InitLogger(f)
	// app.InitScheduler()
	app.Init()
}

//https://dzone.com/articles/try-and-catch-in-golang
//http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/

//SET GOOS=windows, SET GOARCH=386 for windows
//SET GOOS=linux, SET GOARCH=amd64 for linux
