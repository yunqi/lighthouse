package main

import (
	_ "embed"
	"github.com/go-playground/validator/v10"
	"github.com/yunqi/lighthouse/config"
	_ "github.com/yunqi/lighthouse/internal/persistence/session/memory"
	_ "github.com/yunqi/lighthouse/internal/persistence/session/redis"
	"github.com/yunqi/lighthouse/internal/server"
	"github.com/yunqi/lighthouse/internal/xlog"
	"github.com/yunqi/lighthouse/internal/xtrace"
	"gopkg.in/yaml.v3"
	"net/http"
	_ "net/http/pprof"
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

	xtrace.StartAgent(&c.Trace)

	go func() {
		_ = http.ListenAndServe("localhost:6060", nil)
	}()

	newServer := server.NewServer(server.WithTcpListen(":1883"), server.WithPersistence(&c.Persistence))
	newServer.ServeTCP()
}
