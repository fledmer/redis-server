package main

import (
	"context"
	"fmt"
	"net"
	"redis-server/redis"
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
	server *redis.Server
}

func getServer() *server {
	redisServerOnce.Do(func() {
		redisServerObj = &server{
			server: redis.NewServer(),
		}
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
	session := s.server.NewSession()
	for {
		n, err := connect.Read(buff)
		if err != nil {
			//fmt.Println(err)
			return
		}
		rawRequest := buff[:n]
		fmt.Println("requets is", string(rawRequest))
		resp := session.Process(string(rawRequest))
		fmt.Println("responce is", resp)
		_, err = connect.Write([]byte(resp))
		if err != nil {
			fmt.Println("Error!", err.Error())
		}
	}
}
