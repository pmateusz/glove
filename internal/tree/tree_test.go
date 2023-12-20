/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package tree

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRejectsRouteWithoutOpeningSlash(t *testing.T) {
	// GIVEN
	a := assert.New(t)
	route := "one"

	// WHEN
	r, err := validateRoute(route)

	// THEN
	a.Equal("", r)
	a.EqualError(err, "route must start with '/': "+route)
}

func TestRejectsRouteWithWildcardAtWrongPosition(t *testing.T) {
	// GIVEN
	a := assert.New(t)
	route := "/*path"

	// WHEN
	r, err := validateRoute(route)

	// THEN
	a.Equal("", r)
	a.EqualError(err, "wildcard '*' is only allowed at the end of the route: "+route)
}

func TestRejectsRouteWithEmptyParameter(t *testing.T) {
	// GIVEN
	a := assert.New(t)
	route := "/:/path"

	// WHEN
	r, err := validateRoute(route)

	// THEN
	a.Equal("", r)
	a.EqualError(err, "at least one parameter is not named in the route: "+route)
}

func TestRejectsRouteWithMalformedParameter(t *testing.T) {
	// GIVEN
	a := assert.New(t)
	route := "/:one:two/"

	// WHEN
	r, err := validateRoute(route)

	// THEN
	a.Equal("", r)
	a.EqualError(err, "parameter ':two/' must directly follow '/' in the route: "+route)
}

func TestRemovesTrailingSlash(t *testing.T) {
	// GIVEN
	a := assert.New(t)
	route := "/path/"

	// WHEN
	r, err := validateRoute(route)

	// THEN
	a.Equal("/path", r)
	a.NoError(err)
}

type assertions struct {
	*assert.Assertions
}

var method = "GET"

func newAssertions(t *testing.T) *assertions {
	return &assertions{
		assert.New(t),
	}
}

func (a *assertions) Properties(node *node, kind nodeKind, route string, hasWildcard bool, cost int) {
	a.Equal(route, node.route)
	a.Equal(kind, node.kind)

	if hasWildcard {
		a.True(hasWildcard)
	} else {
		a.False(hasWildcard)
	}

	a.Equal(cost, node.cost)
}

func (a *assertions) Index(node *node, index ...byte) {
	a.Equal(index, node.index)
	a.Len(node.children, len(index))
}

func (a *assertions) EmptyIndex(node *node) {
	a.Empty(node.index)
	a.Empty(node.children)
}

func (a *assertions) Leaf(node *node) {
	a.Nil(node.special)
	a.EmptyIndex(node)
}

func TestCanAddOneRoute(t *testing.T) {
	// GIVEN
	a := newAssertions(t)
	r := require.New(t)

	var index tree

	// WHEN
	index.add(method, "/v1/entities/:entity_id/assets", 1)
	rootNode := index.get(method)

	// THEN
	r.NotNil(rootNode)
	a.Properties(rootNode, text, "/v1/entities/", false, 0)
	a.EmptyIndex(rootNode)

	entityNode := rootNode.special
	r.NotNil(entityNode)
	a.Properties(entityNode, parameter, ":entity_id/", false, 0)
	a.Index(entityNode, 'a')

	assetsNode := entityNode.children[0]
	r.NotNil(assetsNode)
	a.Properties(assetsNode, text, "assets", false, 1)
	a.Leaf(assetsNode)
}

func TestAddSameLengthRoutesWithSharedParameter(t *testing.T) {
	// GIVEN
	a := newAssertions(t)
	r := require.New(t)
	var index tree

	// WHEN
	index.add(method, "/v1/entities/:entity_id/assets", 1)
	index.add(method, "/v1/entities/:entity_id/users", 2)
	rootNode := index.get(method)

	// THEN
	r.NotNil(rootNode)
	a.Properties(rootNode, text, "/v1/entities/", false, 0)
	a.EmptyIndex(rootNode)

	entityNode := rootNode.special
	r.NotNil(entityNode)
	a.Properties(entityNode, parameter, ":entity_id/", false, 0)
	a.Index(entityNode, 'a', 'u')
	r.Len(entityNode.children, 2)

	assetsNode := entityNode.children[0]
	a.Properties(assetsNode, text, "assets", false, 1)

	usersNode := entityNode.children[1]
	a.Properties(usersNode, text, "users", false, 2)
}

