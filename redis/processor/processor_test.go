package processor

import (
	"redis-server/redis/memory_storage"
	"strconv"
	"strings"
	"testing"
)

func createMessage(args ...string) string {
	resp := []string{}
	argsCount := len(args)
	resp = append(resp, "*"+strconv.Itoa(argsCount))
	for _, obj := range args {
		resp = append(resp, "$"+strconv.Itoa(len(obj)))
		resp = append(resp, obj)
	}
	return messagesToRaw(resp)
}

func messagesToRaw(messages []string) string {
	str := strings.Join(messages, "\r\n")
	if len(messages) != 0 {
		str += "\r\n"
	}
	return str
}

func Test_facade_ProcessMessages(t *testing.T) {
	InitStorage(memory_storage.New())
	type fields struct {
		processor MessageProcessor
	}
	type args struct {
		messages string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "get inside set",
			args: args{messages: createMessage("SET", "K1", "K2") +
				createMessage("SET", "GET", "K1", "V2") +
				createMessage("GET", "K2")},
			want: simpleString("OK") + simpleString("OK") + simpleString("V2"),
			fields: fields{
				processor: defaultParser,
			},
		},
		{
			name: "echo",
			args: args{messages: createMessage("ECHO", "hey")},
			want: simpleString("hey"),
			fields: fields{
				processor: defaultParser,
			},
		},
		{
			name: "One full ping",
			args: args{messages: createMessage("PING")},
			want: simpleString("PONG"),
			fields: fields{
				processor: defaultParser,
			},
		},
		{
			name: "Two full ping",
			args: args{messages: createMessage("PING", "PING")},
			want: simpleString("PONG") + simpleString("PONG"),
			fields: fields{
				processor: defaultParser,
			},
		},
		/*{
			name: "simple ping",
			args: args{messages: MessagesToRaw([]string{"PING"})},
			want: simpleString("PONG"),
			fields: fields{
				processor: defaultParser,
			},
		},*/
		{
			name: "4 full ping",
			args: args{messages: createMessage("PING", "PING", "PING", "PING")},
			want: simpleString("PONG") + simpleString("PONG") + simpleString("PONG") + simpleString("PONG"),
			fields: fields{
				processor: defaultParser,
			},
		},
		{
			name: "simple set && simple get",
			args: args{messages: createMessage("SET", "K1", "V1") +
				createMessage("GET", "K1")},
			want: simpleString("OK") + simpleString("V1"),
			fields: fields{
				processor: defaultParser,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &facade{
				processor: tt.fields.processor,
			}
			if got := f.ProcessMessages(tt.args.messages); got != tt.want {
				t.Errorf("facade.ProcessMessages() = \n %v \n want: \n %v", got, tt.want)
			}
		})
	}
}
