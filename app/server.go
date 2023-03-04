package main

import (
	"context"
	"fmt"
	"net"
	"strings"
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
		case connect := <-conChan:
			s.HandleConnect(connect)
		}
	}
}

func (s *server) runConnectionLoop(ctx context.Context, conChan chan connection, listener net.Listener) {
	for {
		con, err := listener.Accept()
		conChan <- connection{con, err}
	}
}

func (s *server) HandleConnect(connect connection) {
	defer connect.Close()
	if connect.error != nil {
		//println(connect.Error())
		return
	}
	buff := make([]byte, 256)
	fmt.Println("Get ", connect.RemoteAddr())
	for {
		n, err := connect.Read(buff)
		if err != nil {
			//fmt.Println(err)
			return
		}
		request := string(buff[:n])
		println("requets is", request)
		resp := PushRedisMessage(strings.Split(request, "/r/n")...)
		println("responce is", resp)
		writed, err := connect.Write([]byte(resp + "\n"))
		if err != nil {
			println(err.Error())
		}
		println("Writed in responce: ", writed)
	}
}

func PushRedisMessage(args ...string) (resp string) {
	if len(args) < 1 {
		return ""
	}
	return redisMessageDistributor(args[0])(args...)
}

func redisMessageDistributor(command string) (calculator func(args ...string) (resp string)) {
	switch command {
	default:
		return pingHandler
	}
}

func pingHandler(args ...string) (resp string) {
	return "+pong\r\n"
}

func unknownHandler(args ...string) (resp string) {
	return ""
}

func main() {
	println("starting")
	getServer().Run(context.Background())
}
