package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// DefaultPort is the default port to use if once is not specified by the SERVER_PORT environment variable
const HTTPPort = 8888
const HTTPSPort = 8443

func getParams() map[string]string {
	params := make(map[string]string)
	httpPtr := flag.Int("http", HTTPPort, "http port value")
	httpsPtr := flag.Int("https", HTTPSPort, "https port value")
	flag.Parse()
	params["http"] = strconv.Itoa(*httpPtr)
	params["https"] = strconv.Itoa(*httpsPtr)
	return params
}

// EchoHandler echos back the request as a response
func EchoHandler(writer http.ResponseWriter, request *http.Request) {
	log.Println("Echoing back request made to " + request.URL.Path + " to client (" + request.RemoteAddr + ")")
	var attr map[string]interface{}

	name, err := os.Hostname()
	if err != nil {
		log.Println(err)
	}

	if request.URL.Path == "/hostname" {
		writer.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(writer, name)
		return
	}

	attr = make(map[string]interface{})
	attr["os"] = map[string]string{
		"hostname": name,
	}
	log.Println(request.RemoteAddr)
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
	if err = request.Write(&buf); err != nil {
		log.Printf("Error: %s", err)
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

func listenAndServceTLS(port string) {
	cmd := exec.Command("./generate-cert.sh")
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	_, err = os.Stat("server.crt")
	if os.IsNotExist(err) {
		log.Fatal("server.crt: ", err)
	}
	_, err = os.Stat("server.key")
	if os.IsNotExist(err) {
		log.Fatal("server.key: ", err)
	}
	err = http.ListenAndServeTLS(":"+port, "server.crt", "server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func main() {

	params := getParams()
	http.HandleFunc("/", EchoHandler)
	log.Printf("starting echo server, listening on ports HTTP:%s/HTTPS:%s", params["http"], params["https"])
	// HTTPS
	go func() {
		listenAndServceTLS(params["https"])
	}()
	// HTTP
	err := http.ListenAndServe(":"+params["http"], nil)
	if err != nil {
		log.Fatal("Echo server (HTTP): ", err)
	}
}
