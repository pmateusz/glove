/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package tree

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type root struct {
	method string
	node   *node
}

type tree struct {
	roots []root
}

type nodeKind uint8

const (
	text nodeKind = iota
	parameter
)

type node struct {
	kind        nodeKind
	hasWildcard bool
	route       string
	index       []byte
	children    []*node
	special     *node
	cost        int
}

func (t *tree) dump() []Route {
	results := make(map[string]map[string]int)
	for _, r := range t.roots {
		results[r.method] = make(map[string]int)
		dumpNode(r.node, r.method, "", results)
	}

	routes := make([]Route, 0, len(results))
	for method, costByRoute := range results {
		for route, cost := range costByRoute {
			routes = append(routes, Route{method, route, cost})
		}
	}
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].method < routes[j].method {
			return true
		}
		return routes[i].method == routes[j].method && routes[i].route < routes[j].route
	})
	return routes
}

func dumpNode(node *node, method string, prefix string, results map[string]map[string]int) {
	fullRoute := prefix + node.route
	if node.cost > 0 {
		results[method][stripTrailingSlash(fullRoute)] = node.cost
	}

	if node.kind == parameter {
		fullRoute = addTrailingSlash(fullRoute)
	}

	if node.special != nil {
		dumpNode(node.special, method, fullRoute, results)
	}

	for pos := range node.children {
		dumpNode(node.children[pos], method, fullRoute, results)
	}
}

func stripTrailingSlash(r string) string {
	if lastPos := len(r) - 1; lastPos != -1 && r[lastPos] == '/' {
		return r[:lastPos]
	}
	return r
}

func addTrailingSlash(r string) string {
	if lastPos := len(r) - 1; lastPos != -1 && r[lastPos] != '/' {
		return r + "/"
	}
	return r
}

func takeText(r string) string {
	for idx := range r {
		if r[idx] == ':' || r[idx] == '*' {
			return r[:idx]
		}
	}
	return r
}

func takeParameter(r string) string {
	if pos := strings.IndexByte(r, '/'); pos != -1 {
		return r[:pos+1]
	}
	return r
}

func skipParameter(r string, offset int) int {
	for pos := offset; pos < len(r); pos++ {
		if r[pos] == '/' {
			return pos + 1
		}
	}
	return len(r)
}

func takePrefix(a, b string) string {
	maxLen := min(len(a), len(b))
	for pos := 0; pos < maxLen; pos++ {
		if a[pos] != b[pos] {
			return a[:pos]
		}
	}

	if len(a) < len(b) {
		return a
	}
	return b
}

func newNode(route string) *node {
	if route == "" {
		return nil
	}

	if route[0] == ':' {
		return &node{
			kind:  parameter,
			route: takeParameter(route),
		}
	}

	return &node{
		kind:        text,
		hasWildcard: route[len(route)-1] == '*',
		route:       takeText(route),
	}
}

func newList(route string, cost int) *node {
	remainder := route
	head := newNode(remainder)
	if head == nil {
		return nil
	}

	current := head
	for remainder = remainder[len(current.route):]; remainder != "" && remainder != "*"; remainder = remainder[len(current.route):] {
		next := newNode(remainder)
		current.insertNode(next)
		current = next
	}
	current.cost = cost

	return head
}

func (t *tree) get(method string) *node {
	for _, branch := range t.roots {
		if branch.method == method {
			return branch.node
		}
	}
	return nil
}

func (n *node) insertNode(child *node) {
	if child.kind == text {
		n.index = append(n.index, child.route[0])
		n.children = append(n.children, child)
	} else {
		n.special = child
	}
}

