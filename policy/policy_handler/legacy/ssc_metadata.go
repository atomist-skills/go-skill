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
	PredicateType *string     `edn:"intoto.predicate/type"`
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

		if statement.PredicateType == ProvenancePredicateType {
			pr, err := decodeProvenance(statement.Predicate)
			if err != nil {
				log.Errorf("Failed to decode provenance predicate: %+v", err)
				continue
			}

			if step0, found := pr.Metadata.Buildkit.Source.Locations["step0"]; found && len(step0.Locations) > 0 {
				ranges := step0.Locations[0].Ranges
				if len(ranges) > 0 {
					subject.Predicates = []Predicate{{StartLine: &ranges[0].Start.Line}}
				}
			}
		}

		subjects = append(subjects, subject)
	}

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

type provenanceDocument struct {
	Invocation struct {
		ConfigSource struct {
			Uri    string `json:"uri"`
			Digest struct {
				SHA1 string `json:"sha1"`
			}
			EntryPoint string `json:"entryPoint"`
		} `json:"configSource"`
		Parameters struct {
			Args map[string]string `json:"args"`
		} `json:"parameters"`
	} `json:"invocation"`
	BuildConfig struct {
		DigestMapping map[string]string `json:"digestMapping"`
		LLBDefinition []llbDefinition   `json:"llbDefinition"`
	} `json:"buildConfig"`
	Metadata struct {
		Buildkit struct {
			VCS struct {
				Revision string `json:"revision"`
				Source   string `json:"source"`
			} `json:"vcs"`
			Source struct {
				Locations map[string]struct {
					Locations []struct {
						SourceIndex int `json:"sourceIndex"`
						Ranges      []struct {
							Start struct {
								Line int `json:"line"`
							} `json:"start"`
							End struct {
								Line int `json:"line"`
							} `json:"end"`
						} `json:"ranges"`
					} `json:"locations"`
				} `json:"locations"`
				Infos []struct {
					Path string `json:"filename"`
					Data string `json:"data"`
				} `json:"infos"`
			} `json:"source"`
			Layers map[string][][]struct {
				MediaType string `json:"mediaType"`
				Digest    string `json:"digest"`
				Size      int    `json:"size"`
			} `json:"layers"`
		} `json:"https://mobyproject.org/buildkit@v1#metadata"`
	} `json:"metadata"`
}

type llbDefinition struct {
	ID string `json:"id"`
	OP struct {
		OP struct {
			Source struct {
				Identifier string `json:"identifier"`
			} `json:"source"`
		} `json:"Op"`
		Platform struct {
			OS           string `json:"OS"`
			Architecture string `json:"Architecture"`
			Variant      string `json:"Variant"`
		}
	} `json:"op"`
}

func decodeProvenance(dt []byte) (s *provenanceDocument, err error) {
	var stmt provenanceDocument
	if err = json.Unmarshal(dt, &stmt); err != nil {
		return nil, err
	}
	return &stmt, nil
}
