package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

type Proxy struct {
	Transport  http.RoundTripper
	Credential string
}

func NewProxy() *Proxy {
	return &Proxy{}
}

func (p *Proxy) handleTunnel(rw http.ResponseWriter, req *http.Request) {
	host := req.URL.Host

	hij, ok := rw.(http.Hijacker)
	if !ok {
		panic("HTTP Server does not support hijacking")
	}

	client, _, err := hij.Hijack()
	if err != nil {
		return
	}
	client.Write([]byte("HTTP/1.0 200 Connection Established\r\n\r\n"))

	server, err := net.Dial("tcp", host)
	if err != nil {
		return
	}

	go io.Copy(server, client)
	io.Copy(client, server)
}

func (p *Proxy) proxyAuthCheck(r *http.Request) (ok bool) {
	if p.Credential == "" { // no auth
		return true
	}
	auth := r.Header.Get("Proxy-Authorization")
	if auth == "" {
		return
	}
	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return
	}
	credential := auth[len(prefix):]
	return credential == p.Credential
}

func (p *Proxy) handleProxyAuth(w http.ResponseWriter, r *http.Request) bool {
	if p.proxyAuthCheck(r) {
		return true
	}
	w.Header().Add("Proxy-Authenticate", "Basic realm=\"*\"")
	w.WriteHeader(http.StatusProxyAuthRequired)
	w.Write(nil)
	return false
}

// ServeHTTP is the main handler for all requests.
func (p *Proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if !p.handleProxyAuth(rw, req) {
		return
	}
	fmt.Printf("Received request %s %s %s\n",
		req.Method,
		req.Host,
		req.RemoteAddr,
	)

	if req.Method == "CONNECT" {
		p.handleTunnel(rw, req)
		return
	}

	transport := p.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	outReq := new(http.Request)
	*outReq = *req

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		if prior, ok := outReq.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		outReq.Header.Set("X-Forwarded-For", clientIP)
	}

	res, err := transport.RoundTrip(outReq)
	if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		return
	}

	for key, value := range res.Header {
		for _, v := range value {
			rw.Header().Add(key, v)
		}
	}

	rw.WriteHeader(res.StatusCode)
	io.Copy(rw, res.Body)
	res.Body.Close()
}

func main() {
	addr := flag.String("addr", ":8780", "listen address")
	auth := flag.String("auth", "", "http auth, eg: admin:admin")
	flag.Parse()

	proxy := NewProxy()
	if *auth != "" {
		proxy.Credential = base64.StdEncoding.EncodeToString([]byte(*auth))
	}
	fmt.Printf("listening on %s\n", *addr)
	http.ListenAndServe(*addr, proxy)
}
