package main

import (
	"context"
	"fmt"
	"net"
	"sync"
)

type server struct {
}

type connection struct {
	net.Conn
	error
}

var redisServerObj *server
var redisServerOnce sync.Once

func getServer() *server {
	redisServerOnce.Do(func() {
		redisServerObj = &server{}
	})
	return redisServerObj
}

func (s *server) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		return err
	}
	conChan := make(chan connection)
	go s.runConnectionLoop(ctx, conChan, listener)
	for {
		select {
		case <-ctx.Done():
			return nil
		case newConnect := <-conChan:
			if newConnect.error != nil {
				//println(newConnect.Error())
				newConnect.Close()
				continue
			}
			buff := make([]byte, 2048)
			fmt.Println("Get ", newConnect.RemoteAddr())
			_, err := newConnect.Read(buff)
			if err != nil {
				//fmt.Println(err)
				newConnect.Close()
				continue
			}
			println("requets is", string(buff[:]))
			PushRedisMessage(string(buff))
			newConnect.Close()
		}
	}
}

func (s *server) runConnectionLoop(ctx context.Context, conChan chan connection, listener net.Listener) {
	con, err := listener.Accept()
	conChan <- connection{con, err}
}

func PushRedisMessage(args ...string) (resp string) {
	if len(args) < 1 {
		return ""
	}
	return redisMessageDistributor(args[0])(args...)
}

func redisMessageDistributor(command string) (calculator func(args ...string) (resp string)) {
	switch command {
	case "PING":
		return pingHandler
	default:
		return unknownHandler
	}
}

func pingHandler(args ...string) (resp string) {
	return "PONG"
}

func unknownHandler(args ...string) (resp string) {
	return ""
}

func main() {
	println("starting")
	getServer().Run(context.Background())
}
