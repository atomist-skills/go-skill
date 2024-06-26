package storage

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/atomist-skills/go-skill/policy/types"
	"google.golang.org/api/googleapi"
)

const (
	bucketName = "atm-prod-stored-manifests"
)

type (
	Cache struct {
		ctx        context.Context
		client     *storage.Client
		bucketName string
		directory  string
	}
)

func (c *Cache) Read(ref, digest string) (*types.SBOM, bool) {
	bucket := c.client.Bucket(c.bucketName)
	seg := make([]string, 0)
	seg = append(seg, c.directory)
	seg = append(seg, ref)
	seg = append(seg, digest)

	storageObject := bucket.Object(strings.Join(seg, "/"))

	r, err := storageObject.NewReader(c.ctx)
	if err != nil {
		switch e := err.(type) {
		case *googleapi.Error:
			if e.Code == http.StatusNotFound {
				return nil, false
			}
		default:
			return nil, false
		}
	}
	defer r.Close()

	sb := &types.SBOM{}
	if err := json.NewDecoder(r).Decode(sb); err != nil {
		return nil, false
	}
	return sb, true
}

func NewSBOMStore(ctx context.Context) *Cache {
	gcs, _ := storage.NewClient(ctx)
	return &Cache{
		ctx:        ctx,
		client:     gcs,
		bucketName: bucketName,
		directory:  "sbom",
	}
}
