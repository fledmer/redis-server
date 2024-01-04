package redis

import (
	"errors"
	"strconv"
	"strings"
)

// * - array $ - string

type storage struct {
	storage map[string]any
}

var storage storage

func init() {
	storage = storage{
		storage: make(map[string]any),
	}
}

type MessageProcessor interface {
	ProcessMessages([]string) (string, MessageProcessor, []string, error)
}

type StartProcessor struct {
}

func (i *StartProcessor) ProcessMessages(messages []string) (string, MessageProcessor, []string, error) {
	(&initProcessor{}).ProcessMessages(messages)

}

type initProcessor struct {
}

func (i *initProcessor) ProcessMessages(messages []string) (string, MessageProcessor, []string, error) {
	processor, messages, err := messageToProcessor(messages)
	if err != nil {
		return "", nil, nil, err
	}
	return processor.ProcessMessages(messages)
}

type arrayProcesser struct {
	count    int
	commands []string
}

func newArrayProcessor(count int) *arrayProcesser {
	return &arrayProcesser{
		count: count,
	}
}

func (ap *arrayProcesser) ProcessMessages(messages []string) (string, MessageProcessor, []string, error) {
	var (
		processor MessageProcessor
		err       error
		rawResp   string
		//messages  []string
	)
	for element := 0; element < ap.count; element++ {
		processor, messages, err = messageToProcessor(messages)
		if err != nil {
			return "", nil, nil, err
		}
		rawResp, _, messages, err = processor.ProcessMessages(messages)
		if err != nil {
			return "", nil, nil, err
		}
		ap.commands = append(ap.commands, rawResp)
	}

	var (
		message string
	)
	processor = &initProcessor{}
	resp := strings.Builder{}
	messages = ap.commands
	for len(messages) != 0 {
		message, processor, messages, err = processor.ProcessMessages(messages)
		if err != nil {
			return "", nil, nil, err
		}
		resp.WriteString(message)
	}
	//processor.ProcessMessages([]string{"end"})
	return resp.String(), processor, messages, err
}

type bulkStringsProcessor struct {
}

func (bsp *bulkStringsProcessor) ProcessMessages(messages []string) (string, MessageProcessor, []string, error) {
	return messages[0], &initProcessor{}, messages[1:], nil
}

type commandProcessor struct {
}

func (bsp *commandProcessor) ProcessMessages(messages []string) (string, MessageProcessor, []string, error) {
	switch messages[0] {
	case "PING":
		return newSimpleString("PONG"), &initProcessor{}, messages[1:], nil
	case "ECHO":
		return "", &echoProcessor{}, messages[1:], nil
	default:
		return "", nil, nil, errors.New("failed to parse")
	}
}

type echoProcessor struct {
}

func (p *echoProcessor) ProcessMessages(messages []string) (string, MessageProcessor, []string, error) {
	return newSimpleString(messages[0]), &initProcessor{}, messages[1:], nil
}

type setProcessor struct {
}

func (p *setProcessor) ProcessMessages(messages []string) (string, MessageProcessor, []string, error) {
	return newSimpleString(messages[0]), &initProcessor{}, messages[1:], nil
}

func newSimpleString(raw string) string {
	builder := strings.Builder{}
	builder.WriteString("+")
	builder.WriteString(raw)
	builder.WriteString("\r\n")
	return builder.String()
}

// * - array $ - string
func messageToProcessor(messages []string) (MessageProcessor, []string, error) {
	switch messages[0][0] {
	case '*':
		count, err := strconv.Atoi(messages[0][1:])
		if err != nil {
			return nil, messages[1:], err
		}
		return &arrayProcesser{count: count}, messages[1:], nil
	case '$':
		return &bulkStringsProcessor{}, messages[1:], nil
	default:
		return &commandProcessor{}, messages, nil
	}
}
