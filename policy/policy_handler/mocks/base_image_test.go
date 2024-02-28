package mocks

import (
	"github.com/atomist-skills/go-skill/policy/types"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_parseFromReference(t *testing.T) {
	type args struct {
		ref string
	}
	tests := []struct {
		name string
		args args
		want *SubscriptionRepository
	}{
		{
			name: "empty",
			args: args{ref: ""},
			want: nil,
		},
		{
			name: "one part",
			args: args{ref: "nginx"},
			want: &SubscriptionRepository{
				Host:       "hub.docker.com",
				Repository: "nginx",
			},
		},
		{
			name: "two parts",
			args: args{ref: "john/nginx"},
			want: &SubscriptionRepository{
				Host:       "hub.docker.com",
				Repository: "john/nginx",
			},
		},
		{
			name: "three parts",
			args: args{ref: "docker.io/john/nginx"},
			want: &SubscriptionRepository{
				Host:       "docker.io",
				Repository: "john/nginx",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseFromReference(tt.args.ref); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFromReference() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertDistro(t *testing.T) {
	testCases := []struct {
		name     string
		arg      types.Distro
		expected *SubscriptionDistro
	}{
		{
			name:     "Returns nil when distro information is not present",
			arg:      types.Distro{},
			expected: nil,
		},
		{
			name: "Correctly converts distro to subscription form when available",
			arg: types.Distro{
				OsName:    "debian",
				OsVersion: "10",
				OsDistro:  "buster",
			},
			expected: &SubscriptionDistro{
				Name:    "debian",
				Version: "10",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := convertDistro(tc.arg)
			assert.Equal(t, tc.expected, actual, "convertDistro() = %+v, want %+v", actual, tc.expected)
		})
	}
}
