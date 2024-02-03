package policy_handler

import (
	"context"
	"testing"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/goals"
	"github.com/stretchr/testify/assert"
	"olympos.io/encoding/edn"
)

type TestQueryResultFields struct {
	SomeString     string `edn:"stringField"`
	SomeOtherField int    `edn:"intField"`
}

type TestQueryResultInner struct {
	SomeList []TestQueryResultFields `edn:"myList"`
}
type TestQueryResult struct {
	InnerResult TestQueryResultInner `edn:"inner"`
}

func Test_buildLocalDataSources_preservesQueryResultsCorrectly(t *testing.T) {
	const queryName = "test-query"
	expected := TestQueryResult{
		InnerResult: TestQueryResultInner{
			SomeList: []TestQueryResultFields{{
				SomeString:     "abc",
				SomeOtherField: 5,
			}},
		},
	}
	expectedEdn, err := edn.Marshal(expected)
	if err != nil {
		t.Fatalf("failed to marshal data: %s", err.Error())
	}

	srMetaEdn, err := edn.Marshal(SyncRequestMetadata{
		QueryResults: map[edn.Keyword]edn.RawMessage{
			queryName: expectedEdn,
		},
	})

	ds, err := buildLocalDataSources(
		context.TODO(),
		skill.RequestContext{
			Event: skill.EventIncoming{
				Context: skill.EventContext{
					SyncRequest: skill.EventContextSyncRequest{
						Name:     eventNameLocalEval,
						Metadata: srMetaEdn,
					},
				},
			},
		},
		goals.EvaluationMetadata{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(ds) == 0 {
		t.Fatal("buildLocalDataSources returned no data sources")
	}

	var actual TestQueryResult
	r, err := ds[0].Query(context.TODO(), queryName, "", map[string]interface{}{}, &actual)
	if err != nil {
		t.Fatal(err)
	}
	if r.AsyncRequestMade {
		t.Fatal("AsyncRequestMade was set to true, but expected false")
	}

	assert.Equal(t, expected, actual)
}
