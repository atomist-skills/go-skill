package skills

import (
	"testing"
)

func TestParseMultiChoiceArg(t *testing.T) {
	tests := []struct {
		name string
		arg  interface{}
		want []string
	}{
		{
			name: "Parse nil arg",
			arg:  nil,
			want: []string{},
		},
		{
			name: "Parse multiple args",
			arg:  []interface{}{"CRITICAL", "HIGH"},
			want: []string{"CRITICAL", "HIGH"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseMultiChoiceArg(tt.arg)
			if len(got) != len(tt.want) {
				t.Errorf("TestParseMultiChoiceArg() = %v, want %v", got, tt.arg)
			}

			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("TestParseMultiChoiceArg() = %v, want %v", got, tt.arg)
				}
			}
		})
	}
}

func TestParseStringArrayArg(t *testing.T) {
	tests := []struct {
		name string
		arg  interface{}
		want []string
	}{
		{
			name: "Parse nil arg",
			arg:  nil,
			want: []string{},
		},
		{
			name: "Parse multiple args",
			arg:  []interface{}{"MIT", "GPL-3.0"},
			want: []string{"MIT", "GPL-3.0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseStringArrayArg(tt.arg)
			if len(got) != len(tt.want) {
				t.Errorf("ParseStringArrayArgs() = %v, want %v", got, tt.arg)
			}

			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("ParseStringArrayArgs() = %v, want %v", got, tt.arg)
				}
			}
		})
	}
}

func TestParseIntArgs(t *testing.T) {
	tests := []struct {
		name string
		arg  interface{}
		want int64
	}{
		{
			name: "Parse nil arg",
			arg:  nil,
			want: 0,
		},
		{
			name: "Parse single arg",
			arg:  int64(30),
			want: 30,
		},
		{
			name: "Parse int32 arg",
			arg:  int32(30),
			want: 30,
		},
		{
			name: "Parse float64 arg",
			arg:  float64(30),
			want: 30,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseIntArg(tt.arg); got != tt.want {
				t.Errorf("ParseIntArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
