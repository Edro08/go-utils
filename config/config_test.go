package config

import (
	"reflect"
	"sort"
	"testing"
)

func TestConfig_Get(t *testing.T) {
	sampleData := map[string]interface{}{
		"name": "my-app",
		"server": map[string]interface{}{
			"port": 8080,
		},
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		keys string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		{
			name: "Get existing top-level value",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "name"},
			want: "my-app",
		},
		{
			name: "Get existing nested value",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "server.port"},
			want: 8080,
		},
		{
			name: "Get non-existing key returns nil",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "non.existent"},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			if got := c.Get(tt.args.keys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetBool(t *testing.T) {
	sampleData := map[string]interface{}{
		"enabled":  true,
		"disabled": false,
		"strBool":  "true",
		"invalid":  "not-a-bool",
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		keys string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Direct bool true",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "enabled"},
			want: true,
		},
		{
			name: "Direct bool false",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "disabled"},
			want: false,
		},
		{
			name: "String parsed to bool",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "strBool"},
			want: true,
		},
		{
			name: "Invalid string defaults to false",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "invalid"},
			want: false,
		},
		{
			name: "Non-existent key defaults to false",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "missing"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			if got := c.GetBool(tt.args.keys); got != tt.want {
				t.Errorf("GetBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetFloat(t *testing.T) {
	sampleData := map[string]interface{}{
		"floatVal": 3.14,
		"intVal":   100,
		"strFloat": "2.718",
		"invalid":  "abc",
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		keys string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
	}{
		{
			name: "Direct float64",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "floatVal"},
			want: 3.14,
		},
		{
			name: "Int converted to float64",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "intVal"},
			want: 100.0,
		},
		{
			name: "String parsed to float64",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "strFloat"},
			want: 2.718,
		},
		{
			name: "Invalid string defaults to 0.0",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "invalid"},
			want: 0.0,
		},
		{
			name: "Non-existent key defaults to 0.0",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "missing"},
			want: 0.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			if got := c.GetFloat(tt.args.keys); got != tt.want {
				t.Errorf("GetFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetInt(t *testing.T) {
	sampleData := map[string]interface{}{
		"intVal":   42,
		"floatVal": 50.0,
		"strInt":   "123",
		"invalid":  "xyz",
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		keys string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "Direct int",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "intVal"},
			want: 42,
		},
		{
			name: "Float converted to int",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "floatVal"},
			want: 50,
		},
		{
			name: "String parsed to int",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "strInt"},
			want: 123,
		},
		{
			name: "Invalid string defaults to 0",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "invalid"},
			want: 0,
		},
		{
			name: "Missing key defaults to 0",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "missing"},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			if got := c.GetInt(tt.args.keys); got != tt.want {
				t.Errorf("GetInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetKeys(t *testing.T) {
	sampleData := map[string]interface{}{
		"routes": map[string]interface{}{
			"live":    "/live",
			"ready":   "/ready",
			"startup": "/startup",
		},
		"scalar": 123,
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		keys string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name: "Map with multiple keys",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "routes"},
			want: []string{"live", "ready", "startup"},
		},
		{
			name: "Non-map target returns empty slice",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "scalar"},
			want: []string{},
		},
		{
			name: "Missing key returns empty slice",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "missing"},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			got := c.GetKeys(tt.args.keys)
			// Sort both to guarantee order-independent comparison
			sort.Strings(got)
			sortedWant := make([]string, len(tt.want))
			copy(sortedWant, tt.want)
			sort.Strings(sortedWant)

			if !reflect.DeepEqual(got, sortedWant) {
				t.Errorf("GetKeys() = %v, want %v", got, sortedWant)
			}
		})
	}
}

