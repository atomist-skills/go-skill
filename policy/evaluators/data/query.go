package data

import (
	"context"
	"fmt"

	"github.com/atomist-skills/go-skill/policy/graphql"
	"github.com/atomist-skills/go-skill/policy/query"

	"github.com/atomist-skills/go-skill"
	"olympos.io/encoding/edn"
)

type QueryDataSource struct {
	client                       graphql.GraphqlClient
	asyncClient                  *query.AsyncQueryClient
	hasAsyncResult               bool
	pkgWithVulnerabilityOverride map[string][]Package
	metadataPackages             map[string][]MetadataPackage
	imageDetailsOverride         map[string]*graphql.ImageDetailsByDigest
}

type Option func(s *QueryDataSource) error

type MetadataPackage struct {
	Licenses  []string `edn:"licenses,omitempty"` // only needed for the license policy evaluation
	Name      string   `edn:"name"`
	Namespace string   `edn:"namespace"`
	Version   string   `edn:"version"`
	Purl      string   `edn:"purl"`
	Type      string   `edn:"type"`
}

func NewQueryDataSource(ctx context.Context, req skill.RequestContext, opt ...Option) (DataSource, error) {
	c, err := graphql.NewGraphqlSkillClient(ctx, req)
	if err != nil {
		return nil, err
	}

	ds := QueryDataSource{
		client: c,
	}
	for _, option := range opt {
		err = option(&ds)
		if err != nil {
			return nil, err
		}
	}

	return ds, nil
}

func WithAsyncClient(asyncClient query.AsyncQueryClient) Option {
	return func(s *QueryDataSource) error {
		s.asyncClient = &asyncClient
		return nil
	}
}

func WithFixedPackageList(packageList map[string][]MetadataPackage) Option {
	return func(s *QueryDataSource) error {
		s.metadataPackages = packageList
		return nil
	}
}

func WithAsyncQueryResult(name string, result map[edn.Keyword]edn.RawMessage) Option {
	return func(s *QueryDataSource) error {
		switch name {
		case graphql.ImagePackagesAsyncQueryName:
			packagesResponse, err := graphql.GetImagePackagesByDigestAsyncCallback(result)
			if err != nil {
				return err
			}

			if packagesResponse != nil {
				packages, err := convertGraphqlToPackages(*packagesResponse)
				if err != nil {
					return err
				}

				s.pkgWithVulnerabilityOverride = map[string][]Package{
					packagesResponse.Digest: packages,
				}
			}
		case graphql.ImageDetailsAsyncQueryName:
			detailsResponse, err := graphql.GetImageDetailsByDigestAsyncCallback(result)
			if err != nil {
				return err
			}

			if detailsResponse != nil {
				s.imageDetailsOverride = map[string]*graphql.ImageDetailsByDigest{
					detailsResponse.Digest: detailsResponse,
				}
			}
		}

		s.hasAsyncResult = true
		return nil
	}
}

func (s QueryDataSource) canMakeAsyncRequest() bool {
	return s.asyncClient != nil && !s.hasAsyncResult
}

func (s QueryDataSource) GetPackages(ctx context.Context, digest string) (*GetPackagesResult, error) {
	if s.pkgWithVulnerabilityOverride != nil {
		return &GetPackagesResult{
			AsyncQueryMade: false,
			Result:         s.pkgWithVulnerabilityOverride[digest],
		}, nil
	}

	if s.metadataPackages != nil {
		metadataPackages := s.metadataPackages[digest]
		purls := getPurlsFromPackages(metadataPackages)
		vulnerabilitiesByPackage, err := s.client.GetVulnerabilitiesByPackage(ctx, purls)
		if err != nil {
			return nil, err
		}

		packages, err := convertMetadataPackagesToPackages(metadataPackages, vulnerabilitiesByPackage)
		if err != nil {
			return nil, err
		}

		return &GetPackagesResult{
			AsyncQueryMade: false,
			Result:         packages,
		}, nil
	}

	if s.canMakeAsyncRequest() {
		err := s.client.GetImagePackagesByDigestAsync(ctx, digest, *s.asyncClient)
		if err != nil {
			return nil, err
		}

		return &GetPackagesResult{
			AsyncQueryMade: true,
			Result:         nil,
		}, nil
	}

	packagesResponse, err := s.client.GetImagePackagesByDigest(ctx, digest)
	if err != nil {
		return nil, err
	}

	var packages []Package
	if packagesResponse != nil {
		packages, err = convertGraphqlToPackages(*packagesResponse)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("nil result received from imagePackagesByDigest query")
	}

	return &GetPackagesResult{
		AsyncQueryMade: false,
		Result:         packages,
	}, nil
}

func (s QueryDataSource) GetImageDetailsByDigest(ctx context.Context, digest string, platform query.ImagePlatform) (*GetImageDetailsByDigestResult, error) {
	if s.imageDetailsOverride != nil {
		return &GetImageDetailsByDigestResult{
			AsyncQueryMade: false,
			Result:         s.imageDetailsOverride[digest],
		}, nil
	}

	if s.canMakeAsyncRequest() {
		err := s.client.GetImageDetailsByDigestAsync(ctx, digest, platform, *s.asyncClient)
		if err != nil {
			return nil, err
		}

		return &GetImageDetailsByDigestResult{
			AsyncQueryMade: true,
			Result:         nil,
		}, nil
	}

	detailsResponse, err := s.client.GetImageDetailsByDigest(ctx, digest, platform)
	if err != nil {
		return nil, err
	}

	return &GetImageDetailsByDigestResult{
		AsyncQueryMade: false,
		Result:         detailsResponse,
	}, nil
}
