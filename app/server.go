package main

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
)

type connection struct {
	net.Conn
	error
}

var (
	redisServerObj  *server
	redisServerOnce sync.Once
)

type server struct {
}

func getServer() *server {
	redisServerOnce.Do(func() {
		redisServerObj = &server{}
	})
	return redisServerObj
}

func (s *server) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	defer listener.Close()
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
		//fmt.Println(connect.Error())
		return
	}
	buff := make([]byte, 2048)
	fmt.Println("Get ", connect.RemoteAddr())
	for {
		n, err := connect.Read(buff)
		if err != nil {
			//fmt.Println(err)
			return
		}
		requests := strings.Split(string(buff[:n]), "/r/n")
		fmt.Println("requets is", requests)
		//resp := processRedisMessages(requests)
		//fmt.Println("responce is", resp)
		//writed, err := connect.Write([]byte(resp))
		if err != nil {
			fmt.Println(err.Error())
		}
		//fmt.Println("Writed in responce: ", writed)
	}
}
