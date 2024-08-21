/*
 * Copyright © 2022 Atomist, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package skill

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"olympos.io/encoding/edn"
)

// Start initiates startup of the skills given the provided Handlers
func Start(handlers Handlers) {
	Log.Info("Starting skill...")
	http.HandleFunc("/", CreateHttpHandler(handlers))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	Log.Debugf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		Log.Fatal(err)
	}
}

func CreateHttpHandler(handlers Handlers) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		rc := r.Body
		defer rc.Close()
		_, err := io.Copy(buf, rc)
		if err != nil {
			w.WriteHeader(201)
			return
		}
		body := buf.String()

		var event EventIncoming
		err = edn.NewDecoder(strings.NewReader(body)).Decode(&event)
		if err != nil {
			w.WriteHeader(201)
			return
		}

		name := NameFromEvent(event)
		ctx := context.Background()
		logger := createLogger(ctx, event, r.Header)
		req := RequestContext{
			Event: event,
			Log:   logger,

			ctx: ctx,
		}

		defer func() {
			if err := recover(); err != nil {
				sendStatus(ctx, req, Status{
					State:  Failed,
					Reason: fmt.Sprintf("Unsuccessfully invoked handler %s/%s@%s", event.Skill.Namespace, event.Skill.Name, name),
				})
				w.WriteHeader(201)
				logger.Errorf("Unhandled error occurred: %v", err)
				return
			}
		}()

		start := time.Now()
		logger.Debugf("Skill execution started")
		if req.Event.Type != "sync-request" {
			logger.Debugf("Incoming event message: %s", sanitizeEvent(body))
		}

		defer func() {
			logger.Debugf("Closing event handler '%s'", name)
			logger.Debugf("Skill execution took %d ms", time.Now().UnixMilli()-start.UnixMilli())
		}()

		if handle, ok := handlers[name]; ok {
			logger.Debugf("Invoking event handler '%s'", name)

			err = sendStatus(ctx, req, Status{
				State: running,
			})
			if err != nil {
				Log.Panicf("Failed to send status: %s", err)
			}

			status := handle(ctx, req)

			err = sendStatus(ctx, req, status)
			if err != nil {
				Log.Panicf("Failed to send status: %s", err)
			}
			w.WriteHeader(201)

			if req.Event.Type == "sync-request" {
				r := struct {
					Result interface{} `edn:"result"`
				}{
					Result: status.SyncRequest,
				}
				b, err := edn.Marshal(r)
				if err != nil {
					Log.Panicf("Failed to edn result in sync-request: %s", err)
				}
				w.Write(b)
			}
		} else {
			err = sendStatus(ctx, req, Status{
				State:  Failed,
				Reason: fmt.Sprintf("Event handler '%s' not found", name),
			})
			w.WriteHeader(201)
		}
	}
}
