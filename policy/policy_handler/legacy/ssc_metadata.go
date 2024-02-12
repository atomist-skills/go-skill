package legacy

import (
	"encoding/base64"
	"encoding/json"
	"slices"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/types"
	intoto "github.com/in-toto/in-toto-golang/in_toto"
)

const (
	GetInTotoAttestationsQueryName = "get-intoto-attestations"
)

const (
	SPDXPredicateType       = "https://spdx.dev/Document"
	ProvenancePredicateType = "https://slsa.dev/provenance/v0.2"
)

var (
	allowedPredicateTypes = []string{SPDXPredicateType, ProvenancePredicateType}
)

type ImageAttestationQueryResult struct {
	Digest   *string   `edn:"docker.image/digest"`
	Subjects []Subject `edn:"intoto.attestation/_subject"`
}

type Subject struct {
	PredicateType *string     `edn:"intoto.attestation/predicate-type"`
	Predicates    []Predicate `edn:"intoto.predicate/_attestation"`
}

type Predicate struct {
	StartLine *int `edn:"slsa.provenance.from/start-line"` // if field is present then provenance is max-mode
}

// https://github.com/in-toto/attestation/blob/main/spec/README.md
// https://github.com/secure-systems-lab/dsse/blob/master/envelope.md
func MockGetInTotoAttestationsForLocalEval(sb *types.SBOM, log skill.Logger) ImageAttestationQueryResult {
	subjects := []Subject{}

	// The envelope is the outtermost layer of the attestation
	for _, env := range sb.Attestations {
		if env.PayloadType != "application/vnd.in-toto+json" {
			log.Warnf("Ignoring non-in-toto attestation, payload type: %s", env.PayloadType)
			continue
		}

		// The envelope's payload is a base64-encoded JSON string
		payload, err := base64.StdEncoding.DecodeString(env.Payload)
		if err != nil {
			log.Errorf("Failed to base64-decode in-toto attestation payload: %v", err)
			continue
		}

		statement, err := unmarshalInTotoStatement(payload)
		if err != nil {
			log.Errorf("Failed to unmarshal in-toto statement %s: %v", string(payload), err)
			continue
		}

		if !slices.Contains(allowedPredicateTypes, statement.PredicateType) {
			log.Warnf("Skipping in-toto statement due to unknown predicate type: %s. Allowed predicated types: %+v", statement.PredicateType, allowedPredicateTypes)
			continue
		}

		subject := Subject{
			PredicateType: &statement.PredicateType,
		}

		if statement.PredicateType == ProvenancePredicateType && sb.Source.Provenance != nil && sb.Source.Provenance.SourceMap != nil {
			for _, i := range sb.Source.Provenance.SourceMap.Instructions {
				if i.Instruction == "FROM_RUNTIME" {
					log.Infof("Found max-mode provenance instruction: %+v", i)
					subject.Predicates = []Predicate{{StartLine: &i.StartLine}}
					break
				}
			}
		}

		subjects = append(subjects, subject)
	}

	log.Infof("Subjects: %+v", subjects)

	return ImageAttestationQueryResult{
		Digest:   &sb.Source.Image.Digest,
		Subjects: subjects,
	}
}

type intotoStatement struct {
	intoto.StatementHeader
	Predicate json.RawMessage `json:"predicate"`
}

func unmarshalInTotoStatement(content []byte) (*intotoStatement, error) {
	var stmt intotoStatement
	if err := json.Unmarshal(content, &stmt); err != nil {
		return nil, err
	}
	return &stmt, nil
}
