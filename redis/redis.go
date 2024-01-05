package redis

import (
	"redis-server/redis/memory_storage"
	"redis-server/redis/processor"
)

type MessageProcessor interface {
	ProcessMessages(string) string
}

type Server struct {
}

func NewServer() *Server {
	processor.InitStorage(memory_storage.New())
	return &Server{}
}

func (s *Server) NewSession() *Session {
	return &Session{
		processor: processor.NewProcessor(),
	}
}

type Session struct {
	processor MessageProcessor
}

func NewSession() *Session {
	return &Session{
		processor: processor.NewProcessor(),
	}
}

func (s *Session) Process(message string) string {
	return s.processor.ProcessMessages(message)
}
