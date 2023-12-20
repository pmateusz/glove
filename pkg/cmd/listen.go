/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package cmd

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/pmateusz/glove/internal/acl"
	"github.com/pmateusz/glove/internal/ca"
	"github.com/pmateusz/glove/internal/cancel"
	"github.com/pmateusz/glove/internal/logging"
	"github.com/pmateusz/glove/pkg/proxy"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"net"
	"net/http"
	"os"
	"time"
)

var host string
var port int
var caCertFilePath string
var caPrivateKeyFilePath string
var whitelistEntries []string
var defaultAction string
var whitelistOptions []acl.WhitelistOption
var engineOptions []proxy.EngineOption

func newListedCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "listen",
		Short: "Start the proxy server",
		Args:  parseListenArgs,
		Run:   runListen,
	}
	flags := command.Flags()
	flags.StringVar(&host, "host", "127.0.0.1", "bind socket to the host")
	flags.IntVar(&port, "port", 8080, "bind socket to the port")
	flags.StringArrayVar(&whitelistEntries, "whitelist", nil, "add an IP address or CIDR mask to the whitelist of allowed clients")
	flags.StringVar(&caCertFilePath, "caCert", "", "path to the CA certificate in the PEM format")
	flags.StringVar(&caPrivateKeyFilePath, "caPrivateKey", "", "path to the CA private key in the PEM format")
	flags.StringVar(&defaultAction, "defaultAction", "tunnel", "set the default strategy for handling connections to any host [block, tunnel, mitm]")

	command.MarkFlagsRequiredTogether("caCert", "caPrivateKey")
	if err := command.MarkFlagFilename("caCert", "pem", "cert", "cer", "crt"); err != nil {
		panic(err)
	}
	if err := command.MarkFlagFilename("caPrivateKey", "pem", "key"); err != nil {
		panic(err)
	}

	return command
}

func parseListenArgs(_ *cobra.Command, _ []string) error {
	logConfig, logConfigErr := newLoggingSettings(logMode, logLevel)
	if logConfigErr != nil {
		return logConfigErr
	}
	logging.SetupGlobal(logConfig.mode, logConfig.level)

	var whitelistErr error
	whitelistOptions, whitelistErr = parseWhitelistEntries(whitelistEntries)
	if whitelistErr != nil {
		return whitelistErr
	}

	var localOptions []proxy.EngineOption
	if caCertFilePath != "" && caPrivateKeyFilePath != "" {
		serverConfigOpt, serverConfigErr := parseServerConfig(caCertFilePath, caPrivateKeyFilePath)
		if serverConfigErr != nil {
			return serverConfigErr
		}

		localOptions = append(localOptions, serverConfigOpt)
	}

	if defaultAction != "" {
		defaultRuleOpt, defaultRuleErr := parseDefaultRule(defaultAction)
		if defaultRuleErr != nil {
			return defaultRuleErr
		}

		localOptions = append(localOptions, defaultRuleOpt)
	}

	if len(localOptions) > 0 {
		// put options passed in the commandline first, so they can be overriden
		engineOptions = append(localOptions, engineOptions...)
	}

	return nil
}

func parseWhitelistEntries(entries []string) ([]acl.WhitelistOption, error) {
	if len(entries) == 0 {
		return nil, nil
	}

	var options []acl.WhitelistOption
	for _, entry := range entries {
		option, optionErr := acl.NewOption(entry)
		if optionErr != nil {
			var parseErr *net.ParseError
			if errors.As(optionErr, &parseErr) {
				return nil, fmt.Errorf("failed to parse the whitelist entry %q as %s", parseErr.Text, parseErr.Type)
			}
			return nil, fmt.Errorf("failed to parse the whitelist entry %q", entry)
		}
		options = append(options, option)
	}

	return options, nil
}

func parseServerConfig(caCertFilePath, caPrivateKeyFilePath string) (proxy.EngineOption, error) {
	proxyCA, err := ca.LoadCA(caCertFilePath, caPrivateKeyFilePath, nil)

	if err != nil {
		return nil, err
	}

	return proxy.WithServerConfig(func(host string) (*tls.Config, error) {
		cert, signErr := proxyCA.SignHosts(host)
		if signErr != nil {
			return nil, signErr
		}
		return &tls.Config{
			Certificates: []tls.Certificate{*cert},
		}, nil
	}), nil
}

func parseDefaultRule(actionName string) (proxy.EngineOption, error) {
	action := proxy.TunnelAction
	if actionName != "" {
		var parseErr error
		action, parseErr = proxy.ParseAction(actionName)
		if parseErr != nil {
			return nil, parseErr
		}
	}
	return proxy.WithDefaultRule(&proxy.Rule{Action: action}), nil
}

func runListen(_ *cobra.Command, _ []string) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	var listenConfig net.ListenConfig
	address := fmt.Sprintf("%s:%d", host, port)
	tcpListener, tcpListenErr := listenConfig.Listen(ctx, "tcp", address)
	if tcpListenErr != nil {
		log.Error().Err(tcpListenErr).Msg("listen")
		return
	}

	var listener net.Listener
	if len(whitelistOptions) > 0 {
		whitelist := acl.NewWhitelist(whitelistOptions...)
		listener = acl.WrapListener(log.Logger, whitelist, tcpListener)
	} else {
		listener = tcpListener
	}

	engine := proxy.NewEngine(engineOptions...)
	var server http.Server
	server.Handler = engine
	hook := cancel.NewHook(ctx, log.Logger)
	hook.Register("server", cancel.WrapServer(&server, 5*time.Second))
	hook.Start()

	log.Info().
		Int("pid", os.Getpid()).
		Int("port", port).
		Str("host", host).
		Msg("listen")

	if err := server.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
		log.Error().Err(err).Msg("server")
	}
}