func (n *node) insertRoute(route string, cost int) {
	remainder := route
	current := n

walk:
	for remainder != "" {
		if remainder[0] == ':' {
			// remaining route starts with a parameter
			if current.special == nil {
				// insertNode remaining route starting as a special routeNode
				break
			}

			current = current.special
			param := takeParameter(remainder)
			remainder = remainder[len(param):]
			continue walk
		}

		if current.kind == text {
			// neither remainder nor current's route start with a parameter
			prefix := takePrefix(remainder, current.route)

			// handle trailing prefix
			// don't split nodes only if difference is due to trailing / or *
			if len(prefix) < len(current.route) {
				// current routeNode strictly includes prefix
				// split current routeNode into two nodes: prefix and stem
				stem := &node{
					kind:        text,
					route:       current.route[len(prefix):],
					index:       current.index,
					special:     current.special,
					children:    current.children,
					hasWildcard: current.hasWildcard,
					cost:        current.cost,
				}

				current.index = []byte{stem.route[0]}
				current.children = []*node{stem}
				current.route = prefix
				current.cost = 0
				current.special = nil
			}

			// current routeNode's route is equal to prefix
			remainder = remainder[len(prefix):]
			if remainder == "" {
				break
			}

			if remainder[0] == ':' {
				continue
			}
		}

		c := remainder[0]
		for idx, key := range current.index {
			if key == c {
				current = current.children[idx]
				continue walk
			}
		}

		break
	}

	// current routeNode is the last available routeNode indexing the route
	if remainder == "" {
		// route is fully included in the index
		current.cost = cost
		return
	}

	if remainder == "*" {
		current.cost = cost
		current.hasWildcard = true
		return
	}

	// route is partially included in the index
	stem := newList(remainder, cost)
	current.insertNode(stem)
}

type checkpoint struct {
	node   *node
	offset int
	cost   int
}

type Route struct {
	method string
	route  string
	cost   int
}

func (n *node) get(route string) int {
	offset := 0
	current := n
	cost := 0

	var checkpoints []checkpoint
walk:
	for current != nil {
		if current.hasWildcard {
			cost = current.cost
		}

		remLen := len(route) - offset

		if current.kind == text {
			currentLen := len(current.route)
			maxLen := min(currentLen, remLen)
			prefixLen := 0
			for prefixLen < maxLen && current.route[prefixLen] == route[offset+prefixLen] {
				prefixLen++
			}

			offset += prefixLen
			if prefixLen == maxLen {
				// found match with a text node
				if prefixLen == remLen || (prefixLen == remLen-1 && route[offset] == '/') {
					// final found full match
					return current.cost
				}
			}
		} else {
			offset = skipParameter(route, offset)
			if offset == len(route) {
				if current.cost > 0 {
					return current.cost
				}

				lastCheckpointPos := len(checkpoints) - 1
				if lastCheckpointPos != -1 {
					lastCheckpoint := checkpoints[lastCheckpointPos]
					checkpoints = checkpoints[:lastCheckpointPos]
					current = lastCheckpoint.node
					offset = lastCheckpoint.offset
					cost = lastCheckpoint.cost
					continue walk
				}

				return cost
			}
		}

		// find next node
		c := route[offset]
		for pos, key := range current.index {
			if key == c {
				if current.special != nil {
					checkpoints = append(checkpoints, checkpoint{current.special, offset, cost})
				}
				current = current.children[pos]
				continue walk
			}
		}

		// no text node found, check if we can match parameter
		if current.special != nil {
			current = current.special
			continue walk
		}

		if current.hasWildcard {
			return current.cost
		}

		lastCheckpointPos := len(checkpoints) - 1
		if lastCheckpointPos != -1 {
			lastCheckpoint := checkpoints[lastCheckpointPos]
			checkpoints = checkpoints[:lastCheckpointPos]
			current = lastCheckpoint.node
			offset = lastCheckpoint.offset
			cost = lastCheckpoint.cost
			continue walk
		}

		break
	}

	return cost
}

var parameterPattern = regexp.MustCompile(`\:\w*`)

func validateRoute(text string) (string, error) {
	if text == "" || text[0] != '/' {
		return "", errors.New("route must start with '/': " + text)
	}

	lastPos := len(text) - 1
	if wildcardPos := strings.IndexByte(text, '*'); wildcardPos >= 0 && wildcardPos < lastPos {
		return "", errors.New("wildcard '*' is only allowed at the end of the route: " + text)
	}

	locations := parameterPattern.FindAllStringIndex(text, len(text))
	for idx := range locations {
		start, end := locations[idx][0], locations[idx][1]
		if text[start-1] != '/' {
			paramName := text[start : end+1]
			return "", fmt.Errorf("parameter '%s' must directly follow '/' in the route: %s", paramName, text)
		}

		if end-start < 2 {
			return "", errors.New("at least one parameter is not named in the route: " + text)
		}
	}

	if text[lastPos] == '/' {
		return text[:lastPos], nil
	}

	return text, nil
}

func (t *tree) add(method string, route string, cost int) {
	if rootNode := t.get(method); rootNode != nil {
		rootNode.insertRoute(route, cost)
		return
	}

	head := newList(route, cost)
	t.roots = append(t.roots, root{method: method, node: head})
}
