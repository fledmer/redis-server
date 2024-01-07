package processor

import (
	"strings"
	"testing"
)

func MessagesToRaw(messages []string) string {
	str := strings.Join(messages, "\r\n")
	//if len(messages) != 0 {
	//	str += "\r\n"
	//}
	return str
}

func Test_facade_ProcessMessages(t *testing.T) {
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
			name: "echo",
			args: args{messages: MessagesToRaw([]string{"*2", "$4", "ECHO", "$3", "hey"})},
			want: simpleString("hey"),
			fields: fields{
				processor: defaultParser,
			},
		},
		{
			name: "One full ping",
			args: args{messages: MessagesToRaw([]string{"*1", "$4", "PING"})},
			want: simpleString("PONG"),
			fields: fields{
				processor: defaultParser,
			},
		},
		{
			name: "One full ping",
			args: args{messages: MessagesToRaw([]string{"*1", "$4", "PING"})},
			want: simpleString("PONG"),
			fields: fields{
				processor: defaultParser,
			},
		},
		{
			name: "Two full ping",
			args: args{messages: MessagesToRaw([]string{"*2", "$4", "PING", "$4", "PING"})},
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
			args: args{messages: MessagesToRaw([]string{"*4", "$4", "PING", "$4", "PING", "$4", "PING", "$4", "PING"})},
			want: simpleString("PONG") + simpleString("PONG") + simpleString("PONG") + simpleString("PONG"),
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
				t.Errorf("facade.ProcessMessages() = %v, want %v", got, tt.want)
			}
		})
	}
}