func TestAddLongRouteWithSharedParameter(t *testing.T) {
	// GIVEN
	a := newAssertions(t)
	r := require.New(t)
	var index tree

	// WHEN
	index.add(method, "/v1/entities/:entity_id/assets", 1)
	index.add(method, "/v1/entities/:entity_id/payment-methods/:payment_method_id", 1)
	branch := index.get(method)

	// THEN
	a.Properties(branch, text, "/v1/entities/", false, 0)
	a.EmptyIndex(branch)

	entityNode := branch.special
	r.NotNil(entityNode)
	a.Properties(entityNode, parameter, ":entity_id/", false, 0)
	a.Index(entityNode, 'a', 'p')
	r.Len(entityNode.children, 2)

	assetsNode := entityNode.children[0]
	r.NotNil(assetsNode)
	a.Properties(assetsNode, text, "assets", false, 1)
	a.Leaf(assetsNode)

	paymentMethodsNode := entityNode.children[1]
	r.NotNil(paymentMethodsNode)
	a.Properties(paymentMethodsNode, text, "payment-methods/", false, 0)
	a.EmptyIndex(paymentMethodsNode)

	paymentMethodNode := paymentMethodsNode.special
	r.NotNil(paymentMethodsNode)
	a.Properties(paymentMethodNode, parameter, ":payment_method_id", false, 1)
	a.EmptyIndex(paymentMethodNode)
}

func TestAddShortRoutePrecedingParameter(t *testing.T) {
	// GIVEN
	a := newAssertions(t)
	r := require.New(t)
	var index tree

	// WHEN
	index.add(method, "/v1/portfolios/:portfolio_id/users", 1)
	index.add(method, "/v1/portfolios", 2)
	rootNode := index.get(method)

	// THEN
	a.Properties(rootNode, text, "/v1/portfolios", false, 2)
	a.Index(rootNode, '/')
	r.Len(rootNode.children, 1)
	separatorNode := rootNode.children[0]

	portfolioNode := separatorNode.special
	r.NotNil(portfolioNode)
	a.Properties(portfolioNode, parameter, ":portfolio_id/", false, 0)
	a.Index(portfolioNode, 'u')
	r.Len(portfolioNode.children, 1)

	usersNode := portfolioNode.children[0]
	a.Properties(usersNode, text, "users", false, 1)
	a.EmptyIndex(usersNode)
}

func TestAddLongRouteWithSharedStaticText(t *testing.T) {
	// GIVEN
	a := newAssertions(t)
	r := require.New(t)
	var index tree

	// WHEN
	index.add(method, "/v1/portfolios/:portfolio_id/order", 1)
	index.add(method, "/v1/portfolios/:portfolio_id/order/:order_id", 2)
	rootNode := index.get(method)

	// THEN
	a.Properties(rootNode, text, "/v1/portfolios/", false, 0)
	a.EmptyIndex(rootNode)

	portfolioNode := rootNode.special
	r.NotNil(portfolioNode)
	a.Properties(portfolioNode, parameter, ":portfolio_id/", false, 0)
	a.Index(portfolioNode, 'o')
	r.Len(portfolioNode.children, 1)

	orderNode := portfolioNode.children[0]
	r.NotNil(t, orderNode)
	a.Properties(orderNode, text, "order", false, 1)
	a.Index(orderNode, '/')
	r.Len(orderNode.children, 1)

	orderSeparatorNode := orderNode.children[0]
	r.NotNil(orderSeparatorNode)
	a.Properties(orderSeparatorNode, text, "/", false, 0)
	a.EmptyIndex(orderSeparatorNode)

	orderParamNode := orderSeparatorNode.special
	r.NotNil(orderParamNode)
	a.Properties(orderParamNode, parameter, ":order_id", false, 2)
	a.Leaf(orderParamNode)
}

