# Developer Guide

## Requirements

Install the following tools to build the project:

- Docker
- go runtime, minimum version 1.21
- make

## Walkthrough 

The Glove framework consists of two public packages:

`cmd` - an auxiliary package that provides a universal and reusable command line interface to simplify the bootstrapping of projects implemented using the framework.

`proxy` - a core package that implements the framework.

Besides Golang packages, we provide a standalone command line application (CLI) that demonstrates how to build applications using the framework. The walkthrough begins with the CLI. We explain how to start the forward proxy from the terminal, load a CA private key and certificate, and add an IP address or a CIDR mask to the whitelist of clients that are allowed to connect to the proxy. The second part of the walkthrough describes the framework API and explains how to implement custom handlers for processing HTTP/HTTPS traffic.

### CLI

Glove's CLI is built using the [`cmd`](../pkg/cmd) package that aims to streamline the bootstrapping of forward proxies implemented using Glove. The package provides a command line interface to run a proxy on a specific network endpoint and port, load a CA private key and certificate, set up an IP or CIDR-based whitelist of allowed client hosts, configure logging level, or display a help message.

The remainder of the section explains how to build CLI and run it as a standalone application.

1. Build the CLI executable

   ```shell
   make cli
   ```

Find the `glove` executable in the `./bin` directory.

2. Run the forward proxy locally using port `8080`

   ```shell
   glove listen --port=8080
   ```

The proxy will bind to the loopback interface and listen on port 8080. The proxy will accept any incoming TCP connection and perform tunneling.

Use `--host=0.0.0.0` if you want to bind on all available network interfaces. If your aim is to run the proxy in a public subnet, you should consider whitelisting IP addresses to restrict clients who can use the proxy.

Use the `--whitelist` option to add an IP address or a CIDR mask to the whitelist. Suppose the whitelist is enabled. If a connection is requested by a client with an IP address outside the list, it will be rejected.

``shell
listen --host=0.0.0.0 --port=8080 --whitelist=127.0.0.1
``

3. Handle incoming connections in the MIM mode and establish the TLS handshake using a certificate signed by a custom CA.

Use `--defaultAction=mitm` to handle connections using MITM.

With no additional settings, the proxy will generate a private key and a self-signed certificate for the CA. The proxy will use the CA to sign a certificate generated for a TLS handshake with the client. If the client verifies the signing authority, which is the default behavior, the TLS handshake won't be accepted.

