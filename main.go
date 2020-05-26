package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// DefaultPort is the default port to use if once is not specified by the SERVER_PORT environment variable
const HTTPPort = "8888"
const HTTPSPort = "8443"

func getServerPorts() (ports [2]string) {
	ports[0] = os.Getenv("HTTP_PORT")
	ports[1] = os.Getenv("HTTPS_PORT")
	if ports[0] == "" {
		ports[0] = HTTPPort
	}
	if ports[1] == "" {
		ports[1] = HTTPSPort
	}

	return ports
}

// EchoHandler echos back the request as a response
func EchoHandler(writer http.ResponseWriter, request *http.Request) {

	log.Println("Echoing back request made to " + request.URL.Path + " to client (" + request.RemoteAddr + ")")

	name, err := os.Hostname()
	if err != nil {
		log.Println(err)
	}
	attr := make(map[string]interface{})
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
	request.Write(&buf)
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

func main() {

	http.HandleFunc("/", EchoHandler)
	ports := getServerPorts()
	log.Printf("starting echo server, listening on ports HTTP:%s/HTTPS:%s", ports[0], ports[1])
	// HTTP
	go func() {
		err := http.ListenAndServe(":"+ports[0], nil)
		if err != nil {
			log.Fatal("Echo server (HTTP): ", err)
		}
	}()
	//HTTPS
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
	err = http.ListenAndServeTLS(":"+ports[1], "server.crt", "server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}