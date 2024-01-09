package processor

type setOptions interface {
	ApplyOption(*setCommand, []string) ([]string, error)
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

func (s *setCommand) ProcessMessages(messages []string) (ReturnValue, MessageProcessor, []string, error) {
	if s.state == 0 {
		resp, messages, err := parseArgument(messages)
		if err != nil {
			return resp, defaultParser, messages, err
		}
		//FIXME: CHECK KEY
		s.key = resp.Raw[0]
		s.state = 1
		return s.ProcessMessages(messages)
	} else if s.state == 1 {
		resp, messages, err := parseArgument(messages)
		if err != nil {
			return resp, defaultParser, messages, err
		}
		//FIXME: CHECK VALUE
		s.value = resp.Raw[0]
		s.state = 2
		return s.ProcessMessages(messages)
	} else if s.state == 2 {
		GetStorage().Set(s.key, s.value)
		return ReturnValue{
			Raw:       []string{"OK"},
			Processed: []string{simpleString("OK")},
		}, defaultParser, nil, nil
	}
	return ReturnValue{}, defaultParser, messages[1:], nil
}

func (s *setCommand) applyNextOption(messages []string) (bool, []string) {
	processor, newMessages, err := findProcessor(messages)
	if err != nil {
		return false, messages
	}
	switch processor.(type) {
	case *commandProcessor:
		value, _, newMessages, err := (&bulkStringParser{}).ProcessMessages(newMessages)
		if err != nil {
			return false, newMessages
		}
		option, find := s.findOption(value.Raw[0])
		if !find {
			return false, messages
		}
		newMessages, err = option.ApplyOption(s, newMessages)
		if err != nil {
			return false, messages
		}
		return true, newMessages
	}
	return false, messages
}

func (s *setCommand) findOption(name string) (option setOptions, isFind bool) {
	switch name {
	case "PX":

	}
	return nil, false
}

type timeOutOption struct {
}
