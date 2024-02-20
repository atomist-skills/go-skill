package storage

import (
	"context"
	"net/http"

	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/atomist-skills/go-skill/util"

	"cloud.google.com/go/storage"
	"github.com/atomist-skills/go-skill"
	"google.golang.org/api/googleapi"
	"olympos.io/encoding/edn"
)

func getBucketName() string {
	if util.IsStaging() {
		return "atm-policy-evaluation-results-staging"
	}

	return "atm-policy-evaluation-results"
}

type GcsStorage struct {
	client      *storage.Client
	bucketName  string
	environment string
}

func NewGcsStorage(ctx context.Context) (*GcsStorage, error) {
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &GcsStorage{
		client:     storageClient,
		bucketName: getBucketName(),
	}, nil
}

func (gcs *GcsStorage) Store(ctx context.Context, results []goals.GoalEvaluationQueryResult, storageId, environment string, log skill.Logger) error {
	log.Infof("Storing %d results", len(results))

	content, err := edn.Marshal(results)
	if err != nil {
		return err
	}
	log.Infof("Content to store: %s", string(content))

	environmentBucketName := gcs.bucketName

	if gcs.environment != "" {
		environmentBucketName = gcs.environment + "-" + gcs.environment
	}

	log.Infof("Storing results in bucket %s", environmentBucketName)

	bucket := gcs.client.Bucket(environmentBucketName)
	storageObject := bucket.Object(storageId)

	w := storageObject.If(storage.Conditions{DoesNotExist: true}).NewWriter(ctx)

	_, err = w.Write(content)
	if err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		switch e := err.(type) {
		case *googleapi.Error:
			// ignore if object already exists
			if e.Code != http.StatusPreconditionFailed {
				return err
			}
		default:
			return err
		}
	}

	return nil
}
