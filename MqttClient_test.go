package MqttPool

import (
	"testing"

	"github.com/gogf/gf/frame/g"
)

func TestPublishMsg(t *testing.T) {

	g.Cfg().SetFileName("Config.ini")

	type args struct {
		param *MqttContext
	}

	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{"", args{param: &MqttContext{Topic: "test"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MqttPublishMsg(tt.args.param, 3, 5)
		})
	}
}
