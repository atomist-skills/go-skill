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

func Start(handlers Handlers) {
	log.Print("Starting server...")
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
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Temp logging
		log.Print(r.Body)

		var event EventIncoming[any]
		err := edn.NewDecoder(r.Body).Decode(&event)
		if err != nil {
			w.WriteHeader(201)
			return
		}

		name := event.Context.Subscription.Name

		logger := CreateLogger(event.Urls.Logs, event.Token)

		if handle, ok := handlers[name]; ok {
			logger.Logf("Invoking event handler '%s'", name)

			eventContext := EventContext[any]{
				Event:   event,
				Log:     logger,
				Context: ctx,
			}

			messageSender, err := CreateMessageSender(eventContext)
			if err != nil {
				logger.Logf("Error occurred creating message sender: %v", err)
				w.WriteHeader(201)
				return
			}
			eventContext.Transact = messageSender.Transact

			defer func() {
				if err := recover(); err != nil {
					SendStatus(eventContext, Status{
						State:  Failed,
						Reason: fmt.Sprintf("Unsuccessfully invoked handler %s/%s@%s", event.Skill.Namespace, event.Skill.Name, name),
					})
					w.WriteHeader(201)
					logger.Logf("Unhandled error occurred: %v", err)
					return
				}
			}()

			SendStatus(eventContext, Status{
				State: Running,
			})
			status := handle.(EventHandler[any])(eventContext)
			SendStatus(eventContext, status)
			w.WriteHeader(201)

		} else {
			logger.Logf("Event handler '%s' not found", name)
			w.WriteHeader(201)
		}
	}
}
