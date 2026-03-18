//go:build wireinject
// +build wireinject

package main

import (
	"net/http"

	"github.com/google/wire"
)

//go:generate wire

func InitializeServer() (*http.Server, func(), error) {
	wire.Build(
		NewLogger,
		ProvidePort,
		ProvideDB,
		NewPurchaseRepository,
		NewRequestGateway,
		NewTransactionUsecase,
		NewRouter,
		NewHTTPServer,
	)
	return nil, nil, nil
}