func TestAddLongRouteWithPrefix(t *testing.T) {
	// GIVEN
	a := newAssertions(t)
	r := require.New(t)
	var index tree

	// WHEN
	index.add(method, "/v1/portfolios/:portfolio_id/order", 1)
	index.add(method, "/v1/portfolios/:portfolio_id/order_preview", 1)
	rootNode := index.get(method)

	// THEN
	a.Properties(rootNode, text, "/v1/portfolios/", false, 0)
	a.EmptyIndex(rootNode)

	portfolioNode := rootNode.special
	r.NotNil(portfolioNode)
	a.Properties(portfolioNode, parameter, ":portfolio_id/", false, 0)
	a.Index(portfolioNode, 'o')
	r.Len(portfolioNode.children, 1)

	orderNode := portfolioNode.children[0]
	r.NotNil(t, orderNode)
	a.Properties(orderNode, text, "order", false, 1)
	a.Index(orderNode, '_')
	r.Len(orderNode.children, 1)

	orderPreviewNode := orderNode.children[0]
	r.NotNil(orderPreviewNode)
	a.Properties(orderPreviewNode, text, "_preview", false, 1)
	a.Leaf(orderPreviewNode)
}

func TestAddParameterOnlyRoutesLongerFirst(t *testing.T) {
	// GIVEN
	a := newAssertions(t)
	r := require.New(t)
	var index tree

	// WHEN
	index.add(method, "/:one/:two/:three", 3)
	index.add(method, "/:one/:two", 2)
	rootNode := index.get(method)

	// THEN
	r.NotNil(rootNode)
	a.Properties(rootNode, text, "/", false, 0)
	a.EmptyIndex(rootNode)

	oneNode := rootNode.special
	r.NotNil(oneNode)
	a.Properties(oneNode, parameter, ":one/", false, 0)
	a.EmptyIndex(oneNode)

	twoNode := oneNode.special
	r.NotNil(twoNode)
	a.Properties(twoNode, parameter, ":two/", false, 2)
	a.EmptyIndex(twoNode)

	threeNode := twoNode.special
	r.NotNil(threeNode)
	a.Properties(threeNode, parameter, ":three", false, 3)
	a.Leaf(threeNode)
}

func TestAddParameterOnlyRoutesShorterFirst(t *testing.T) {
	// GIVEN
	a := newAssertions(t)
	r := require.New(t)
	var index tree

	// WHEN
	index.add(method, "/:one/:two", 2)
	index.add(method, "/:one/:two/:three", 3)
	rootNode := index.get(method)

	// THEN
	r.NotNil(rootNode)
	a.Properties(rootNode, text, "/", false, 0)
	a.EmptyIndex(rootNode)

	oneNode := rootNode.special
	r.NotNil(oneNode)
	a.Properties(oneNode, parameter, ":one/", false, 0)
	a.EmptyIndex(oneNode)

	twoNode := oneNode.special
	r.NotNil(twoNode)
	a.Properties(twoNode, parameter, ":two", false, 2)
	a.EmptyIndex(twoNode)

	threeNode := twoNode.special
	r.NotNil(threeNode)
	a.Properties(threeNode, parameter, ":three", false, 3)
	a.Leaf(threeNode)
}

func TestAddPathsWithSegmentDifference(t *testing.T) {
	// GIVEN
	a := newAssertions(t)
	r := require.New(t)
	var index tree

	// WHEN
	index.add(method, "/v1/entities/:entity_id/users", 2)
	index.add(method, "/v1/portfolios/:portfolio_id/users", 3)
	rootNode := index.get(method)

	// THEN
	r.NotNil(rootNode)
	a.Properties(rootNode, text, "/v1/", false, 0)
	a.Index(rootNode, 'e', 'p')
	r.Len(rootNode.children, 2)

	entitiesNode := rootNode.children[0]
	r.NotNil(entitiesNode)
	a.Properties(entitiesNode, text, "entities/", false, 0)
	a.EmptyIndex(entitiesNode)
	r.NotNil(entitiesNode.special)

	entityNode := entitiesNode.special
	r.NotNil(entityNode)
	a.Properties(entityNode, parameter, ":entity_id/", false, 0)
	a.Index(entityNode, 'u')
	r.Len(entityNode.children, 1)

	entityUsersNode := entityNode.children[0]
	a.Properties(entityUsersNode, text, "users", false, 2)
	a.EmptyIndex(entityUsersNode)

	portfoliosNode := rootNode.children[1]
	a.Properties(portfoliosNode, text, "portfolios/", false, 0)
	a.EmptyIndex(portfoliosNode)

	portfolioNode := portfoliosNode.special
	r.NotNil(portfolioNode)
	a.Properties(portfolioNode, parameter, ":portfolio_id/", false, 0)
	a.Index(portfolioNode, 'u')
	r.Len(entityNode.children, 1)

	portfolioUsersNode := portfolioNode.children[0]
	a.Properties(portfolioUsersNode, text, "users", false, 3)
	a.Leaf(portfolioUsersNode)
}

