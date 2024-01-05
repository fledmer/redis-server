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
	resp, processor, _, err := f.processor.ProcessMessages(strings.Split(messages, "\r\n"))
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
		rawResp   string
	)
	for element := 0; element < ap.count; element++ {
		processor, messages, err = findProcessor(messages)
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
	processor = defaultParser
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

type bulkStringParser struct {
}

func (bsp *bulkStringParser) ProcessMessages(messages []string) (string, MessageProcessor, []string, error) {
	return messages[0], defaultParser, messages[1:], nil
}

type commandParser struct {
}

func (bsp *commandParser) ProcessMessages(messages []string) (string, MessageProcessor, []string, error) {
	switch messages[0] {
	case "PING":
		return simpleString("PONG"), defaultParser, messages[1:], nil
	case "ECHO":
		return "", &echoCommand{}, messages[1:], nil
	case "SET":
		return "", &setCommand{}, messages[1:], nil
	default:
		return "", nil, nil, errors.New("failed to parse")
	}
}

type echoCommand struct {
}

func (p *echoCommand) ProcessMessages(messages []string) (string, MessageProcessor, []string, error) {
	return simpleString(messages[0]), defaultParser, messages[1:], nil
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
		return simpleString("OK"), defaultParser, messages, nil
	}
	processor, messages, err := findProcessor(messages)
	if err != nil {
		return "", nil, messages, err
	}
	if s.state == 0 {
		switch p := processor.(type) {
		case *bulkStringParser:
			message, _, _, err := p.ProcessMessages(messages)
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
		switch p := processor.(type) {
		case *bulkStringParser:
			message, _, _, err := p.ProcessMessages(messages)
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
		return &bulkStringParser{}, messages[1:], nil
	default:
		return &commandParser{}, messages, nil
	}
}

func errorToMessage(err error) string {
	builder := strings.Builder{}
	builder.WriteString("!")
	builder.WriteString(err.Error())
	builder.WriteString("\r\n")
	return builder.String()
}
