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
	"bytes"
	"log"
	"net/http"
	"olympos.io/encoding/edn"
	"reflect"
)

type Transact func(entities interface{}) error

type MessageSender struct {
	Transact Transact
}

type TransactBody struct {
	Transactions []interface{} `edn:"transactions"`
}

func CreateMessageSender(ctx EventContext) MessageSender {
	messageSender := MessageSender{}

	messageSender.Transact = func(entities interface{}) error {
		var entityArray []interface{}
		rt := reflect.TypeOf(entities)
		switch rt.Kind() {
		case reflect.Array:
		case reflect.Slice:
			entityArray = entities.([]interface{})
		default:
			entityArray = []any{entities}
		}

		bs, err := edn.MarshalIndent(TransactBody{Transactions: entityArray}, "", " ")
		if err != nil {
			return err
		}

		client := &http.Client{}

		ctx.Log.Printf("Transacting entities: %s", string(bs))

		req, err := http.NewRequest(http.MethodPost, ctx.Event.Urls.Transactions, bytes.NewBuffer(bs))
		req.Header.Set("Authorization", "Bearer "+ctx.Event.Token)
		req.Header.Set("Content-Type", "application/edn")
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != 202 {
			log.Printf("Error transacting entities: %s", resp.Status)
		}
		defer resp.Body.Close()

		return nil
	}

	return messageSender
}
