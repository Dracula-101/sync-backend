package application

import (
	"net/http/httptest"

	"sync-backend/arch/config"
	"sync-backend/arch/network"
)

type Teardown = func()

func TestServer() (network.Router, Module, Teardown) {
	env := config.NewEnv("../.test.env")
	config := config.LoadConfig("../configs")
	router, module, shutdown := create(&env, &config)
	ts := httptest.NewServer(router.GetEngine())
	teardown := func() {
		ts.Close()
		shutdown()
	}
	return router, module, teardown
}
