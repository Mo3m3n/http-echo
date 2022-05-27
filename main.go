package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
)

// DefaultPort is the default port to use if once is not specified by the SERVER_PORT environment variable
const HTTPPort = 8888
const HTTPSPort = 8443

type context struct {
	httpPort  int
	httpsPort int
	hostname  string
}

func (c *context) getParams() {
	httpPtr := flag.Int("http", HTTPPort, "http port value")
	httpsPtr := flag.Int("https", HTTPSPort, "https port value")
	flag.Parse()
	c.httpPort = *httpPtr
	c.httpsPort = *httpsPtr
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
	var err error
	ctx := context{}
	ctx.hostname, err = os.Hostname()
	if err != nil {
		log.Println(err)
	}
	ctx.getParams()
	http.HandleFunc("/hostname", ctx.echoHostname)
	http.HandleFunc("/", ctx.echoAll)
	log.Printf("starting echo server, listening on ports HTTP:%d/HTTPS:%d", ctx.httpPort, ctx.httpsPort)
	// HTTPS
	go func() {
		listenAndServceTLS(fmt.Sprintf("%d", ctx.httpsPort))
	}()
	// HTTP
	err = http.ListenAndServe(fmt.Sprintf(":%d", ctx.httpPort), nil)
	if err != nil {
		log.Fatal("Echo server (HTTP): ", err)
	}
}
