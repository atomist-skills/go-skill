package legacy

import "github.com/atomist-skills/go-skill/policy/types"

const (
	GetInTotoAttestationsQueryName = "get-intoto-attestations"
)

type ImageAttestationQueryResult struct {
	Digest   *string   `edn:"docker.image/digest"`
	Subjects []Subject `edn:"intoto.attestation/_subject"`
}

type Subject struct {
	PredicateType *string     `edn:"intoto.predicate/type"`
	Predicates    []Predicate `edn:"intoto.predicate/_attestation"`
}

type Predicate struct {
	StartLine *int `edn:"slsa.provenance.from/start-line"` // if field is present then provenance is max-mode
}

func MockGetInTotoAttestationsForLocalEval(sb *types.SBOM) ImageAttestationQueryResult {
	return ImageAttestationQueryResult{} // incompatible with local evaluation until SBOM includes the raw attestations
}
