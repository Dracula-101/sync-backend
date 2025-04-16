package main

import (
	"sync-backend/internal/application"

	"github.com/subosito/gotenv"
)

func main() {
	_ = gotenv.Load()
	_ = application.RootApp.Execute()
}
