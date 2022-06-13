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
	"cloud.google.com/go/pubsub"
	"encoding/json"
	"fmt"
	"olympos.io/encoding/edn"
	"os"
	"reflect"
)

type Transact func(entities interface{}) error

type MessageSender struct {
	Send     func(status Status) error
	Transact Transact
}

func CreateMessageSender(eventContext EventContext) (MessageSender, *pubsub.Client, error) {
	messageSender := MessageSender{}

	client, err := pubsub.NewClient(eventContext.Context, "atomist-skill-production")
	if err != nil {
		return messageSender, client, err
	}
	t := client.Topic(os.Getenv("ATOMIST_TOPIC"))
	t.EnableMessageOrdering = true

	messageSender.Send = func(status Status) error {
		message := StatusHandlerResponse{
			ApiVersion:    "1",
			CorrelationId: eventContext.CorrelationId,
			Team: Team{
				Id: eventContext.WorkspaceId,
			},
			Skill:  eventContext.Skill,
			Status: status,
		}

		encodedMessage, _ := json.Marshal(message)

		publishResult := t.Publish(eventContext.Context, &pubsub.Message{
			Data:        encodedMessage,
			OrderingKey: eventContext.CorrelationId,
		})

		eventContext.Log.Printf("Sending message: %s", encodedMessage)
		serverId, err := publishResult.Get(eventContext.Context)
		if err != nil {
			fmt.Println(err)
			return err
		}
		eventContext.Log.Printf("Sent message with '%s'", serverId)
		return nil
	}

	messageSender.Transact = func(entities interface{}) error {
		var entityArray any
		rt := reflect.TypeOf(entities)
		switch rt.Kind() {
		case reflect.Array:
		case reflect.Slice:
			entityArray = entities
		default:
			entityArray = []any{entities}
		}

		bs, err := edn.Marshal(entityArray)
		if err != nil {
			return err
		}

		message := TransactEntitiesResponse{
			ApiVersion:    "1",
			CorrelationId: eventContext.CorrelationId,
			Team: Team{
				Id: eventContext.WorkspaceId,
			},
			Type:     "facts_ingestion",
			Entities: string(bs),
		}
		encodedMessage, _ := json.Marshal(message)

		publishResult := t.Publish(eventContext.Context, &pubsub.Message{
			Data:        encodedMessage,
			OrderingKey: eventContext.CorrelationId,
		})

		eventContext.Log.Printf("Transacting entities: %s", encodedMessage)
		serverId, err := publishResult.Get(eventContext.Context)
		if err != nil {
			fmt.Println(err)
			return err
		}
		eventContext.Log.Printf("Transacted entities with '%s'", serverId)
		return nil
	}

	return messageSender, client, nil
}
