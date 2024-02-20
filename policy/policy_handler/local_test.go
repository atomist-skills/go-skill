package policy_handler

import (
	"os"
	"testing"

	"github.com/atomist-skills/go-skill"
	"olympos.io/encoding/edn"
)

func Test_parseMetadata_NullAttestations(t *testing.T) {
	req, err := createSyncReqFromFile("./test_data/sync_req_attest_null.edn")
	if err != nil {
		t.Fatal(err)
	}

	_, got, err := parseMetadata(*req)
	if err != nil {
		t.Fatalf("parseMetadata() error = %v, want nil", err)
		return
	}

	if got.Attestations != nil {
		t.Fatalf("parseMetadata() got.Attestations = %+v, want nil", got.Attestations)
	}
}

func Test_parseMetadata_NoAttestations(t *testing.T) {
	req, err := createSyncReqFromFile("./test_data/sync_req_attest_empty.edn")
	if err != nil {
		t.Fatal(err)
	}

	_, got, err := parseMetadata(*req)
	if err != nil {
		t.Fatalf("parseMetadata() error = %v, want nil", err)
		return
	}

	if got.Attestations == nil || len(got.Attestations) != 0 {
		t.Fatalf("parseMetadata() got.Attestations = %+v, want empty slice", got.Attestations)
	}
}

func Test_parseMetadata_Attestations(t *testing.T) {
	req, err := createSyncReqFromFile("./test_data/sync_req_attest.edn")
	if err != nil {
		t.Fatal(err)
	}

	_, got, err := parseMetadata(*req)
	if err != nil {
		t.Fatalf("parseMetadata() error = %v, want nil", err)
		return
	}

	if len(got.Attestations) != 2 {
		t.Fatalf("parseMetadata() got.Attestations = %+v, want 2", got.Attestations)
	}
}

// createSyncReqFromFile creates a skill.RequestContext from a file.
// The file represents the sync-request payload which contains the base64-encoded and gzipped SBOM from a local evaluation.
func createSyncReqFromFile(filename string) (*skill.RequestContext, error) {
	f, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var syncReq skill.EventContextSyncRequest
	if err := edn.Unmarshal(f, &syncReq); err != nil {
		return nil, err
	}

	return &skill.RequestContext{
		Event: skill.EventIncoming{
			Context: skill.EventContext{
				SyncRequest: syncReq,
			},
		},
	}, nil
}
