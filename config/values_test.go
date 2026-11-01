package config

import (
	"reflect"
	"testing"
)

func TestConfig_findValue(t *testing.T) {
	sampleData := map[string]interface{}{
		"server": map[string]interface{}{
			"host": "localhost",
			"port": 8080,
			"ssl": map[string]interface{}{
				"enabled": true,
			},
		},
		"scalar": "just a string",
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
		want1  bool
	}{
		{
			name: "Empty key",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args:  args{key: ""},
			want:  nil,
			want1: false,
		},
		{
			name: "Top-level existing key",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args:  args{key: "scalar"},
			want:  "just a string",
			want1: true,
		},
		{
			name: "Nested existing key",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args:  args{key: "server.port"},
			want:  8080,
			want1: true,
		},
		{
			name: "Deeply nested key",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args:  args{key: "server.ssl.enabled"},
			want:  true,
			want1: true,
		},
		{
			name: "Non-existing top-level key",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args:  args{key: "missing"},
			want:  nil,
			want1: false,
		},
		{
			name: "Non-existing nested key",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args:  args{key: "server.missing"},
			want:  nil,
			want1: false,
		},
		{
			name: "Key traversing through non-map value",
			fields: fields{
				data: sampleData,
				opts: Options{Separator: "."},
			},
			args:  args{key: "scalar.nested"},
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			got, got1 := c.findValue(tt.args.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findValue() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("findValue() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestConfig_getValue(t *testing.T) {
	data := map[string]interface{}{
		"map": map[string]interface{}{
			"foo": "bar",
		},
		"num": 42,
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
		want1  bool
	}{
		{
			name: "Existing value returns clone",
			fields: fields{
				data: data,
				opts: Options{Separator: "."},
			},
			args:  args{key: "num"},
			want:  42,
			want1: true,
		},
		{
			name: "Existing map returns cloned map",
			fields: fields{
				data: data,
				opts: Options{Separator: "."},
			},
			args:  args{key: "map"},
			want:  map[string]interface{}{"foo": "bar"},
			want1: true,
		},
		{
			name: "Non-existing key",
			fields: fields{
				data: data,
				opts: Options{Separator: "."},
			},
			args:  args{key: "notFound"},
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			got, got1 := c.getValue(tt.args.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getValue() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getValue() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestConfig_getMap(t *testing.T) {
	data := map[string]interface{}{
		"validMap": map[string]interface{}{"k": "v"},
		"notAMap":   123,
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]interface{}
		want1  bool
	}{
		{
			name: "Existing valid map",
			fields: fields{
				data: data,
				opts: Options{Separator: "."},
			},
			args:  args{key: "validMap"},
			want:  map[string]interface{}{"k": "v"},
			want1: true,
		},
		{
			name: "Non-map value",
			fields: fields{
				data: data,
				opts: Options{Separator: "."},
			},
			args:  args{key: "notAMap"},
			want:  nil,
			want1: false,
		},
		{
			name: "Key not found",
			fields: fields{
				data: data,
				opts: Options{Separator: "."},
			},
			args:  args{key: "missing"},
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			got, got1 := c.getMap(tt.args.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getMap() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getMap() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestConfig_getSlice(t *testing.T) {
	data := map[string]interface{}{
		"validSlice": []interface{}{"a", "b", "c"},
		"notASlice":  "hello",
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []interface{}
		want1  bool
	}{
		{
			name: "Existing valid slice",
			fields: fields{
				data: data,
				opts: Options{Separator: "."},
			},
			args:  args{key: "validSlice"},
			want:  []interface{}{"a", "b", "c"},
			want1: true,
		},
		{
			name: "Non-slice value",
			fields: fields{
				data: data,
				opts: Options{Separator: "."},
			},
			args:  args{key: "notASlice"},
			want:  nil,
			want1: false,
		},
		{
			name: "Key not found",
			fields: fields{
				data: data,
				opts: Options{Separator: "."},
			},
			args:  args{key: "missing"},
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			got, got1 := c.getSlice(tt.args.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSlice() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getSlice() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestConfig_set(t *testing.T) {
	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		key   string
		value interface{}
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		wantData map[string]interface{}
	}{
		{
			name: "Empty key error",
			fields: fields{
				data: make(map[string]interface{}),
				opts: Options{Separator: "."},
			},
			args:    args{key: "", value: "val"},
			wantErr: true,
		},
		{
			name: "Set top level key",
			fields: fields{
				data: make(map[string]interface{}),
				opts: Options{Separator: "."},
			},
			args:    args{key: "name", value: "test-app"},
			wantErr: false,
			wantData: map[string]interface{}{
				"name": "test-app",
			},
		},
		{
			name: "Set nested key with new hierarchy",
			fields: fields{
				data: make(map[string]interface{}),
				opts: Options{Separator: "."},
			},
			args:    args{key: "database.master.host", value: "127.0.0.1"},
			wantErr: false,
			wantData: map[string]interface{}{
				"database": map[string]interface{}{
					"master": map[string]interface{}{
						"host": "127.0.0.1",
					},
				},
			},
		},
		{
			name: "Set nested key inside existing map",
			fields: fields{
				data: map[string]interface{}{
					"database": map[string]interface{}{
						"port": 5432,
					},
				},
				opts: Options{Separator: "."},
			},
			args:    args{key: "database.host", value: "localhost"},
			wantErr: false,
			wantData: map[string]interface{}{
				"database": map[string]interface{}{
					"port": 5432,
					"host": "localhost",
				},
			},
		},
		{
			name: "Overwrite non-map intermediate node with map",
			fields: fields{
				data: map[string]interface{}{
					"database": "not-a-map",
				},
				opts: Options{Separator: "."},
			},
			args:    args{key: "database.port", value: 3306},
			wantErr: false,
			wantData: map[string]interface{}{
				"database": map[string]interface{}{
					"port": 3306,
				},
			},
		},
		{
			name: "Invalid map key during normalization",
			fields: fields{
				data: make(map[string]interface{}),
				opts: Options{Separator: "."},
			},
			args:    args{key: "invalid", value: map[int]string{1: "val"}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			err := c.set(tt.args.key, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantData != nil && !reflect.DeepEqual(c.data, tt.wantData) {
				t.Errorf("set() data = %v, want %v", c.data, tt.wantData)
			}
		})
	}
}

func Test_cloneValue(t *testing.T) {
	mOriginal := map[string]interface{}{"key": "orig"}
	sOriginal := []interface{}{"item1", "item2"}

	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "Primitive int",
			args: args{v: 123},
			want: 123,
		},
		{
			name: "Primitive string",
			args: args{v: "hello"},
			want: "hello",
		},
		{
			name: "Map clone",
			args: args{v: mOriginal},
			want: map[string]interface{}{"key": "orig"},
		},
		{
			name: "Slice clone",
			args: args{v: sOriginal},
			want: []interface{}{"item1", "item2"},
		},
		{
			name: "Nil value",
			args: args{v: nil},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cloneValue(tt.args.v)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cloneValue() = %v, want %v", got, tt.want)
			}
		})
	}

	// Verify deep isolation
	clonedMap := cloneValue(mOriginal).(map[string]interface{})
	clonedMap["key"] = "modified"
	if mOriginal["key"] == "modified" {
		t.Errorf("cloneValue() did not create a deep copy of map")
	}

	clonedSlice := cloneValue(sOriginal).([]interface{})
	clonedSlice[0] = "modified"
	if sOriginal[0] == "modified" {
		t.Errorf("cloneValue() did not create a deep copy of slice")
	}
}

func Test_normalizeValue(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name:    "Nil value",
			args:    args{v: nil},
			want:    nil,
			wantErr: false,
		},
		{
			name:    "Primitive string",
			args:    args{v: "text"},
			want:    "text",
			wantErr: false,
		},
		{
			name:    "Array conversion to slice",
			args:    args{v: [2]string{"a", "b"}},
			want:    []interface{}{"a", "b"},
			wantErr: false,
		},
		{
			name:    "Slice of ints to slice of interface{}",
			args:    args{v: []int{1, 2}},
			want:    []interface{}{1, 2},
			wantErr: false,
		},
		{
			name:    "Map with string keys",
			args:    args{v: map[string]int{"k": 10}},
			want:    map[string]interface{}{"k": 10},
			wantErr: false,
		},
		{
			name:    "Map with non-string keys errors",
			args:    args{v: map[int]string{1: "val"}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeValue(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("normalizeValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("normalizeValue() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToBoolMap(t *testing.T) {
	type args struct {
		input map[string]interface{}
	}
	tests := []struct {
		name  string
		args  args
		want  map[string]bool
		want1 bool
	}{
		{
			name: "Map with boolean conversions",
			args: args{
				input: map[string]interface{}{
					"a": true,
					"b": "true",
					"c": "FALSE",
					"d": "invalid",
				},
			},
			want: map[string]bool{
				"a": true,
				"b": true,
				"c": false,
				"d": false,
			},
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := convertToBoolMap(tt.args.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToBoolMap() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("convertToBoolMap() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_convertToBoolSlice(t *testing.T) {
	type args struct {
		input []interface{}
	}
	tests := []struct {
		name  string
		args  args
		want  []bool
		want1 bool
	}{
		{
			name: "Slice with bool conversions",
			args: args{
				input: []interface{}{true, "True", "false", 123},
			},
			want:  []bool{true, true, false, false},
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := convertToBoolSlice(tt.args.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToBoolSlice() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("convertToBoolSlice() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_convertToFloatMap(t *testing.T) {
	type args struct {
		input map[string]interface{}
	}
	tests := []struct {
		name  string
		args  args
		want  map[string]float64
		want1 bool
	}{
		{
			name: "Map with float conversions",
			args: args{
				input: map[string]interface{}{
					"int":    10,
					"float":  20.5,
					"str":    "30.25",
					"invStr": "xyz",
				},
			},
			want: map[string]float64{
				"int":    10.0,
				"float":  20.5,
				"str":    30.25,
				"invStr": 0.0,
			},
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := convertToFloatMap(tt.args.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToFloatMap() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("convertToFloatMap() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_convertToFloatSlice(t *testing.T) {
	type args struct {
		input []interface{}
	}
	tests := []struct {
		name  string
		args  args
		want  []float64
		want1 bool
	}{
		{
			name: "Slice with float conversions",
			args: args{
				input: []interface{}{5, 6.5, "7.25", "bad"},
			},
			want:  []float64{5.0, 6.5, 7.25, 0.0},
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := convertToFloatSlice(tt.args.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToFloatSlice() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("convertToFloatSlice() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_convertToIntMap(t *testing.T) {
	type args struct {
		input map[string]interface{}
	}
	tests := []struct {
		name  string
		args  args
		want  map[string]int
		want1 bool
	}{
		{
			name: "Map with int conversions",
			args: args{
				input: map[string]interface{}{
					"int":    42,
					"float":  50.0,
					"str":    "100",
					"badStr": "abc",
				},
			},
			want: map[string]int{
				"int":    42,
				"float":  50,
				"str":    100,
				"badStr": 0,
			},
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := convertToIntMap(tt.args.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToIntMap() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("convertToIntMap() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_convertToIntSlice(t *testing.T) {
	type args struct {
		input []interface{}
	}
	tests := []struct {
		name  string
		args  args
		want  []int
		want1 bool
	}{
		{
			name: "Slice with int conversions",
			args: args{
				input: []interface{}{1, 2.0, "3", "invalid"},
			},
			want:  []int{1, 2, 3, 0},
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := convertToIntSlice(tt.args.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToIntSlice() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("convertToIntSlice() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_convertToStringMap(t *testing.T) {
	type args struct {
		input map[string]interface{}
	}
	tests := []struct {
		name  string
		args  args
		want  map[string]string
		want1 bool
	}{
		{
			name: "Map with string conversions",
			args: args{
				input: map[string]interface{}{
					"str":  "plain",
					"int":  123,
					"bool": true,
				},
			},
			want: map[string]string{
				"str":  "plain",
				"int":  "123",
				"bool": "true",
			},
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := convertToStringMap(tt.args.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToStringMap() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("convertToStringMap() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_convertToStringSlice(t *testing.T) {
	type args struct {
		input []interface{}
	}
	tests := []struct {
		name  string
		args  args
		want  []string
		want1 bool
	}{
		{
			name: "Slice with string conversions",
			args: args{
				input: []interface{}{"hello", 456, false},
			},
			want:  []string{"hello", "456", "false"},
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := convertToStringSlice(tt.args.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToStringSlice() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("convertToStringSlice() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_toBool(t *testing.T) {
	type args struct {
		value interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Bool true", args{value: true}, true},
		{"Bool false", args{value: false}, false},
		{"String true", args{value: "true"}, true},
		{"String True upper", args{value: "TRUE"}, true},
		{"String false", args{value: "false"}, false},
		{"String FALSE upper", args{value: "FALSE"}, false},
		{"Invalid string", args{value: "not_a_bool"}, false},
		{"Integer", args{value: 1}, false},
		{"Nil value", args{value: nil}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toBool(tt.args.value); got != tt.want {
				t.Errorf("toBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toFloat64(t *testing.T) {
	type args struct {
		value interface{}
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{"float64", args{value: float64(12.34)}, 12.34},
		{"float32", args{value: float32(12.5)}, 12.5},
		{"int", args{value: int(10)}, 10.0},
		{"int8", args{value: int8(8)}, 8.0},
		{"int16", args{value: int16(16)}, 16.0},
		{"int32", args{value: int32(32)}, 32.0},
		{"int64", args{value: int64(64)}, 64.0},
		{"uint", args{value: uint(10)}, 10.0},
		{"uint8", args{value: uint8(8)}, 8.0},
		{"uint16", args{value: uint16(16)}, 16.0},
		{"uint32", args{value: uint32(32)}, 32.0},
		{"uint64", args{value: uint64(64)}, 64.0},
		{"valid numeric string", args{value: "123.45"}, 123.45},
		{"invalid string", args{value: "abc"}, 0.0},
		{"boolean", args{value: true}, 0.0},
		{"nil", args{value: nil}, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toFloat64(tt.args.value); got != tt.want {
				t.Errorf("toFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toInt(t *testing.T) {
	type args struct {
		value interface{}
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"int", args{value: int(10)}, 10},
		{"int8", args{value: int8(8)}, 8},
		{"int16", args{value: int16(16)}, 16},
		{"int32", args{value: int32(32)}, 32},
		{"int64", args{value: int64(64)}, 64},
		{"uint", args{value: uint(10)}, 10},
		{"uint8", args{value: uint8(8)}, 8},
		{"uint16", args{value: uint16(16)}, 16},
		{"uint32", args{value: uint32(32)}, 32},
		{"uint64", args{value: uint64(64)}, 64},
		{"float64 integer", args{value: float64(50.0)}, 50},
		{"float64 non-integer", args{value: float64(50.5)}, 0},
		{"float32 integer", args{value: float32(25.0)}, 25},
		{"float32 non-integer", args{value: float32(25.5)}, 0},
		{"valid string", args{value: "100"}, 100},
		{"invalid string", args{value: "abc"}, 0},
		{"boolean", args{value: true}, 0},
		{"nil", args{value: nil}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toInt(tt.args.value); got != tt.want {
				t.Errorf("toInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toString(t *testing.T) {
	type args struct {
		value interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"string", args{value: "hello"}, "hello"},
		{"int", args{value: 123}, "123"},
		{"int8", args{value: int8(8)}, "8"},
		{"int16", args{value: int16(16)}, "16"},
		{"int32", args{value: int32(32)}, "32"},
		{"int64", args{value: int64(64)}, "64"},
		{"uint", args{value: uint(10)}, "10"},
		{"uint8", args{value: uint8(8)}, "8"},
		{"uint16", args{value: uint16(16)}, "16"},
		{"uint32", args{value: uint32(32)}, "32"},
		{"uint64", args{value: uint64(64)}, "64"},
		{"float32", args{value: float32(1.5)}, "1.5"},
		{"float64", args{value: float64(2.5)}, "2.5"},
		{"bool true", args{value: true}, "true"},
		{"bool false", args{value: false}, "false"},
		{"unsupported struct", args{value: struct{}{}}, ""},
		{"nil", args{value: nil}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toString(tt.args.value); got != tt.want {
				t.Errorf("toString() = %v, want %v", got, tt.want)
			}
		})
	}
}
