package main

import (
	_ "embed"
	"github.com/go-playground/validator/v10"
	"github.com/yunqi/lighthouse/config"
	"github.com/yunqi/lighthouse/internal/server"
	"github.com/yunqi/lighthouse/internal/xlog"
	"gopkg.in/yaml.v3"
	"net/http"
)

//go:embed config.yaml
var configBytes []byte

func main() {
	c := new(config.Config)
	err := yaml.Unmarshal(configBytes, &c)
	if err != nil {
		panic(err)
	}
	validate := validator.New()
	err = validate.Struct(c)
	if err != nil {
		panic(err)
	}

	err = xlog.InitLogger(&c.Log)
	if err != nil {
		panic(err)
	}
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	newServer := server.NewServer(server.WithTcpListen(":1883"))
	newServer.ServeTCP()
}