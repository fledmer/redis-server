package processor

import (
	"errors"
	"strconv"
	"strings"
	"sync"
)

type MessageProcessor interface {
	ProcessMessages([]string) (ReturnValue, MessageProcessor, []string, error)
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

type ReturnValue struct {
	Raw       []string
	Processed []string
}

func (r ReturnValue) Plus(ar ReturnValue) ReturnValue {
	return ReturnValue{
		Raw:       append(r.Raw, ar.Raw...),
		Processed: append(r.Processed, ar.Processed...),
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
	var (
		resp      ReturnValue
		fullResp  ReturnValue
		err       error
		processor MessageProcessor
	)
	for len(messagesSlice) != 0 {
		resp, processor, messagesSlice, err = f.processor.ProcessMessages(messagesSlice)
		if err != nil {
			return errorToMessage(err)
		}
		fullResp = fullResp.Plus(resp)
		f.processor = processor
	}

	return strings.Join(fullResp.Processed, "")
}

// Добавить Context в котором будет все для запрсов, напирмер Хранилище
type baseParser struct {
}

func (i *baseParser) ProcessMessages(messages []string) (ReturnValue, MessageProcessor, []string, error) {
	processor, messages, err := findProcessor(messages)
	if err != nil {
		return ReturnValue{}, nil, nil, err
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

func (ap *arrayParser) ProcessMessages(messages []string) (ReturnValue, MessageProcessor, []string, error) {
	var (
		processor MessageProcessor = &baseParser{}
		err       error
		resp      ReturnValue
		fullResp  ReturnValue
	)
	messagesSlice := messages[:ap.count*2]
	for len(messagesSlice) != 0 {
		resp, processor, messagesSlice, err = (processor).ProcessMessages(messagesSlice)
		if err != nil {
			return ReturnValue{}, nil, nil, err
		}
		fullResp = fullResp.Plus(resp)
	}
	return fullResp, processor, messages[ap.count*2:], nil
}

type bulkStringParser struct {
}

func (bsp *bulkStringParser) ProcessMessages(messages []string) (ReturnValue, MessageProcessor, []string, error) {
	return ReturnValue{
		Raw: []string{messages[0]},
	}, defaultParser, messages[1:], nil
}

type commandProcessor struct {
}

func (cp *commandProcessor) ProcessMessages(messages []string) (ReturnValue, MessageProcessor, []string, error) {
	switch messages[0] {
	case "PING":
		//TODO: maybe move ping logic to own struct
		return ReturnValue{
			Processed: []string{simpleString("PONG")},
		}, defaultParser, messages[1:], nil
	case "ECHO":
		return (&echoCommand{}).ProcessMessages(messages[1:])
	case "SET":
		return (&setCommand{}).ProcessMessages(messages[1:])
	case "GET":
		return (&getCommand{}).ProcessMessages(messages[1:])
	default:
		return ReturnValue{}, nil, messages[:], UnknownCommandErr{}
	}
}

type echoCommand struct {
}

func (p *echoCommand) ProcessMessages(messages []string) (ReturnValue, MessageProcessor, []string, error) {
	message, proc, messages, err := (&bulkStringParser{}).ProcessMessages(messages[1:])
	return ReturnValue{
		Raw:       []string{message.Raw[0]},
		Processed: []string{simpleString(message.Raw[0])},
	}, proc, messages, err
}

type getCommand struct {
}

func (s *getCommand) ProcessMessages(messages []string) (ReturnValue, MessageProcessor, []string, error) {
	processor, messages, err := findProcessor(messages)
	if err != nil {
		return ReturnValue{}, nil, messages, err
	}
	switch processor.(type) {
	case *commandProcessor:
		message, _, messages, err := (&bulkStringParser{}).ProcessMessages(messages)
		if err != nil {
			return ReturnValue{}, nil, messages, err
		}
		val, find := GetStorage().Get(message.Raw[0])
		if !find {
			val = "-1"
		}
		return ReturnValue{
			Raw:       []string{val.(string)},
			Processed: []string{simpleString(val.(string))},
		}, &baseParser{}, messages, nil
	default:
		return ReturnValue{}, nil, nil, errors.New("failed to parse")
	}
}

func simpleString(raw string) string {
	builder := strings.Builder{}
	builder.WriteString("+")
	builder.WriteString(raw)
	builder.WriteString("\r\n")
	return builder.String()
}

// * - array $ - command
func findProcessor(messages []string) (MessageProcessor, []string, error) {
	if len(messages) == 0 || len(messages[0]) == 0 {
		return nil, nil, errors.New("failed to parse")
	}
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
		//FIXME:
		return nil, messages, errors.New("failed to parse")
	}
}

func errorToMessage(err error) string {
	builder := strings.Builder{}
	builder.WriteString("!")
	builder.WriteString(err.Error())
	builder.WriteString("\r\n")
	return builder.String()
}

func parseArgument(messages []string) (ReturnValue, []string, error) {
	processor, messages, err := findProcessor(messages)
	if err != nil {
		return ReturnValue{}, messages, err
	}
	switch p := processor.(type) {
	case *commandProcessor:
		resp, _, messages, err := p.ProcessMessages(messages)
		if errors.Is(err, UnknownCommandErr{}) {
			resp, _, messages, err = (&bulkStringParser{}).ProcessMessages(messages)
		}
		return resp, messages, err
	case *bulkStringParser:
		resp, _, messages, err := (&bulkStringParser{}).ProcessMessages(messages)
		return resp, messages, err
	default:
		return ReturnValue{}, messages, errors.New("failed to parse")
	}
}

//TODO: нужно возвращать из процессора не просто строку, а структуру внутри которой будут аргументы RAW аргументы и возможность спарсить их
