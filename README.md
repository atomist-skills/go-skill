# github.com/atomist-skills/go-skill

[Go](https://go.dev) implementation of the Atomist Skill v2 contract.

## Set up your Go project

We suggest the following layout for a your skill Go project:

| File/Directory         | Description                                                    |
| ---------------------- | -------------------------------------------------------------- |                                                           
| `datalog/subscription` | Put all `.edn` files defining subscriptions into this directory |
| `datalog/schema`       | Place all `.edn` schema files into this directory              |
| `go.mod`               | Go module descriptor                                           |
| `main.go`              | Main entry point to start up the skill instance                | 
| `Dockerfile`           | Dockerfile required to build your skill runtime container      |
| `skill.yaml`           | Skill metadata                                                 |

## Skill entrypoint

The `main.go` is the main entry point into your Go application.

In the `main` method you can start the skill instance passing a mapping of subscription
or webhook names to handler functions. 

```go
package main

import "github.com/atomist-skills/go-skill"

func main() {
	skill.Start(skill.Handlers{
		"on_push": TransactCommitSignature,
		"on_commit_signature": LogCommitSignature,
	})
}
```
                                       
## Handler function

A function to handle incoming subscription or webhook events is defined as:

```go
type EventHandler func(ctx context.Context, req RequestContext) Status
```
                                                                  
Here's an example `EventHandler` implementation from the [go-sample-skill](https://github.com/atomist-skills/go-sample-skill):

```go
// LogCommitSignature handles new commit signature entities as they are transacted into
// the database and logs the signature
func LogCommitSignature(ctx context.Context, req skill.RequestContext) skill.Status {
	result := req.Event.Context.Subscription.Result[0]
	commit := skill.Decode[GitCommit](result[0])
	signature := skill.Decode[GitCommitSignature](result[1])

	req.Log.Printf("Commit %s is signed and verified by: %s ", commit.Sha, signature.Signature)

	return skill.Status{
		State:  skill.Completed,
		Reason: "Detected signed and verified commit",
	}
}
```

### `RequestContext`

The passed `RequestContext` provides access to the incoming payload using the `Event` property as well 
as properties to get access to a `Logger` instance and `transact` functions:

```go
type RequestContext struct {
	Event           EventIncoming
	Log             Logger
	Transact        Transact
	TransactOrdered TransactOrdered
}
```

### Transacting entities

Transacting new entities or facts can be done by calling `Transact` on the `RequestConext` passed the entities:

```go
type GitRepoEntity struct {
    EntityType edn.Keyword `edn:"schema/entity-type"`
    Entity     string      `edn:"schema/entity,omitempty"`
    SourceId   string      `edn:"git.repo/source-id"`
    Url        string      `edn:"git.provider/url"`
}

err := req.Transact([]any{GitRepoEntity{
		EntityType: "git/repo",
		Entity:     "$repo",
		SourceId:   commit.Repo.SourceId,
		Url:        commit.Repo.Org.Url,
	}, GitCommitEntity{
		EntityType: "git/commit",
		Entity:     "$commit",
		Sha:        commit.Sha,
		Repo:       "$repo",
		Url:        commit.Repo.Org.Url,
	}, GitCommitSignatureEntity{
		EntityType: "git.commit/signature",
		Commit:     "$commit",
		Signature:  signature,
		Status:     verified,
		Reason:     *gitCommit.Commit.Verification.Reason,
	}})
```

### Sending logs

To send logs to the skill platform to be viewed on `go.atomist.com`, the `RequestContext` provides access to a `Logger`
struct:

```go
type Logger struct {
    Debug  func(msg string)
    Debugf func(format string, a ...any)
    
    Info  func(msg string)
    Infof func(format string, a ...any)
    
    Warn  func(msg string)
    Warnf func(format string, a ...any)
    
    Error  func(msg string)
    Errorf func(format string, a ...any)
}
```

## Skill metadata

The Skill metadata is defined in `skill.yaml` in the root of the project. 

Here's an example of a minimum `skill.yaml` file:

```yaml
skill:
  apiVersion: v2
  namespace: atomist
  name: go-sample-skill
  displayName: Go Sample Skill
  description: Very basic sample skill written in Go
  author: Atomist
  license: Apache-2.0
```

## Building the runtime image

The v2 version of the skill contract is only available to Docker-based skills. The following `Dockerfile` 
is a good starting point for building a skill runtime image:

```dockerfile
   # build stage
FROM golang:1.18-alpine as build

RUN apk add --no-cache git build-base

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

RUN go test
RUN go build

# runtime stage
FROM golang:1.18-alpine

LABEL com.docker.skill.api.version="container/v2"
COPY skill.yaml /
COPY datalog /datalog

WORKDIR /skill
COPY --from=build /app/go-sample-skill .

ENTRYPOINT ["/skill/go-sample-skill"]
```

This `Dockerfile` uses a multi stage approach to build and test the Go project before 
setting up the runtime container in the 2nd stage.
