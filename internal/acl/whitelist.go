/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package acl

import (
	"net"
	"strings"
)

type Whitelist struct {
	ip   map[string]bool
	mask []net.IPNet
}

type WhiteListOptions struct {
	ip   map[string]bool
	mask []net.IPNet
}

type WhitelistOption func(options *WhiteListOptions)

func NewWhitelist(opts ...WhitelistOption) *Whitelist {
	options := WhiteListOptions{
		ip: make(map[string]bool),
	}
	for _, opt := range opts {
		opt(&options)
	}
	return &Whitelist{
		ip:   options.ip,
		mask: options.mask,
	}
}

func NewOption(entry string) (WhitelistOption, error) {
	isIp := strings.IndexByte(entry, '/') < 0
	if isIp {
		ip := net.ParseIP(entry)
		if ip == nil {
			return nil, &net.ParseError{Type: "IP address", Text: entry}
		}
		return WithIP(ip), nil
	}

	_, mask, err := net.ParseCIDR(entry)
	if err != nil {
		return nil, err
	}
	return WithMask(*mask), nil
}

func WithIP(ip net.IP) WhitelistOption {
	return func(opts *WhiteListOptions) {
		opts.ip[ip.String()] = true
	}
}

func WithMask(net net.IPNet) WhitelistOption {
	return func(opts *WhiteListOptions) {
		opts.mask = append(opts.mask, net)
	}
}

func (w *Whitelist) Allowed(ip net.IP) bool {
	_, ok := w.ip[ip.String()]
	if ok {
		return true
	}

	for _, mask := range w.mask {
		if mask.Contains(ip) {
			return true
		}
	}

	return false
}
