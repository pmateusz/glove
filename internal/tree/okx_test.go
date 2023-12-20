/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package tree

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// 600 credits replenish for 1 second with upper bound of 6000

var okxRoutes = []Route{
	{"GET", "/api/v5/account/balance", 120},
	{"GET", "/api/v5/account/positions", 120},
	{"GET", "/api/v5/account/positions-history", 6000},
	{"GET", "/api/v5/account/account-position-risk", 120},
	{"GET", "/api/v5/account/bills", 120},
	{"GET", "/api/v5/account/bills-archive", 240},
	{"GET", "/api/v5/account/config", 240},
	{"POST", "/api/v5/account/set-position-mode", 240},
	{"POST", "/api/v5/account/set-leverage", 60},
	{"GET", "/api/v5/account/max-size", 60},
	{"GET", "/api/v5/account/max-avail-size", 60},
	{"POST", "/api/v5/account/position/margin-balance", 60},
	{"GET", "/api/v5/account/leverage-info", 60},
	{"GET", "/api/v5/account/adjust-leverage-info", 240},
	{"GET", "/api/v5/account/max-loan", 60},
	{"GET", "/api/v5/account/trade-fee", 240},
	{"GET", "/api/v5/account/interest-accrued", 240},
	{"GET", "/api/v5/account/interest-rate", 240},
	{"POST", "/api/v5/account/set-greeks", 240},
	{"POST", "/api/v5/account/set-isolated-mode", 240},
	{"GET", "/api/v5/account/max-withdrawal", 60},
	{"GET", "/api/v5/account/risk-state", 120},
	{"POST", "/api/v5/account/quick-margin-borrow-repay", 240},
	{"GET", "/api/v5/account/quick-margin-borrow-repay-history", 240},
	{"POST", "/api/v5/account/borrow-repay", 100},
	{"GET", "/api/v5/account/borrow-repay-history", 240},
	{"GET", "/api/v5/account/vip-interest-accrued", 240},
	{"GET", "/api/v5/account/vip-interest-deducted", 240},
	{"GET", "/api/v5/account/vip-loan-order-list", 240},
	{"GET", "/api/v5/account/vip-loan-order-detail", 240},
	{"GET", "/api/v5/account/interest-limits", 240},
	{"POST", "/api/v5/account/simulated_margin", 600},
	{"GET", "/api/v5/account/greeks", 120},
	{"GET", "/api/v5/account/position-tiers", 120},
	{"POST", "/api/v5/account/set-riskOffset-type", 120},
	{"POST", "/api/v5/account/activate-option", 240},
	{"POST", "/api/v5/account/set-auto-loan", 240},
	{"POST", "/api/v5/account/set-account-level", 240},
	{"POST", "/api/v5/account/mmp-reset", 240},
	{"POST", "/api/v5/account/mmp-config", 3000},
	{"GET", "/api/v5/account/mmp-config", 240},
	{"POST", "/api/v5/trade/order", 20},
	{"POST", "/api/v5/trade/batch-orders", 80},
	{"POST", "/api/v5/trade/cancel-order", 20},
	{"POST", "/api/v5/trade/cancel-batch-orders", 80},
	{"POST", "/api/v5/trade/amend-order", 20},
	{"POST", "/api/v5/trade/amend-batch-orders", 80},
	{"POST", "/api/v5/trade/close-position", 60},
	{"GET", "/api/v5/trade/order", 20},
	{"GET", "/api/v5/trade/orders-pending", 20},
	{"GET", "/api/v5/trade/orders-history", 30},
	{"GET", "/api/v5/trade/orders-history-archive", 60},
	{"GET", "/api/v5/trade/fills", 20},
	{"GET", "/api/v5/trade/fills-history", 120},
	{"GET", "/api/v5/trade/easy-convert-currency-list", 1200},
	{"POST", "/api/v5/trade/easy-convert", 1200},
	{"GET", "/api/v5/trade/easy-convert-history", 1200},
	{"GET", "/api/v5/trade/one-click-repay-currency-list", 1200},
	{"POST", "/api/v5/trade/one-click-repay", 1200},
	{"GET", "/api/v5/trade/one-click-repay-history", 1200},
	{"POST", "/api/v5/trade/mass-cancel", 240},
	{"POST", "/api/v5/trade/cancel-all-after", 600},
	{"POST", "/api/v5/trade/order-algo", 60},
	{"POST", "/api/v5/trade/cancel-algos", 60},
	{"POST", "/api/v5/trade/amend-algos", 60},
	{"POST", "/api/v5/trade/cancel-advance-algos", 60},
	{"GET", "/api/v5/trade/order-algo", 60},
	{"GET", "/api/v5/trade/orders-algo-pending", 60},
	{"GET", "/api/v5/trade/orders-algo-history", 60},
	{"POST", "/api/v5/tradingBot/grid/order-algo", 60},
	{"POST", "/api/v5/tradingBot/grid/amend-order-algo", 60},
	{"POST", "/api/v5/tradingBot/grid/stop-order-algo", 60},
	{"POST", "/api/v5/tradingBot/grid/close-position", 60},
	{"POST", "/api/v5/tradingBot/grid/cancel-close-order", 60},
	{"POST", "/api/v5/tradingBot/grid/order-instant-trigger", 60},
	{"GET", "/api/v5/tradingBot/grid/orders-algo-pending", 60},
	{"GET", "/api/v5/tradingBot/grid/orders-algo-history", 60},
	{"GET", "/api/v5/tradingBot/grid/orders-algo-details", 60},
	{"GET", "/api/v5/tradingBot/grid/sub-orders", 60},
	{"GET", "/api/v5/tradingBot/grid/positions", 60},
	{"POST", "/api/v5/tradingBot/grid/withdraw-income", 60},
	{"POST", "/api/v5/tradingBot/grid/compute-margin-balance", 60},
	{"POST", "/api/v5/tradingBot/grid/margin-balance", 60},
	{"GET", "/api/v5/tradingBot/grid/ai-param", 60},
	{"POST", "/api/v5/tradingBot/grid/min-investment", 60},
	{"GET", "/api/v5/tradingBot/public/rsi-back-testing", 60},
	{"POST", "/api/v5/tradingBot/recurring/order-algo", 60},
	{"POST", "/api/v5/tradingBot/recurring/amend-order-algo", 60},
	{"POST", "/api/v5/tradingBot/recurring/stop-order-algo", 60},
	{"GET", "/api/v5/tradingBot/recurring/orders-algo-pending", 60},
	{"GET", "/api/v5/tradingBot/recurring/orders-algo-history", 60},
	{"GET", "/api/v5/tradingBot/recurring/orders-algo-details", 60},
	{"GET", "/api/v5/tradingBot/recurring/sub-orders", 60},
	{"GET", "/api/v5/copytrading/current-subpositions", 60},
	{"GET", "/api/v5/copytrading/subpositions-history", 60},
	{"POST", "/api/v5/copytrading/algo-order", 60},
	{"POST", "/api/v5/copytrading/close-subposition", 60},
	{"GET", "/api/v5/copytrading/instruments", 240},
	{"POST", "/api/v5/copytrading/set-instruments", 240},
	{"GET", "/api/v5/copytrading/profit-sharing-details", 240},
	{"GET", "/api/v5/copytrading/total-profit-sharing", 240},
	{"GET", "/api/v5/copytrading/unrealized-profit-sharing-details", 240},
	{"GET", "/api/v5/market/tickers", 60},
	{"GET", "/api/v5/market/ticker", 60},
	{"GET", "/api/v5/market/books", 30},
	{"GET", "/api/v5/market/books-lite", 100},
	{"GET", "/api/v5/market/candles", 30},
	{"GET", "/api/v5/market/history-candles", 60},
	{"GET", "/api/v5/market/trades", 12},
	{"GET", "/api/v5/market/history-trades", 120},
	{"GET", "/api/v5/market/option/instrument-family-trades", 60},
	{"GET", "/api/v5/public/option-trades", 60},
	{"GET", "/api/v5/rfq/counterparties", 240},
	{"POST", "/api/v5/rfq/create-rfq", 240},
	{"POST", "/api/v5/rfq/cancel-rfq", 240},
	{"POST", "/api/v5/rfq/cancel-batch-rfqs", 600},
	{"POST", "/api/v5/rfq/cancel-all-rfqs", 600},
	{"POST", "/api/v5/rfq/execute-quote", 400},
	{"GET", "/api/v5/rfq/maker-instrument-settings", 240},
	{"POST", "/api/v5/rfq/maker-instrument-settings", 240},
	{"POST", "/api/v5/rfq/mmp-reset", 240},
	{"POST", "/api/v5/rfq/create-quote", 24},
	{"POST", "/api/v5/rfq/cancel-quote", 24},
	{"POST", "/api/v5/rfq/cancel-batch-quotes", 600},
	{"POST", "/api/v5/rfq/cancel-all-quotes", 600},
	{"GET", "/api/v5/rfq/rfqs", 600},
	{"GET", "/api/v5/rfq/quotes", 600},
	{"GET", "/api/v5/rfq/trades", 240},
	{"GET", "/api/v5/rfq/public-trades", 240},
	{"GET", "/api/v5/market/block-tickers", 60},
	{"GET", "/api/v5/market/block-ticker", 60},
	{"GET", "/api/v5/market/block-trades", 60},
	{"POST", "/api/v5/sprd/order", 60},
	{"POST", "/api/v5/sprd/cancel-order", 60},
	{"POST", "/api/v5/sprd/mass-cancel", 120},
	{"GET", "/api/v5/sprd/order", 60},
	{"GET", "/api/v5/sprd/orders-pending", 120},
	{"GET", "/api/v5/sprd/orders-history", 60},
	{"GET", "/api/v5/sprd/trades", 60},
	{"GET", "/api/v5/sprd/spreads", 60},
	{"GET", "/api/v5/sprd/books", 60},
	{"GET", "/api/v5/sprd/ticker", 60},
	{"GET", "/api/v5/sprd/public-trades", 60},
	{"GET", "/api/v5/public/instruments", 60},
	{"GET", "/api/v5/public/delivery-exercise-history", 30},
	{"GET", "/api/v5/public/open-interest", 60},
	{"GET", "/api/v5/public/funding-rate", 60},
	{"GET", "/api/v5/public/funding-rate-history", 120},
	{"GET", "/api/v5/public/price-limit", 60},
	{"GET", "/api/v5/public/opt-summary", 60},
	{"GET", "/api/v5/public/estimated-price", 120},
	{"GET", "/api/v5/public/discount-rate-interest-free-quota", 600},
	{"GET", "/api/v5/public/time", 120},
	{"GET", "/api/v5/public/mark-price", 120},
	{"GET", "/api/v5/public/position-tiers", 120},
	{"GET", "/api/v5/public/interest-rate-loan-quota", 600},
	{"GET", "/api/v5/public/vip-interest-rate-loan-quota", 600},
	{"GET", "/api/v5/public/underlying", 60},
	{"GET", "/api/v5/public/insurance-fund", 120},
	{"GET", "/api/v5/public/convert-contract-coin", 120},
	{"GET", "/api/v5/public/instrument-tick-bands", 240},
	{"GET", "/api/v5/market/index-tickers", 60},
	{"GET", "/api/v5/market/index-candles", 60},
	{"GET", "/api/v5/market/history-index-candles", 120},
	{"GET", "/api/v5/market/mark-price-candles", 60},
	{"GET", "/api/v5/market/history-mark-price-candles", 120},
	{"GET", "/api/v5/market/open-oracle", 3000},
	{"GET", "/api/v5/market/exchange-rate", 1200},
	{"GET", "/api/v5/market/index-components", 60},
	{"GET", "/api/v5/rubik/stat/trading-data/support-coin", 240},
	{"GET", "/api/v5/rubik/stat/taker-volume", 240},
	{"GET", "/api/v5/rubik/stat/margin/loan-ratio", 240},
	{"GET", "/api/v5/rubik/stat/contracts/long-short-account-ratio", 240},
	{"GET", "/api/v5/rubik/stat/contracts/open-interest-volume", 240},
	{"GET", "/api/v5/rubik/stat/option/open-interest-volume", 240},
	{"GET", "/api/v5/rubik/stat/option/open-interest-volume-ratio", 240},
	{"GET", "/api/v5/rubik/stat/option/open-interest-volume-expiry", 240},
	{"GET", "/api/v5/rubik/stat/option/open-interest-volume-strike", 240},
	{"GET", "/api/v5/rubik/stat/option/taker-block-volume", 240},
	{"GET", "/api/v5/asset/currencies", 100},
	{"GET", "/api/v5/asset/balances", 100},
	{"GET", "/api/v5/asset/non-tradable-assets", 100},
	{"GET", "/api/v5/asset/asset-valuation", 600},
	{"POST", "/api/v5/asset/transfer", 600},
	{"GET", "/api/v5/asset/transfer-state", 600},
	{"GET", "/api/v5/asset/bills", 100},
	{"GET", "/api/v5/asset/deposit-lightning", 300},
	{"GET", "/api/v5/asset/deposit-address", 100},
	{"GET", "/api/v5/asset/deposit-history", 100},
	{"POST", "/api/v5/asset/withdrawal", 100},
	{"POST", "/api/v5/asset/withdrawal-lightning", 300},
	{"POST", "/api/v5/asset/cancel-withdrawal", 100},
	{"GET", "/api/v5/asset/withdrawal-history", 100},
	{"GET", "/api/v5/asset/deposit-withdraw-status", 1200},
	{"POST", "/api/v5/asset/convert-dust-assets", 600},
	{"GET", "/api/v5/asset/exchange-list", 100},
	{"GET", "/api/v5/asset/convert/currencies", 100},
	{"GET", "/api/v5/asset/convert/currency-pair", 100},
	{"POST", "/api/v5/asset/convert/estimate-quote", 60},
	{"POST", "/api/v5/asset/convert/trade", 60},
	{"GET", "/api/v5/asset/convert/history", 100},
	{"GET", "/api/v5/users/subaccount/list", 600},
	{"POST", "/api/v5/users/subaccount/modify-apikey", 600},
	{"GET", "/api/v5/account/subaccount/balances", 200},
	{"GET", "/api/v5/asset/subaccount/balances", 200},
	{"GET", "/api/v5/account/subaccount/max-withdrawal", 60},
	{"GET", "/api/v5/asset/subaccount/bills", 100},
	{"GET", "/api/v5/asset/subaccount/managed-subaccount-bills", 100},
	{"POST", "/api/v5/asset/subaccount/transfer", 600},
	{"POST", "/api/v5/users/subaccount/set-transfer-out", 600},
	{"GET", "/api/v5/users/entrust-subaccount-list", 600},
	{"GET", "/api/v5/users/partner/if-rebate", 60},
	{"POST", "/api/v5/account/subaccount/set-loan-allocation", 240},
	{"GET", "/api/v5/account/subaccount/interest-limits", 240},
	{"GET", "/api/v5/finance/staking-defi/offers", 200},
	{"POST", "/api/v5/finance/staking-defi/purchase", 300},
	{"POST", "/api/v5/finance/staking-defi/redeem", 300},
	{"POST", "/api/v5/finance/staking-defi/cancel", 300},
	{"GET", "/api/v5/finance/staking-defi/orders-active", 200},
	{"GET", "/api/v5/finance/staking-defi/orders-history", 200},
	{"GET", "/api/v5/finance/savings/balance", 100},
	{"POST", "/api/v5/finance/savings/purchase-redempt", 100},
	{"POST", "/api/v5/finance/savings/set-lending-rate", 100},
	{"GET", "/api/v5/finance/savings/lending-history", 100},
	{"GET", "/api/v5/finance/savings/lending-rate-summary", 100},
	{"GET", "/api/v5/finance/savings/lending-rate-history", 100},
	{"GET", "/api/v5/system/status", 100},
}

