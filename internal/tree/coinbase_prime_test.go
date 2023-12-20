/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package tree

import (
	"testing"
)

var coinbasePrimeRoutes = []Route{
	{"POST", "/v1/allocations", 25},
	{"POST", "/v1/allocations/net", 25},
	{"GET", "/v1/portfolios/:portfolio_id/allocations", 25},
	{"GET", "/v1/portfolios/:portfolio_id/allocations/:allocation_id", 25},
	{"GET", "/v1/portfolios/:portfolio_id/allocations/net/:netting_id", 25},
	{"GET", "/v1/entities/:entity_id/invoices", 25},
	{"GET", "/v1/entities/:entity_id/assets", 25},
	{"GET", "/v1/entities/:entity_id/payment-methods/:payment_method_id", 25},
	{"GET", "/v1/entities/:entity_id/users", 25},
	{"GET", "/v1/portfolios/:portfolio_id/users", 25},
	{"GET", "/v1/portfolios", 25},
	{"GET", "/v1/portfolios/:portfolio_id", 25},
	{"GET", "/v1/portfolios/:portfolio_id/credit", 25},
	{"GET", "/v1/portfolios/:portfolio_id/activities", 25},
	{"GET", "/v1/portfolios/:portfolio_id/activities/:activity_id", 25},
	{"GET", "/v1/portfolios/:portfolio_id/address_book", 25},
	{"POST", "/v1/portfolios/:portfolio_id/address_book", 25},
	{"GET", "/v1/portfolios/:portfolio_id/balances", 25},
	{"GET", "/v1/portfolios/:portfolio_id/wallets/:wallet_id/balance", 25},
	{"GET", "/v1/portfolios/:portfolio_id/commission", 25},
	{"GET", "/v1/portfolios/:portfolio_id/open_orders", 25},
	{"POST", "/v1/portfolios/:portfolio_id/order", 25},
	{"POST", "/v1/portfolios/:portfolio_id/order_preview", 25},
	{"GET", "/v1/portfolios/:portfolio_id/orders", 25},
	{"GET", "/v1/portfolios/:portfolio_id/orders/:order_id", 25},
	{"POST", "/v1/portfolios/:portfolio_id/orders/:order_id/cancel", 25},
	{"GET", "/v1/portfolios/:portfolio_id/orders/:order_id/fills", 25},
	{"GET", "/v1/portfolios/:portfolio_id/products", 25},
	{"GET", "/v1/portfolios/:portfolio_id/transactions", 25},
	{"GET", "/v1/portfolios/:portfolio_id/transactions/{transaction_id}", 25},
	{"POST", "/v1/portfolios/:portfolio_id/wallets/:wallet_id/conversion", 25},
	{"GET", "/v1/portfolios/:portfolio_id/wallets/:wallet_id/transactions", 25},
	{"POST", "/v1/portfolios/:portfolio_id/wallets/:wallet_id/transfers", 25},
	{"POST", "/v1/portfolios/:portfolio_id/wallets/:wallet_id/withdrawals", 25},
	{"GET", "/v1/portfolios/:portfolio_id/wallets", 25},
	{"POST", "/v1/portfolios/:portfolio_id/wallets", 25},
	{"GET", "/v1/portfolios/:portfolio_id/wallets/:wallet_id", 25},
	{"GET", "/v1/portfolios/:portfolio_id/wallets/:wallet_id/deposit_instructions", 25},
}

func TestIndexCoinbasePrimeRoutes(t *testing.T) {
	testIndexRoutes(t, coinbasePrimeRoutes)
}

func TestDumpCoinbasePrimeRoutes(t *testing.T) {
	testDumpRoutes(t, coinbasePrimeRoutes)
}
