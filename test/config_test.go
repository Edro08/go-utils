package config

import (
	"github.com/edro08/go-utils/config"
	"reflect"
	"testing"
)

func TestConfig_KeyEmpty(t *testing.T) {
	opts := config.Opts{
		File: "app.yaml",
	}

	cfg, err := config.NewConfig(opts)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		key string
		fn  func(string) any
	}

	tests := []struct {
		name string
		args args
		want any
	}{
		{
			name: "GetString(): key empty",
			args: args{
				key: "No.Exists",
				fn:  func(k string) any { return cfg.GetString(k) },
			},
			want: "",
		},
		{
			name: "GetInt(): key empty",
			args: args{
				key: "No.Exists",
				fn:  func(k string) any { return cfg.GetInt(k) },
			},
			want: 0,
		},
		{
			name: "GetBool(): key empty",
			args: args{
				key: "No.Exists",
				fn:  func(k string) any { return cfg.GetBool(k) },
			},
			want: false,
		},
		{
			name: "GetFloat(): key empty",
			args: args{
				key: "No.Exists",
				fn:  func(k string) any { return cfg.GetFloat(k) },
			},
			want: 0.0,
		},
		{
			name: "GetMapString(): key empty",
			args: args{
				key: "No.Exists",
				fn:  func(k string) any { return cfg.GetMapString(k) },
			},
			want: map[string]string{},
		},
		{
			name: "GetMapInt(): key empty",
			args: args{
				key: "No.Exists",
				fn:  func(k string) any { return cfg.GetMapInt(k) },
			},
			want: map[string]int{},
		},
		{
			name: "GetMapInterface(): key empty",
			args: args{
				key: "No.Exists",
				fn:  func(k string) any { return cfg.GetMap(k) },
			},
			want: map[string]interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.fn(tt.args.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("key = %q, got = %v, want = %v", tt.args.key, got, tt.want)
			}
		})
	}
}
