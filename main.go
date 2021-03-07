package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
)

const HTTPPort = 8888
const HTTPSPort = 8443

func getParams() map[string]string {
	params := make(map[string]string)
	httpPtr := flag.Int("http", HTTPPort, "http port value")
	httpsPtr := flag.Int("https", HTTPSPort, "https port value")
	tlsAuthPtr := flag.Bool("tlsAuth", false, "bool value if set to true, client certificate is required")
	flag.Parse()
	params["http"] = strconv.Itoa(*httpPtr)
	params["https"] = strconv.Itoa(*httpsPtr)
	if *tlsAuthPtr {
		params["tlsAuth"] = ""
	}
	return params
}

// EchoHandler prints to stdout the Body of a POST request
func EchoHandler(writer http.ResponseWriter, request *http.Request) {
	buf := make([]byte, 100)
	for {
		n, err := request.Body.Read(buf)
		fmt.Println(string(buf[:n]))
		if err == io.EOF {
			break
		}
	}
}

func main() {
	params := getParams()
	http.HandleFunc("/", EchoHandler)
	log.Printf("starting echo server, listening on ports HTTP:%s/HTTPS:%s", params["http"], params["https"])
	// HTTP
	err := http.ListenAndServe(":"+params["http"], nil)
	if err != nil {
		log.Fatal("Echo server (HTTP): ", err)
	}
	//HTTPS
	go func() {
		listenAndServeHTTPS(params)
	}()
}

func listenAndServeHTTPS(params map[string]string) {
	if https := params["https"]; https == "0" {
		return
	}
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
	https := &http.Server{
		Addr: ":" + params["https"],
	}
	if _, ok := params["tlsAuth"]; ok {
		https.TLSConfig = &tls.Config{
			//ClientCAs: caCertPool,
			ClientAuth: tls.RequireAnyClientCert,
		}
	}
	err = https.ListenAndServeTLS("server.crt", "server.key")
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
