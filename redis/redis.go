package redis

type state struct {
}

type Session struct {
	state state
}

func NewSession() *Session {
	return &Session{}
}

func (s *Session) ProcessMessages(args []string) (resp string) {
	if len(args) == 0 {
		return ""
	}

	return redisMessageDistributor(args[0])(args...)
}

// * - array $ - string
func isArray(args []string) bool {
	if args[0][0] == '*' {
		return true
	}
	return false
}

func redisMessageDistributor(command string) (calculator func(args ...string) (resp string)) {
	switch command {
	default:
		return pingHandler
	}
}

func pingHandler(args ...string) (resp string) {
	return "+PONG\r\n"
}

func unknownHandler(args ...string) (resp string) {
	return ""
}
