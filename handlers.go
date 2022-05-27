package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func (c context) echoHostname(writer http.ResponseWriter, request *http.Request) {
	log.Printf("%s: %s %s", request.RemoteAddr, request.Method, request.URL)
	writer.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(writer, c.hostname)
}

//  echoAll echos back request in response
func (c context) echoAll(writer http.ResponseWriter, request *http.Request) {
	log.Printf("%s: %s %s", request.RemoteAddr, request.Method, request.URL)
	attr := make(map[string]interface{})

	// OS
	attr["os"] = map[string]string{
		"hostname": c.hostname,
	}
	// TCP
	parts := strings.Split(request.RemoteAddr, ":")
	attr["tcp"] = map[string]string{
		"ip":   strings.Join(parts[:(len(parts)-1)], ":"),
		"port": parts[len(parts)-1],
	}
	// TLS
	if request.TLS != nil {
		attr["tls"] = map[string]string{
			"sni":    request.TLS.ServerName,
			"cipher": tls.CipherSuiteName(request.TLS.CipherSuite),
		}
	}
	// HTTP
	headers := make(map[string]string)
	var cookies []string
	var buf bytes.Buffer
	if err := request.Write(&buf); err != nil {
		log.Printf("Error reading request: %s", err)
		return
	}
	if c.verbose {
		log.Printf("Content:\n%s", buf.String())
	}
	for name, value := range request.Header {
		headers[name] = strings.Join(value, " ")
	}
	for _, cookie := range request.Cookies() {
		cookies = append(cookies, cookie.String())
	}
	attr["http"] = map[string]interface{}{
		"protocol": request.Proto,
		"headers":  headers,
		"cookies":  cookies,
		"host":     request.Host,
		"method":   request.Method,
		"path":     request.URL.Path,
		"query":    request.URL.RawQuery,
		"raw":      buf.String(),
	}
	res, _ := json.MarshalIndent(attr, "", "  ")
	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(writer, string(res))
}
