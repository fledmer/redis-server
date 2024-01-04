package redis

import (
	"reflect"
	"testing"
)

func Test_initProcessor_ProcessMessages(t *testing.T) {
	type args struct {
		messages []string
	}
	tests := []struct {
		name    string
		i       *initProcessor
		args    args
		want    string
		want1   messageProcessor
		want2   []string
		wantErr bool
	}{
		{
			name:    "One full ping",
			i:       &initProcessor{},
			args:    args{messages: []string{"*1", "$4", "PING"}},
			want:    newSimpleString("PONG"),
			want1:   &initProcessor{},
			want2:   []string{},
			wantErr: false,
		},
		{
			name:    "Two full ping",
			i:       &initProcessor{},
			args:    args{messages: []string{"*2", "$4", "PING", "$4", "PING"}},
			want:    newSimpleString("PONG") + newSimpleString("PONG"),
			want1:   &initProcessor{},
			want2:   []string{},
			wantErr: false,
		},
		{
			name:    "simple ping",
			i:       &initProcessor{},
			args:    args{messages: []string{"PING"}},
			want:    newSimpleString("PONG"),
			want1:   &initProcessor{},
			want2:   []string{},
			wantErr: false,
		},
		{
			name:    "4 full ping",
			i:       &initProcessor{},
			args:    args{messages: []string{"*4", "$4", "PING", "$4", "PING", "$4", "PING", "$4", "PING"}},
			want:    newSimpleString("PONG") + newSimpleString("PONG") + newSimpleString("PONG") + newSimpleString("PONG"),
			want1:   &initProcessor{},
			want2:   []string{},
			wantErr: false,
		},
		{
			name:    "echo",
			i:       &initProcessor{},
			args:    args{messages: []string{"*2", "$4", "ECHO", "$3", "hey"}},
			want:    newSimpleString("hey"),
			want1:   &initProcessor{},
			want2:   []string{},
			wantErr: false,
		},
		//{
		//	name:    "simple 2 ping",
		//	i:       &initProcessor{},
		//	args:    args{messages: []string{"PING", "PING"}},
		//	want:    newSimpleString("PONG") + newSimpleString("PONG"),
		//	want1:   &initProcessor{},
		//	want2:   []string{},
		//	wantErr: false,
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := tt.i.ProcessMessages(tt.args.messages)
			if (err != nil) != tt.wantErr {
				t.Errorf("initProcessor.ProcessMessages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("initProcessor.ProcessMessages() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("initProcessor.ProcessMessages() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("initProcessor.ProcessMessages() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
