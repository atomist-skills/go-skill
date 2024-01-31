package legacy

import (
	"encoding/base64"
	"os"
	"reflect"
	"testing"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/internal/test_util"
	"github.com/atomist-skills/go-skill/policy/types"
	"github.com/secure-systems-lab/go-securesystemslib/dsse"
)

func TestMockGetInTotoAttestationsForLocalEval(t *testing.T) {
	digest := test_util.Pointer("sha256:1234")

	provBytes, err := os.ReadFile("./test_data/in-toto-max-provenance.json")
	if err != nil {
		t.Fatal(err)
	}

	spdxBytes, err := os.ReadFile("./test_data/in-toto-spdx.json")
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		sb  *types.SBOM
		log skill.Logger
	}
	tests := []struct {
		name string
		args args
		want ImageAttestationQueryResult
	}{
		{
			name: "no in-toto attestations",
			args: args{
				sb: &types.SBOM{
					Source: types.Source{
						Image: &types.ImageSource{
							Digest: *digest,
						},
					},
					Attestations: []dsse.Envelope{}, // no in-toto attestations
				},
			},
			want: ImageAttestationQueryResult{
				Digest:   digest,
				Subjects: []Subject{},
			},
		},
		{
			name: "SBOM and provenance in-toto attestations",
			args: args{
				sb: &types.SBOM{
					Source: types.Source{
						Image: &types.ImageSource{
							Digest: *digest,
						},
						Provenance: &types.Provenance{
							SourceMap: &types.SourceMap{
								Instructions: []types.InstructionSourceMap{ // this instruction indicates max-mode provenance
									{
										Instruction: "FROM_RUNTIME",
										StartLine:   1,
									},
								},
							},
						},
					},
					Attestations: []dsse.Envelope{
						{
							PayloadType: "application/vnd.in-toto+json",
							Payload:     base64.StdEncoding.EncodeToString(spdxBytes),
						},
						{
							PayloadType: "application/vnd.in-toto+json",
							Payload:     base64.StdEncoding.EncodeToString(provBytes),
						},
					},
				},
			},
			want: ImageAttestationQueryResult{
				Digest: digest,
				Subjects: []Subject{
					{
						PredicateType: test_util.Pointer(SPDXPredicateType),
					},
					{
						PredicateType: test_util.Pointer(ProvenancePredicateType),
						Predicates: []Predicate{
							{
								StartLine: test_util.Pointer(1),
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MockGetInTotoAttestationsForLocalEval(tt.args.sb, tt.args.log); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MockGetInTotoAttestationsForLocalEval() = %v, want %v", got, tt.want)
			}
		})
	}
}