Use the `--caPrivateKey` and `--caCert` options to specify a location in the file system of the private key and certificate that the proxy should use for the CA. The files should be saved in the `PEM` format. If you are looking for a private key and a trusted certificate suitable for local development, check the [`mkcert`](https://github.com/FiloSottile/mkcert) utility.

Suppose you generated the private key and the certificate using `mkcert` on Linux. The following command runs the proxy with a custom CA private key and certificate.

```shell
 glove listen --host=0.0.0.0 --port=8080 --whitelist=127.0.0.1 --caCertFile=~/.local/share/mkcert/rootCA.pem --caKeyFile=~/.local/share/mkcert/rootCA-key.pem
  ```

The example above concludes the tour of the CLI.

### API

The section explains set up the Glove framework to run arbitrary Golang code for processing HTTP/HTTPS requests and responses while they are passing through the proxy.

The Glove framework requires that Golang code for processing HTTP/HTTPS requests and reponses is provided as functions compatible with the `Handler` type. Henceforth, we will refer to an arbitrary function that satisfies the Handler type, as a handler for brevity.

```go
package proxy

type Handler func(c *Context)
```

A handler accepts a single argument of the `Context` type and returns no result. The `Context` type is a structure that the framework uses to track the current state of the request's processing. Besides internal fields whose discussion we skip as an implementation detail, the structure contains pointers to the HTTP/HTTPS request and response stored using types from the Golang standard library.

```go
package proxy

type Context struct {
    Request  *http.Request
    Response *http.Response
    // private members are skipped for brevity
}

func (c *Context) Next() {
    // Next moves request processing to the next handler
    // implementation is skipped for brevity
}
```

The `Request` field points to the HTTP/HTTPS request received by the proxy from the client. The `Response` points to the HTTP\HTTPS response sent to the proxy by the origin server. The pointer to the response is initially set to nil until the server's response is received by the proxy.

The `Context` type implements the `Next` method which should be called within a handler function to indicate that an HTTP request has been processed and can be passed to the subsequent handler or sent to the origin server.

Let us explain how to implement a Glove handler using the following example.

```go
package nop

func NopHandler(c *proxy.Context) {
    // code for processing c.Request
    c.Next()
    // code for processing c.Response
}
```

The handlers' code contains two sections separated by the `c.Next()` call. The first section is dedicated to processing an HTTP request. The handler can access and modify any field of the `http.Request`. Once the handler completes processing the request, it should call the `Next` method. Not doing it results in aborting the request. In such circumstances, the proxy responds to the client with an HTTP 500 Internal Server Error. If the handler wants to send a different response, it should set the pointer `c.Response` to an instance of the `http.Response` type.

Suppose `c.Next()` was called by the last handler in the pipeline. Then, the final request with all modifications made by handlers is sent to the origin server.

Once the server replies, the response is processed by handlers in the pipeline in the reverse order. Similarly to the request, handlers can modify any field of `http.Response`. A handler could also implement error handling logic and instead of accepting the response retry the request by calling `c.Next()` again. Finally, once the first handler completes processing, the proxy sends the response to the client.

So far we have shown how to wrap Golang code in handlers so the Glove framework can execute it. Furthermore, the framework needs to be told in which order to execute handlers and whether they should be executed for all hosts or only for some specific hosts.

The framework behavior is controlled using a set of rules. A single rule is defined using the struct type presented below.

```go
package proxy

type Rule struct {
    Action       Action
    Handlers     []Handler
    ClientConfig func(host string) (*tls.Config, error)
    ServerConfig func(host string) (*tls.Config, error)
}
```

The `Action` field defines the connection handling strategy. It could assume one of the following values: `TunnelAction` (default), `BlockAction`, and `MITMAction`. `TunnelAction` instructs the framework to perform TCP tunneling. As a result, the TLS handshake is made between the client and the origin server. The proxy cannot intercept the HTTPS traffic, and all other parameters of the rule are ignored. `BlockAction` instructs the framework to immediately respond to the client with an HTTP 403 Forbidden Error instead of sending the request to the origin server. Finally, `MITMAction` configures the framework to
establish the TLS handshake with the client. The proxy can intercept HTTP/HTTPS traffic and execute handlers passed
using the `Handlers` slice. Functions `ClientConfig` and `ServerConfig` are optional. They are available in the API to allow for custom TLS configuration determined by the origin server. If `ClientConfig` is set to nil, the proxy will generate a private key and a certificate for the TLS handshake with the client and sign it using the CA private key and certificate the proxy was configured to use. If `ServerConfig` is set to nil, the proxy will use a default TLS configuration to connect to the origin server.

Every remote host can have one rule that governs how the framework should handle HTTP/HTTPS traffic. If no rule is defined for the given host, the proxy falls back to the default rule. Only one rule can be nominated as the default.

The example below shows how to set a default rule for the proxy.

```go
package main

import (
    "glove/examples/nop"
    "glove/internal/cmd"
    "glove/pkg/proxy"
)

func main() {
    cmd.Configure(
       proxy.WithDefaultRule(&proxy.Rule{
          Action: proxy.MITMAction,
          Handlers: []proxy.Handler{
             nop.Handle,
          },
       }),
    )
    cmd.Execute()
}
```

We use the `cmd` package to bootstrap the proxy. The `proxy.WithDefaultRule` option defines the default rule. The option is passed to the framework through the `cmd.Configure` function. Finally, the proxy is started using the `cmd.Execute` function.

Conversely, if you want to set up a role for a specific host only, use the `proxy.WithRule` option. Besides the rule definition,
the option requires providing at least one hostname for which the rule should be triggered.

```go

package main

import (
    "glove/examples/nop"
    "glove/internal/cmd"
    "glove/pkg/proxy"
)

func main() {
    cmd.Configure(
       proxy.WithRule(&proxy.Rule{
          Action: proxy.MITMAction,
          Handlers: []proxy.Handler{
             nop.Handle,
          }},
          "api.google.com"),
    )
    cmd.Execute()
}
```

The example above concludes the walkthrough. We began by explaining how to implement a handler for processing HTTP/HTTPS requests and responses. Then, we introduced rules that control the order in which handlers are executed, and whether they are run for a specific group of hosts or all hosts. Finally, we bootstrapped the proxy application using the `cmd` package.

You should now be ready to implement your handlers and bootstrap the proxy application. We hope the tutorial provides the right balance between information indispensable to kick off your project and advanced material you don't need for now. Please let us know if you have suggestions on how to improve the document.
