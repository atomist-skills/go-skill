package util

import (
	"os"
	"strings"
)

// Return whether the skill runs on staging or not, using the same logic as the TypeScript code.
// ref: https://github.com/atomist-skills/skill/blob/21696f154efded41f4cb98eb0e6951ab81839b8c/lib/util.ts#L235-L240
func IsStaging() bool {
	url, found := os.LookupEnv("ATOMIST_GRAPHQL_ENDPOINT")
	if !found {
		url = "https://automation.atomist.com/graphql"
	}

	return strings.Contains(url, "-stage")
}
