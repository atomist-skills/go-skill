package policy_handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/atomist-skills/go-skill/policy/data"
	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/atomist-skills/go-skill/policy/types"
	"github.com/atomist-skills/go-skill/util"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	intoto "github.com/in-toto/in-toto-golang/in_toto"
	"github.com/secure-systems-lab/go-securesystemslib/dsse"
	"olympos.io/encoding/edn"

	"github.com/atomist-skills/go-skill"
)

func withSubscription() Opt {
	return func(h *EventHandler) {
		h.evalInputProviders = append(h.evalInputProviders, getSubscriptionData)
	}
}

func getSubscriptionData(_ context.Context, req skill.RequestContext) (*goals.EvaluationMetadata, skill.Configuration, *types.SBOM, error) {
	if req.Event.Context.Subscription.Name == "" {
		return nil, skill.Configuration{}, nil, nil
	}

	evalMeta := &goals.EvaluationMetadata{
		SubscriptionResult: req.Event.Context.Subscription.GetResultInMapForm(),
		SubscriptionTx:     req.Event.Context.Subscription.Metadata.Tx,
	}

	sbom, err := createSbomFromSubscriptionResult(evalMeta.SubscriptionResult)
	if err != nil {
		return nil, skill.Configuration{}, nil, fmt.Errorf("failed to create SBOM from subscription result: %w", err)
	}

	return evalMeta, req.Event.Context.Subscription.Configuration, &sbom, nil
}

func WithSubscriptionDataSource(queryIndexes map[string]int) Opt {
	return func(h *EventHandler) {
		h.dataSourceProviders = append(h.dataSourceProviders, buildSubscriptionDataSource(queryIndexes))
	}
}

func buildSubscriptionDataSource(queryIndexes map[string]int) dataSourceProvider {
	return func(ctx context.Context, req skill.RequestContext, evalMeta goals.EvaluationMetadata) ([]data.DataSource, error) {
		return []data.DataSource{
			data.NewSubscriptionDataSource(queryIndexes, evalMeta.SubscriptionResult),
		}, nil
	}
}

func createSbomFromSubscriptionResult(subscriptionResult []map[edn.Keyword]edn.RawMessage) (types.SBOM, error) {
	imageEdn, ok := subscriptionResult[0][edn.Keyword("image")]

	if !ok {
		return types.SBOM{}, fmt.Errorf("image not found in subscription result")
	}

	image := util.Decode[goals.ImageSubscriptionQueryResult](imageEdn)

	attestations := []dsse.Envelope{}

	var sourceMap *types.SourceMap

	if image.Attestations != nil {
		for _, attestation := range image.Attestations {
			if attestation.PredicateType == nil {
				continue
			}

			intotoStatement := intotoStatement{
				StatementHeader: intoto.StatementHeader{
					PredicateType: *attestation.PredicateType,
				},
			}

			payloadBytes, _ := json.Marshal(intotoStatement)

			payload := base64.StdEncoding.EncodeToString(payloadBytes)

			env := dsse.Envelope{
				PayloadType: "application/vnd.in-toto+json",
				Payload:     payload,
			}

			for _, predicate := range attestation.Predicates {
				if predicate.StartLine != nil {
					sourceMap = &types.SourceMap{
						Instructions: []types.InstructionSourceMap{
							{
								Instruction: "FROM_RUNTIME",
								StartLine:   *predicate.StartLine,
							},
						},
					}
				}
			}

			attestations = append(attestations, env)
		}
	}

	sbom := types.SBOM{
		Source: types.Source{
			Image: &types.ImageSource{
				Digest: image.ImageDigest,

				Config: &v1.ConfigFile{
					Config: v1.Config{
						User: image.User,
					},
				},
			},
		},
		Attestations: attestations,
	}

	if image.ImagePlatforms != nil && len(image.ImagePlatforms) > 0 {
		sbom.Source.Image.Platform = types.Platform{
			Architecture: image.ImagePlatforms[0].Architecture,
			Os:           image.ImagePlatforms[0].Os,
		}
	}

	if image.FromRepo != nil && image.FromReference != nil {
		sbom.Source.Provenance = &types.Provenance{
			BaseImage: &types.ProvenanceBaseImage{
				Digest: image.FromReference.Digest,
				Tag:    image.FromTag,
				Name:   fmt.Sprintf("%s/%s", image.FromRepo.Host, image.FromRepo.Repository),
			},
			SourceMap: sourceMap,
		}
	}

	return sbom, nil
}
