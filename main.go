package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
)

// DefaultPort is the default port to use if once is not specified by the SERVER_PORT environment variable
const HTTPPort = 8888
const HTTPSPort = 8443

type context struct {
	params   map[string]string
	hostname string
}

func (c *context) getParams() {
	c.params = make(map[string]string)
	httpPtr := flag.Int("http", HTTPPort, "http port value")
	httpsPtr := flag.Int("https", HTTPSPort, "https port value")
	flag.Parse()
	c.params["http"] = strconv.Itoa(*httpPtr)
	c.params["https"] = strconv.Itoa(*httpsPtr)
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
	http.HandleFunc("/", ctx.echoAll)
	http.HandleFunc("/hostname", ctx.echoHostname)
	log.Printf("starting echo server, listening on ports HTTP:%s/HTTPS:%s", ctx.params["http"], ctx.params["https"])
	// HTTPS
	go func() {
		listenAndServceTLS(ctx.params["https"])
	}()
	// HTTP
	err = http.ListenAndServe(":"+ctx.params["http"], nil)
	if err != nil {
		log.Fatal("Echo server (HTTP): ", err)
	}
}
