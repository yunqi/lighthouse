package main

import (
	"github.com/yunqi/lighthouse/internal/server"
	"github.com/yunqi/lighthouse/internal/xlog"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	xlog.Dev()
	newServer := server.NewServer(server.WithTcpListen(":1883"))
	newServer.ServeTCP()

	//select {}
}
