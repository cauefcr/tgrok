package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/clementauger/tor-prebuilt/embedded"
	"github.com/cretz/bine/tor"
)

// [x] tunnel any tcp with embedded tor server (no udp on tor)
// [ ] flags
// [ ] save and reuse private key

func main() {
	if len(os.Args) != 2 {
		os.Exit(80085)
	}
	// Start tor with default config (can set start conf's DebugWriter to os.Stdout for debug logs)
	fmt.Println("Starting and registering onion service, please wait a couple of minutes...")
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	t, err := tor.Start(nil, &tor.StartConf{ProcessCreator: embedded.NewCreator(), DataDir: path + "/tor-dir", TorrcFile: "torrc-defaults"})
	// t.Control.DebugWriter = nil
	if err != nil {
		log.Panicf("Unable to start Tor: %v", err)
	}
	defer t.Close()
	// Wait at most a few minutes to publish the service
	listenCtx, listenCancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer listenCancel()
	// Create a v3 onion service to listen on any port but show as 80
	onion, err := t.Listen(listenCtx, &tor.ListenConf{Version3: true, RemotePorts: []int{80}})
	if err != nil {
		log.Panicf("Unable to create onion service: %v", err)
	}
	defer onion.Close()
	fmt.Printf("Open Tor browser and navigate to http://%v.onion\n", onion.ID)
	fmt.Println("Press enter to exit")
	for c := range clientConns(onion) {
		go handleConn(os.Args[1], c)
	}
}

func clientConns(listener net.Listener) chan net.Conn {
	ch := make(chan net.Conn)
	i := 0
	go func() {
		for {
			client, err := listener.Accept()
			if client == nil {
				fmt.Println("couldn't accept: ", err)
				continue
			}
			i++
			fmt.Printf("%d: %v <-> %v\n", i, client.LocalAddr(), client.RemoteAddr())
			ch <- client
		}
	}()
	return ch
}

func handleConn(local string, client net.Conn) {
	conn, err := net.Dial("tcp", local)
	defer conn.Close()
	if err != nil {
		panic(err)
	}
	go io.Copy(conn, client)
	io.Copy(client, conn)
}
