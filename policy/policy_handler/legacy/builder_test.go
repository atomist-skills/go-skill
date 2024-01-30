package legacy

import (
	"reflect"
	"testing"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/types"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"olympos.io/encoding/edn"
)

func Test_BuildLocalEvalMocks(t *testing.T) {
	type args struct {
		sb *types.SBOM
	}
	tests := []struct {
		name string
		args args
		want map[edn.Keyword]edn.RawMessage
	}{
		{
			name: "Without SBOM",
			args: args{
				sb: nil,
			},
			want: map[edn.Keyword]edn.RawMessage{},
		},
		{
			name: "With SBOM",
			args: args{
				sb: &types.SBOM{
					Source: types.Source{
						Image: &types.ImageSource{
							Config: &v1.ConfigFile{
								Config: v1.Config{
									User: "root",
								},
							},
						},
					},
					Artifacts: []types.Package{
						{
							Name: "pkg1",
						},
					},
				},
			},
			want: map[edn.Keyword]edn.RawMessage{
				"image-packages-by-digest": []byte(`{:imagePackagesByDigest{:imagePackages{:packages[{:package{:licenses nil :name"pkg1":namespace"":version"":purl"":type"":vulnerabilities nil}}]}}}`),
				"get-user":                 []byte(`{:docker.image/user"root"}`),
			},
		},
		{
			name: "SBOM without image source",
			args: args{
				sb: &types.SBOM{
					Source: types.Source{
						Image: nil,
					},
					Artifacts: []types.Package{
						{
							Name: "pkg1",
						},
					},
				},
			},
			want: map[edn.Keyword]edn.RawMessage{
				"image-packages-by-digest": []byte(`{:imagePackagesByDigest{:imagePackages{:packages[{:package{:licenses nil :name"pkg1":namespace"":version"":purl"":type"":vulnerabilities nil}}]}}}`),
			},
		},
		{
			name: "SBOM without image config file",
			args: args{
				sb: &types.SBOM{
					Source: types.Source{
						Image: &types.ImageSource{
							Config: nil,
						},
					},
					Artifacts: []types.Package{
						{
							Name: "pkg1",
						},
					},
				},
			},
			want: map[edn.Keyword]edn.RawMessage{
				"image-packages-by-digest": []byte(`{:imagePackagesByDigest{:imagePackages{:packages[{:package{:licenses nil :name"pkg1":namespace"":version"":purl"":type"":vulnerabilities nil}}]}}}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildLocalEvalMocks(tt.args.sb, skill.Logger{}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildLocalEvalMocks() = %v, want %v", got, tt.want)
			}
		})
	}
}
