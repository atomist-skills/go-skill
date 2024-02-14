package mocks

import (
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
				Repository: "library/nginx",
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
