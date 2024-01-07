package processor

import (
	"errors"
	"strconv"
	"strings"
	"sync"
)

/*
*2
$5
myKEY
$5
HELLO
*/

type MessageProcessor interface {
	ProcessMessages([]string) (string, MessageProcessor, []string, error)
}

type Storage interface {
	Set(string, any)
	Get(string) (any, bool)
}

func GetStorage() Storage {
	return store
}

var (
	once          sync.Once
	defaultParser = &baseParser{}
	store         Storage
)

func InitStorage(s Storage) {
	once.Do(func() {
		store = s
	})
}

func NewProcessor() *facade {
	return &facade{
		processor: defaultParser,
	}
}

type facade struct {
	processor MessageProcessor
}

func (f *facade) ProcessMessages(messages string) string {
	messagesSlice := strings.Split(messages, "\r\n")
	if len(messagesSlice) > 1 {
		messagesSlice = messagesSlice[:len(messagesSlice)-1]
	}
	resp, processor, _, err := f.processor.ProcessMessages(messagesSlice)
	if err != nil {
		return errorToMessage(err)
	}
	f.processor = processor
	return resp
}

// Добавить Context в котором будет все для запрсов, напирмер Хранилище
type baseParser struct {
}

func (i *baseParser) ProcessMessages(messages []string) (string, MessageProcessor, []string, error) {
	processor, messages, err := findProcessor(messages)
	if err != nil {
		return "", nil, nil, err
	}
	return processor.ProcessMessages(messages)
}

type arrayParser struct {
	count    int
	commands []string
}

func newArrayProcessor(count int) *arrayParser {
	return &arrayParser{
		count: count,
	}
}

func (ap *arrayParser) ProcessMessages(messages []string) (string, MessageProcessor, []string, error) {
	var (
		processor MessageProcessor
		err       error
		resp      string
		fullResp  string
	)
	for len(messages) != 0 {
		resp, processor, messages, err = (&commandProcessor{}).ProcessMessages(messages)
		if err != nil {
			return "", nil, nil, err
		}
		fullResp += resp
	}
	return fullResp, processor, messages, nil
}

type bulkStringParser struct {
}

func (bsp *bulkStringParser) ProcessMessages(messages []string) (string, MessageProcessor, []string, error) {
	return messages[0], defaultParser, messages[1:], nil
}

type commandOrStringProcessor struct {
}

func (cp *commandOrStringProcessor) ProcessMessages(messages []string) (string, MessageProcessor, []string, error) {
	resp, processor, messages, err := (&commandProcessor{}).ProcessMessages(messages)
	if err != nil {
		return (&bulkStringParser{}).ProcessMessages(messages)
	}
	return resp, processor, messages, nil
}

type commandProcessor struct {
}

func (cp *commandProcessor) ProcessMessages(messages []string) (string, MessageProcessor, []string, error) {
	if messages[0][0] != '$' {
		return "", nil, messages, errors.New("failed to parse")
	}
	if len(messages) == 1 {
		return "", nil, messages, errors.New("failed to parse")
	}
	switch messages[1] {
	case "PING":
		//TODO: maybe move ping logic to own struct
		return simpleString("PONG"), defaultParser, messages[2:], nil
	case "ECHO":
		return (&echoCommand{}).ProcessMessages(messages[2:])
	case "SET":
		return (&setCommand{}).ProcessMessages(messages[2:])
	case "GET":
		return (&getCommand{}).ProcessMessages(messages[2:])
	default:
		return "", nil, messages[2:], errors.New("failed to parse")
	}
}

type echoCommand struct {
}

func (p *echoCommand) ProcessMessages(messages []string) (string, MessageProcessor, []string, error) {
	resp, proc, messages, err := (&bulkStringParser{}).ProcessMessages(messages[1:])
	resp = simpleString(resp)
	return resp, proc, messages, err
}

type setCommand struct {
	/*
		1 - прочитан key
		2 - прочитан value
		3 - читает остальные аргументы
	*/
	state  int
	key    string
	value  string
	params []string
}

func (s *setCommand) ProcessMessages(messages []string) (string, MessageProcessor, []string, error) {
	if s.state == 2 {
		GetStorage().Set(s.key, s.value)
		return simpleString("OK"), defaultParser, nil, nil
	}
	processor, messages, err := findProcessor(messages)
	if err != nil {
		return "", nil, messages, err
	}
	if s.state == 0 {
		switch processor.(type) {
		case *commandProcessor:
			message, _, messages, err := (&bulkStringParser{}).ProcessMessages(messages)
			if err != nil {
				return "", nil, messages, err
			}
			//FIXME: check key
			s.key = message
			s.state = 1
			return s.ProcessMessages(messages)
		default:
			return "", nil, nil, errors.New("failed to parse")
		}
	} else if s.state == 1 {
		switch processor.(type) {
		case *commandProcessor:
			message, _, messages, err := (&bulkStringParser{}).ProcessMessages(messages)
			if err != nil {
				return "", nil, messages, err
			}
			//FIXME: check key
			s.value = message
			s.state = 2
			return s.ProcessMessages(messages)
		default:
			return "", nil, nil, errors.New("failed to parse")
		}
	}
	panic("state error")
	return simpleString(messages[0]), defaultParser, messages[1:], nil
}

type getCommand struct {
}

func (s *getCommand) ProcessMessages(messages []string) (string, MessageProcessor, []string, error) {
	processor, messages, err := findProcessor(messages)
	if err != nil {
		return "", nil, messages, err
	}
	switch processor.(type) {
	case *commandProcessor:
		message, _, messages, err := (&bulkStringParser{}).ProcessMessages(messages)
		if err != nil {
			return "", nil, messages, err
		}
		val, find := GetStorage().Get(message)
		if !find {
			val = "-1"
		}
		return simpleString(val.(string)), &baseParser{}, messages, nil
	default:
		return "", nil, nil, errors.New("failed to parse")
	}
}

func simpleString(raw string) string {
	builder := strings.Builder{}
	builder.WriteString("+")
	builder.WriteString(raw)
	builder.WriteString("\r\n")
	return builder.String()
}

// * - array $ - string
func findProcessor(messages []string) (MessageProcessor, []string, error) {
	switch messages[0][0] {
	case '*':
		count, err := strconv.Atoi(messages[0][1:])
		if err != nil {
			return nil, messages[1:], err
		}
		return &arrayParser{count: count}, messages[1:], nil
	case '$':
		return &commandProcessor{}, messages[1:], nil
	default:
		return &commandProcessor{}, messages, nil
	}
}

func errorToMessage(err error) string {
	builder := strings.Builder{}
	builder.WriteString("!")
	builder.WriteString(err.Error())
	builder.WriteString("\r\n")
	return builder.String()
}
