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
	"context"
	"encoding/json"
	"fmt"
	"log"
	"olympos.io/encoding/edn"
	"os"
)

type Transact func(entities []interface{}) error

type MessageSender struct {
	Send     func(status Status) error
	Transact Transact
}

func CreateMessageSender(eventContext EventContext) (MessageSender, *pubsub.Client, error) {
	messageSender := MessageSender{}

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "atomist-skill-production")
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

		publishResult := t.Publish(ctx, &pubsub.Message{
			Data:        encodedMessage,
			OrderingKey: eventContext.CorrelationId,
		})

		log.Printf("Sending message: %s", encodedMessage)
		serverId, err := publishResult.Get(ctx)
		if err != nil {
			fmt.Println(err)
			return err
		}
		log.Printf("Sent message with '%s'", serverId)
		return nil
	}

	messageSender.Transact = func(entities []interface{}) error {
		bs, err := edn.Marshal(entities)
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

		publishResult := t.Publish(ctx, &pubsub.Message{
			Data:        encodedMessage,
			OrderingKey: eventContext.CorrelationId,
		})

		log.Printf("Transacting entities: %s", encodedMessage)
		serverId, err := publishResult.Get(ctx)
		if err != nil {
			fmt.Println(err)
			return err
		}
		log.Printf("Transacted entities with '%s'", serverId)
		return nil
	}

	return messageSender, client, nil
}