func TestAddTextSegmentAfterWildcard(t *testing.T) {
	// GIVEN
	a := newAssertions(t)
	r := require.New(t)
	var index tree

	// WHEN
	index.add(method, "/v1/*", 2)
	index.add(method, "/v1/portfolios", 3)
	rootNode := index.get(method)

	// THEN
	r.NotNil(rootNode)
	a.Properties(rootNode, text, "/v1/", true, 2)
	a.Index(rootNode, 'p')
	r.Len(rootNode.children, 1)

	portfoliosNode := rootNode.children[0]
	r.NotNil(portfoliosNode)
	a.Properties(portfoliosNode, text, "portfolios", false, 3)
	a.False(portfoliosNode.hasWildcard)
	a.Leaf(portfoliosNode)
}

func TestAddWildcardSegmentAfterText(t *testing.T) {
	// GIVEN
	a := newAssertions(t)
	r := require.New(t)
	var index tree

	// WHEN
	index.add(method, "/v1/portfolios", 3)
	index.add(method, "/v1/*", 2)
	rootNode := index.get(method)

	// THEN
	r.NotNil(rootNode)
	a.Properties(rootNode, text, "/v1/", true, 2)
	a.Index(rootNode, 'p')
	r.Len(rootNode.children, 1)

	portfoliosNode := rootNode.children[0]
	r.NotNil(portfoliosNode)
	a.Properties(portfoliosNode, text, "portfolios", false, 3)
	a.Leaf(portfoliosNode)
}

func TestBreakTextNodeByWildcard(t *testing.T) {
	// GIVEN
	a := newAssertions(t)
	r := require.New(t)
	var index tree

	// WHEN
	index.add(method, "/v1/portfolios", 3)
	index.add(method, "/v1/port*", 2)
	rootNode := index.get(method)

	// THEN
	r.NotNil(rootNode)
	a.Properties(rootNode, text, "/v1/port", true, 2)
	a.Index(rootNode, 'f')
	r.Len(rootNode.children, 1)

	portfoliosNode := rootNode.children[0]
	r.NotNil(portfoliosNode)
	a.Properties(portfoliosNode, text, "folios", false, 3)
	a.Leaf(portfoliosNode)
}

func TestAddWildcardAfterWildcard(t *testing.T) {
	// GIVEN
	a := newAssertions(t)
	r := require.New(t)
	var index tree

	// WHEN
	index.add(method, "/v1/portfolios*", 3)
	index.add(method, "/v1/port*", 2)
	rootNode := index.get(method)

	// THEN
	r.NotNil(rootNode)
	a.Properties(rootNode, text, "/v1/port", false, 2)
	a.Index(rootNode, 'f')
	r.Len(rootNode.children, 1)

	portfoliosNode := rootNode.children[0]
	r.NotNil(portfoliosNode)
	a.Properties(portfoliosNode, text, "folios", false, 3)
	a.True(portfoliosNode.hasWildcard)
	a.Leaf(portfoliosNode)
}

func TestFindCostExactMatch(t *testing.T) {
	// GIVEN
	a := assert.New(t)
	var index tree
	index.add(method, "/v1/entities", 1)
	index.add(method, "/v1/entities/assets", 2)

	// WHEN
	rootNode := index.get(method)
	entitiesCost := rootNode.get("/v1/entities")
	assetsCost := rootNode.get("/v1/entities/assets")

	// THEN
	a.Equal(1, entitiesCost, "/v1/entities")
	a.Equal(2, assetsCost, "/v1/entities/assets")
}

func TestFindCostWithParameter(t *testing.T) {
	// GIVEN
	a := assert.New(t)
	var index tree
	index.add(method, "/v1/entities", 1)
	index.add(method, "/v1/entities/:entity_id", 2)

	// WHEN
	rootNode := index.get(method)
	entitiesCost := rootNode.get("/v1/entities")
	entityCost := rootNode.get("/v1/entities/1")

	// THEN
	a.Equal(1, entitiesCost, "/v1/entities")
	a.Equal(2, entityCost, "/v1/entities/:entity_id")
}

