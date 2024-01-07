package processor

type UnknownCommandErr struct {
}

func (UnknownCommandErr) Error() string {
	return "unknown command"
}