func testIndexRoutes(t *testing.T, routes []Route) {
	// GIVEN
	a := assert.New(t)
	var index tree

	// WHEN
	for _, r := range routes {
		index.add(r.method, r.route, r.cost)
	}

	// THEN
	for _, r := range routes {
		rootNode := index.get(r.method)
		if a.NotNil(rootNode, "method: "+r.method) {
			cost := rootNode.get(r.route)
			a.Equal(r.cost, cost, r.route)
		}
	}
}

func testDumpRoutes(t *testing.T, routes []Route) {
	// GIVEN
	a := assert.New(t)
	var index tree

	// WHEN
	for _, route := range routes {
		index.add(route.method, route.route, route.cost)
	}

	// THEN
	var filteredExpectedRoutes []Route
	for _, route := range routes {
		if route.cost > 0 {
			filteredExpectedRoutes = append(filteredExpectedRoutes, route)
		}
	}

	expectedIndex := indexByRoute(filteredExpectedRoutes)
	actualRoutes := index.dump()
	actualIndex := indexByRoute(actualRoutes)

	// THEN
	for key, expectedRoute := range expectedIndex {
		actualRoute, ok := actualIndex[key]
		if a.True(ok, "method: %s, route: %s", expectedRoute.method, expectedRoute.route) {
			a.Equal(expectedRoute, actualRoute, "method: %s, route: %s", expectedRoute.method, expectedRoute.route)
		}
	}
}

func indexByRoute(routes []Route) map[string]Route {
	index := make(map[string]Route)
	for _, route := range routes {
		index[fmt.Sprintf("%s:%s", route.method, route.route)] = route
	}
	return index
}

func TestIndexOKXRoutes(t *testing.T) {
	testIndexRoutes(t, okxRoutes)
}

func TestDumpOKXRoutes(t *testing.T) {
	testDumpRoutes(t, okxRoutes)
}