func TestConfig_GetMap(t *testing.T) {
	sampleMap := map[string]interface{}{"k1": "v1", "k2": 2}
	sampleData := map[string]interface{}{
		"section": sampleMap,
		"scalar":  "str",
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		keys string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]interface{}
	}{
		{
			name: "Existing map",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "section"},
			want: sampleMap,
		},
		{
			name: "Non-map returns empty map",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "scalar"},
			want: map[string]interface{}{},
		},
		{
			name: "Missing key returns empty map",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "missing"},
			want: map[string]interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			if got := c.GetMap(tt.args.keys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetMapBool(t *testing.T) {
	sampleData := map[string]interface{}{
		"flags": map[string]interface{}{
			"f1": true,
			"f2": "false",
		},
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		keys string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]bool
	}{
		{
			name: "Map of bools",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "flags"},
			want: map[string]bool{"f1": true, "f2": false},
		},
		{
			name: "Missing key returns empty map",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "missing"},
			want: map[string]bool{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			if got := c.GetMapBool(tt.args.keys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMapBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetMapFloat(t *testing.T) {
	sampleData := map[string]interface{}{
		"floats": map[string]interface{}{
			"a": 1.5,
			"b": 2,
		},
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		keys string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]float64
	}{
		{
			name: "Map of float64",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "floats"},
			want: map[string]float64{"a": 1.5, "b": 2.0},
		},
		{
			name: "Missing key returns empty map",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "missing"},
			want: map[string]float64{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			if got := c.GetMapFloat(tt.args.keys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMapFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetMapInt(t *testing.T) {
	sampleData := map[string]interface{}{
		"ports": map[string]interface{}{
			"http":  80,
			"https": 443,
		},
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		keys string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]int
	}{
		{
			name: "Map of ints",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "ports"},
			want: map[string]int{"http": 80, "https": 443},
		},
		{
			name: "Missing key returns empty map",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "missing"},
			want: map[string]int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			if got := c.GetMapInt(tt.args.keys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMapInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetMapString(t *testing.T) {
	sampleData := map[string]interface{}{
		"tags": map[string]interface{}{
			"env": "prod",
			"app": "backend",
		},
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		keys string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]string
	}{
		{
			name: "Map of strings",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "tags"},
			want: map[string]string{"env": "prod", "app": "backend"},
		},
		{
			name: "Missing key returns empty map",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "missing"},
			want: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			if got := c.GetMapString(tt.args.keys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMapString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetSlice(t *testing.T) {
	sampleData := map[string]interface{}{
		"items":  []interface{}{"a", 1, true},
		"scalar": 99,
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		keys string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []interface{}
	}{
		{
			name: "Existing slice",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "items"},
			want: []interface{}{"a", 1, true},
		},
		{
			name: "Non-slice returns empty slice",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "scalar"},
			want: []interface{}{},
		},
		{
			name: "Missing key returns empty slice",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "missing"},
			want: []interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			if got := c.GetSlice(tt.args.keys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetSliceBool(t *testing.T) {
	sampleData := map[string]interface{}{
		"bools": []interface{}{true, "false"},
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		keys string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []bool
	}{
		{
			name: "Slice of bools",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "bools"},
			want: []bool{true, false},
		},
		{
			name: "Missing key returns empty slice",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "missing"},
			want: []bool{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			if got := c.GetSliceBool(tt.args.keys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSliceBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetSliceFloat(t *testing.T) {
	sampleData := map[string]interface{}{
		"floats": []interface{}{1.1, 2.2, "3.3"},
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		keys string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []float64
	}{
		{
			name: "Slice of floats",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "floats"},
			want: []float64{1.1, 2.2, 3.3},
		},
		{
			name: "Missing key returns empty slice",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "missing"},
			want: []float64{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			if got := c.GetSliceFloat(tt.args.keys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSliceFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetSliceInt(t *testing.T) {
	sampleData := map[string]interface{}{
		"numbers": []interface{}{10, 20.0, "30"},
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		keys string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []int
	}{
		{
			name: "Slice of ints",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "numbers"},
			want: []int{10, 20, 30},
		},
		{
			name: "Missing key returns empty slice",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "missing"},
			want: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			if got := c.GetSliceInt(tt.args.keys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSliceInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetSliceString(t *testing.T) {
	sampleData := map[string]interface{}{
		"hosts": []interface{}{"host1", "host2"},
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		keys string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name: "Slice of strings",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "hosts"},
			want: []string{"host1", "host2"},
		},
		{
			name: "Missing key returns empty slice",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "missing"},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			if got := c.GetSliceString(tt.args.keys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSliceString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetString(t *testing.T) {
	sampleData := map[string]interface{}{
		"name": "service-name",
		"port": 8080,
		"bool": true,
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		keys string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Direct string",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "name"},
			want: "service-name",
		},
		{
			name: "Int converted to string",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "port"},
			want: "8080",
		},
		{
			name: "Bool converted to string",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "bool"},
			want: "true",
		},
		{
			name: "Missing key returns empty string",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{keys: "missing"},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			if got := c.GetString(tt.args.keys); got != tt.want {
				t.Errorf("GetString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_HasKey(t *testing.T) {
	sampleData := map[string]interface{}{
		"title": "sample",
		"obj":   map[string]interface{}{"sub": 1},
		"arr":   []interface{}{1, 2},
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		key       string
		valueType ValueType
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Has key without type check exists",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{key: "title", valueType: ""},
			want: true,
		},
		{
			name: "Has key without type check missing",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{key: "missing", valueType: ""},
			want: false,
		},
		{
			name: "Has key of type Map matching",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{key: "obj", valueType: Map},
			want: true,
		},
		{
			name: "Has key of type Map not matching scalar",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{key: "title", valueType: Map},
			want: false,
		},
		{
			name: "Has key of type Slice matching",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{key: "arr", valueType: Slice},
			want: true,
		},
		{
			name: "Has key of type Slice not matching map",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{key: "obj", valueType: Slice},
			want: false,
		},
		{
			name: "Unknown ValueType returns false",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args: args{key: "title", valueType: ValueType("unknown")},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			if got := c.HasKey(tt.args.key, tt.args.valueType); got != tt.want {
				t.Errorf("HasKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_Set(t *testing.T) {
	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		key   string
		value interface{}
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		verifyKey string
		wantValue interface{}
	}{
		{
			name: "Set and get simple value",
			fields: fields{
				data: make(map[string]interface{}),
				opts: Options{Separator: "."},
			},
			args:      args{key: "app.port", value: 3000},
			wantErr:   false,
			verifyKey: "app.port",
			wantValue: 3000,
		},
		{
			name: "Set empty key returns error",
			fields: fields{
				data: make(map[string]interface{}),
				opts: Options{Separator: "."},
			},
			args:    args{key: "", value: "val"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			if err := c.Set(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.verifyKey != "" {
				if got := c.Get(tt.verifyKey); !reflect.DeepEqual(got, tt.wantValue) {
					t.Errorf("After Set(), Get(%q) = %v, want %v", tt.verifyKey, got, tt.wantValue)
				}
			}
		})
	}
}
