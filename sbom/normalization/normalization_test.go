package normalization

import (
	"testing"

	"github.com/atomist-skills/go-skill/policy/types"
	"github.com/stretchr/testify/assert"
)

func TestPurls(t *testing.T) {
	type test struct {
		input  string
		want   string
		distro struct {
			name    string
			version string
		}
	}

	tests := []test{
		{input: "pkg:deb/debian/curl@7.64.0-4+deb10u2?distro=debian-11", want: "pkg:deb/debian/curl@7.64.0-4+deb10u2?os_name=debian&os_version=11"},
		{input: "pkg:deb/debian/curl@7.64.0-4+deb10u2?distro=debian", want: "pkg:deb/debian/curl@7.64.0-4+deb10u2?os_name=debian&os_version=unstable"},
		{input: "pkg:deb/debian/curl@7.64.0-4+deb10u2", want: "pkg:deb/debian/curl@7.64.0-4+deb10u2"},
		{input: "pkg:golang/github.com/gofiber/template@3.1.8#django/v3", want: "pkg:golang/github.com/gofiber/template/django/v3@3.1.8"},
		{input: "pkg:golang/template@3.1.8#django/v3", want: "pkg:golang/template/django/v3@3.1.8"},
		{input: "pkg:deb/template@3.1.8#django/v3", want: "pkg:deb/template/django/v3@3.1.8?os_name=alpine&os_version=3.14", distro: struct {
			name    string
			version string
		}{name: "alpine", version: "3.14"}},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			if tc.distro.name != "" {
				got, _ := NormalizePURL(tc.input, &types.Distro{
					OsName:    tc.distro.name,
					OsVersion: tc.distro.version,
				})
				assert.Equal(t, tc.want, got)
			} else {
				got, _ := NormalizePURL(tc.input, nil)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}
