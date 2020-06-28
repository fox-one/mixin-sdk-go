package mixin

import (
	"net/url"
	"testing"
)

func Test_trimURLHost(t *testing.T) {
	type args struct {
		u *url.URL
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "url with path",
			args: args{u: parseURL("https://api.mixin.one/assets")},
			want: "/assets",
		},
		{
			name: "url without host",
			args: args{u: parseURL("/assets")},
			want: "/assets",
		},
		{
			name: "url without path",
			args: args{parseURL("https://api.mixin.one")},
			want: "/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trimURLHost(tt.args.u); got != tt.want {
				t.Errorf("trimURLHost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func parseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}

	return u
}