func TestFindCostWithWildcard(t *testing.T) {
	// GIVEN
	a := assert.New(t)
	var index tree
	index.add(method, "/v1/entities/*", 1)
	index.add(method, "/v1/entities/all", 2)

	// WHEN
	rootNode := index.get(method)
	entityCost := rootNode.get("/v1/entities/1")
	allEntityCost := rootNode.get("/v1/entities/all")

	// THEN
	a.Equal(1, entityCost, "/v1/entities/1")
	a.Equal(2, allEntityCost, "/v1/entities/all")
}

func TestFindCostWithoutBacktracking(t *testing.T) {
	// GIVEN
	a := assert.New(t)
	var index tree
	index.add(method, "/v1/entities/*", 1)
	index.add(method, "/v1/entities/:entity_id/payment-methods/fiat", 2)
	index.add(method, "/v1/entities/:entity_id/payment-methods/*", 3)

	// WHEN
	rootNode := index.get(method)
	entityCost := rootNode.get("/v1/entities/1")
	fiatPaymentCost := rootNode.get("/v1/entities/2/payment-methods/fiat")
	wirePaymentCost := rootNode.get("/v1/entities/3/payment-methods/wire")

	// THEN
	a.Equal(1, entityCost, "/v1/entities/1")
	a.Equal(2, fiatPaymentCost, "/v1/entities/2/payment-methods/fiat")
	a.Equal(3, wirePaymentCost, "/v1/entities/3/payment-methods/wire")
}

func TestFindCostWithBacktracking(t *testing.T) {
	// GIVEN
	a := assert.New(t)
	var index tree
	index.add(method, "/v1/entities/self", 1)
	index.add(method, "/:version/entities/:entity_id", 2)
	index.add(method, "/:version/entities/:entity_id/payment-methods/wire", 3)
	index.add(method, "/:version/entities/:entity_id/:operation/dryRun", 4)
	index.add(method, "/:version/:group/:group_id/:operation", 5)

	// WHEN
	rootNode := index.get(method)
	editSelfCost := rootNode.get("/v1/entities/self")
	entityCost := rootNode.get("/v1/entities/1")
	paymentMethodWireCost := rootNode.get("/v1/entities/1/payment-methods/wire")
	cancelDryRunCost := rootNode.get("/v1/entities/1/cancel/dryRun")
	editCost := rootNode.get("/v1/entities/1/edit")
	editTestRunCost := rootNode.get("/v1/entities/1/edit/testRun")

	// THEN
	a.Equal(1, editSelfCost, "/v1/entities/self")
	a.Equal(2, entityCost, "/v1/entities/1")
	a.Equal(3, paymentMethodWireCost, "/v1/entities/1/payment-methods/wire")
	a.Equal(4, cancelDryRunCost, "/v1/entities/1/cancel/dryRun")
	a.Equal(5, editCost, "/v1/entities/1/edit")
	a.Equal(0, editTestRunCost, "/v1/entities/self/edit/testRun")
}

func TestFindCostTrailingSlash(t *testing.T) {
	// GIVEN
	a := assert.New(t)
	var index tree
	index.add(method, "/v1/entities", 1)

	// WHEN
	rootNode := index.get(method)
	entitiesCost := rootNode.get("/v1/entities/")

	// THEN
	a.Equal(1, entitiesCost, "/v1/entities/")
}

func TestBinanceInsertTextThroughSlash(t *testing.T) {
	// GIVEN
	a := newAssertions(t)
	r := require.New(t)
	var index tree

	// WHEN
	index.add(method, "/sapi/v1/margin/isolated/pair", 1)
	index.add(method, "/sapi/v1/margin/isolated/transfer", 1)
	index.add(method, "/sapi/v1/margin/isolatedMarginData", 1)

	// THEN
	rootNode := index.get(method)
	r.NotNil(rootNode)
	a.Equal([]byte{'/', 'M'}, rootNode.index)
	r.Len(rootNode.children, 2)

	separatorNode := rootNode.children[0]
	r.NotNil(separatorNode)
	a.Index(separatorNode, 'p', 't')
	r.Len(separatorNode.children, 2)

	pairNode := separatorNode.children[0]
	r.NotNil(pairNode)
	a.Properties(pairNode, text, "pair", false, 1)

	transferNode := separatorNode.children[1]
	r.NotNil(transferNode)
	a.Properties(transferNode, text, "transfer", false, 1)

	marginDataNode := rootNode.children[1]
	r.NotNil(marginDataNode)
	a.Properties(marginDataNode, text, "MarginData", false, 1)
}
