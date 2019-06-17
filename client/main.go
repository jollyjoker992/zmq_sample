package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	zmq "github.com/pebbe/zmq4"
)

const (
	alphanumeric = "QWERTYUIOPASDFGHJKLZXCVBNMqwertyuiopasdfghjklzxcvbnm1234567890"
	addr         = "tcp://127.0.0.1:5555"
)

func main() {
	socket, err := zmq.NewSocket(zmq.REQ)
	if nil != err {
		fmt.Printf("new socket error: %s", err)
		panic(err)
	}

	// set identity
	randomIDBytes := make([]byte, 32)
	_, err = rand.Read(randomIDBytes)
	if nil != err {
		panic(err)
	}
	randomIdentifier := string(randomIDBytes)
	socket.SetIdentity(randomIdentifier)

	// set timeout
	socket.SetConnectTimeout(20 * time.Second)

	err = socket.Connect(addr)

	shutdown := make(chan struct{})
	go poller(socket, shutdown)

	timer := time.After(5 * time.Second)
loop:
	for {
		select {
		case <-shutdown:
			break loop
		case <-timer:
			timer = time.After(5 * time.Second)
			m := alphanumeric[rand.Intn(len(alphanumeric)-1)]

			fmt.Printf("sending %q to server...\n", m)
			_, err = socket.SendBytes([]byte{m}, 0)
			if nil != err {
				fmt.Printf("sending error: %s\n", err)
			} else {
				fmt.Print("sending success\n")
			}

			fmt.Print("receiving reply...\n")
			mess, err := socket.Recv(0)
			if nil != err {
				fmt.Printf("receive message from server error: %s\n", err)
			} else {
				fmt.Printf("receive message from server success %q\n", mess)
			}
		}
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	fmt.Print("shutting down...\n")
	close(shutdown)
	fmt.Print("stopped\n")
}

func poller(socket *zmq.Socket, shutdown <-chan struct{}) error {
	const monitorSig = "inproc://monitor-signal-client"

	err := socket.Monitor(monitorSig, zmq.EVENT_ALL)
	if nil != err {
		return err
	}

	mon, err := zmq.NewSocket(zmq.PAIR)
	if nil != err {
		return err
	}

	err = mon.Connect(monitorSig)
	if nil != err {
		mon.Close()
		return err
	}

	poller := zmq.NewPoller()
	// poller.Add(socket, zmq.POLLIN)
	poller.Add(mon, zmq.POLLIN)

loop:
	for {
		select {
		case <-shutdown:
			break loop
		default:
			sockets, _ := poller.Poll(-1)
			for _, sw := range sockets {
				switch s := sw.Socket; s {
				case mon:
					ev, addr, v, err := s.RecvEvent(0)

					fmt.Printf("event: %q  address: %q  value: %d err: %s\n", ev, addr, v, err)
					if nil != err {
						return err
					}
				}
			}
		}
	}
	return nil
}
