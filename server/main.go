package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	zmq "github.com/pebbe/zmq4"
)

const (
	addr = "tcp://0.0.0.0:5555"
)

func main() {
	socket, err := zmq.NewSocket(zmq.REP)
	if nil != err {
		panic(err)
	}

	err = socket.Bind(addr)

	if nil != err {
		panic(err)
	}

	shutdown := make(chan struct{})
	go poller(socket, shutdown)

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	fmt.Print("server: shuting down...")
	close(shutdown)
}

func poller(socket *zmq.Socket, shutdown <-chan struct{}) {
	fmt.Printf("server: listening on %s...\n", addr)

	poller := zmq.NewPoller()
	poller.Add(socket, zmq.POLLIN)

loop:
	for {
		select {
		case <-shutdown:
			break loop
		default:
			sockets, _ := poller.Poll(-1)
			for _, sw := range sockets {
				switch s := sw.Socket; s {
				case socket:
					data, err := socket.RecvMessageBytes(0)
					if nil != err {
						fmt.Printf("received error: %s\n", err)
						return
					}

					fmt.Printf("received message %q\n", string(data[0]))

					fmt.Printf("send back to client...\n")
					_, err = s.Send(fmt.Sprintf("rep_%s", string(data[0])), 0)
					if nil != err {
						fmt.Printf("send back to client error: %s\n", err)
					} else {
						fmt.Print("send back to client success\n")
					}
				}
			}
		}
	}
	fmt.Print("stopped")
}
