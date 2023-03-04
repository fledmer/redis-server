package main

func PushMessage(args ...string) (resp string) {
	if len(args) < 1 {
		return ""
	}
	return messageDistributor(args[0])(args...)
}

func messageDistributor(command string) (calculator func(args ...string) (resp string)) {
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
