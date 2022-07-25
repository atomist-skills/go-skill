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
	"log"
	"net/http"
	"olympos.io/encoding/edn"
	"os"
)

// Start initiates startup of the skills given the provided Handlers
func Start(handlers Handlers) {
	log.Print("Starting skill...")
	http.HandleFunc("/", createHttpHandler(handlers))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func createHttpHandler(handlers Handlers) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var event EventIncoming
		err := edn.NewDecoder(r.Body).Decode(&event)
		if err != nil {
			w.WriteHeader(201)
			return
		}

		var name string
		if event.Context.Subscription.Name != "" {
			name = event.Context.Subscription.Name
		} else if event.Context.Webhook.Name != "" {
			name = event.Context.Webhook.Name
		}

		ctx := context.Background()
		logger := createLogger(ctx, event.Urls.Logs, event.Token)

		if handle, ok := handlers[name]; ok {
			logger.Infof("Invoking event handler '%s'", name)

			req := RequestContext{
				Event:   event,
				Log:     logger,
			}

			messageSender := createMessageSender(ctx, req)
			req.Transact = messageSender.Transact
			req.TransactOrdered = messageSender.TransactOrdered

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

			err = sendStatus(ctx, req, Status{
				State: Running,
			})
			if err != nil {
				log.Panicf("Failed to send status: %s", err)
			}
			status := handle(ctx, req)
			err = sendStatus(ctx, req, status)
			if err != nil {
				log.Panicf("Failed to send status: %s", err)
			}
			w.WriteHeader(201)

		} else {
			logger.Errorf("Event handler '%s' not found", name)
			w.WriteHeader(201)
		}
	}
}
