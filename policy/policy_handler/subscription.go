package policy_handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/atomist-skills/go-skill/policy/storage"
	"github.com/atomist-skills/go-skill/policy/types"
	"github.com/atomist-skills/go-skill/util"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	intoto "github.com/in-toto/in-toto-golang/in_toto"
	"github.com/secure-systems-lab/go-securesystemslib/dsse"
	"olympos.io/encoding/edn"
)

func withSubscription() Opt {
	return func(h *EventHandler) {
		h.evalInputProviders = append(h.evalInputProviders, getSubscriptionData)
	}
}

func getSubscriptionData(ctx context.Context, req skill.RequestContext) (*goals.EvaluationMetadata, skill.Configuration, *types.SBOM, error) {
	if req.Event.Context.Subscription.Name == "" {
		return nil, skill.Configuration{}, nil, nil
	}

	evalMeta := &goals.EvaluationMetadata{
		SubscriptionResult: req.Event.Context.Subscription.GetResultInMapForm(),
		SubscriptionTx:     req.Event.Context.Subscription.Metadata.Tx,
		SubscriptionBasisT: req.Event.Context.Subscription.Metadata.AfterBasisT,
	}

	sb, err := createSBOMFromManifest(ctx, evalMeta.SubscriptionResult)
	if err == nil {
		return evalMeta, req.Event.Context.Subscription.Configuration, sb, nil
	}

	sb, err = createSBOMFromSubscriptionResult(req, evalMeta.SubscriptionResult)
	if err != nil {
		return nil, skill.Configuration{}, nil, fmt.Errorf("failed to create SBOM from subscription result: %w", err)
	}

	return evalMeta, req.Event.Context.Subscription.Configuration, sb, nil
}

func createSBOMFromManifest(ctx context.Context, subscriptionResult []map[edn.Keyword]edn.RawMessage) (*types.SBOM, error) {
	imageEdn, ok := subscriptionResult[0][edn.Keyword("image")]

	if !ok {
		return nil, fmt.Errorf("image not found in subscription result")
	}

	image := util.Decode[goals.ImageSubscriptionQueryResult](imageEdn)

	ref := image.ImageRepo.Repository
	if image.ImageRepo.Host != "hub.docker.com" {
		ref = image.ImageRepo.Host + "/" + ref
	}
	digest := image.ImageDigest

	sst := storage.NewSBOMStore(ctx)
	if sb, ok := sst.Read(ref, digest); ok {
		return sb, nil
	} else {
		return nil, fmt.Errorf("sbom not found in storage")
	}
}

func createSBOMFromSubscriptionResult(req skill.RequestContext, subscriptionResult []map[edn.Keyword]edn.RawMessage) (*types.SBOM, error) {
	imageEdn, ok := subscriptionResult[0][edn.Keyword("image")]

	if !ok {
		return nil, fmt.Errorf("image not found in subscription result")
	}

	image := util.Decode[goals.ImageSubscriptionQueryResult](imageEdn)

	attestations := []dsse.Envelope{}

	var provenanceMode *string

	if image.Attestations != nil {
		for _, attestation := range image.Attestations {
			if attestation.PredicateType == nil {
				req.Log.Debug("skipping attestation without predicate type")
				continue
			}

			intotoStatement := intotoStatement{
				StatementHeader: intoto.StatementHeader{
					PredicateType: *attestation.PredicateType,
				},
			}

			req.Log.Debugf("found attestation with predicate type %s", *attestation.PredicateType)

			payloadBytes, _ := json.Marshal(intotoStatement)

			payload := base64.StdEncoding.EncodeToString(payloadBytes)

			env := dsse.Envelope{
				PayloadType: "application/vnd.in-toto+json",
				Payload:     payload,
			}

			attestations = append(attestations, env)

			for _, predicate := range attestation.Predicates {
				if predicate.ProvenanceMode != nil {
					var mode string

					switch predicate.ProvenanceMode.Ident {
					case edn.Keyword("buildkit.provenance.mode/MAX"):
						mode = types.BuildKitMaxMode
					case edn.Keyword("buildkit.provenance.mode/MIN"):
						mode = types.BuildKitMinMode
					}

					provenanceMode = &mode
				}
			}
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
		req.Log.Debugf("found image platform: %s/%s", image.ImagePlatforms[0].Architecture, image.ImagePlatforms[0].Os)
		sbom.Source.Image.Platform = types.Platform{
			Architecture: image.ImagePlatforms[0].Architecture,
			Os:           image.ImagePlatforms[0].Os,
			Variant:      image.ImagePlatforms[0].Variant,
		}
	}

	if provenanceMode != nil {
		req.Log.Debugf("found provenance with mode %s", *provenanceMode)
		sbom.Source.Provenance = &types.Provenance{
			Mode: *provenanceMode,
		}

		if image.FromRepo != nil && image.FromReference != nil {
			req.Log.Debugf("found provenance data for base image: %s/%s:%s", image.FromRepo.Host, image.FromRepo.Repository, image.FromTag)
			sbom.Source.Provenance.BaseImage = &types.ProvenanceBaseImage{
				Digest: image.FromReference.Digest,
				Tag:    image.FromTag,
				Name:   fmt.Sprintf("%s/%s", image.FromRepo.Host, image.FromRepo.Repository),
			}
		}
	}

	return &sbom, nil
}
