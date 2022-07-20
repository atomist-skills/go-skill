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
)

type StatusBody struct {
	Status Status `edn:"status,omitempty"`
}

func SendStatus(ctx RequestContext, status Status) error {
	client := &http.Client{}

	bs, err := edn.MarshalIndent(StatusBody{
		Status: status,
	}, "", " ")
	if err != nil {
		return err
	}

	ctx.Log.Printf("Sending status: %s", string(bs))
	req, err := http.NewRequest(http.MethodPatch, ctx.Event.Urls.Execution, bytes.NewBuffer(bs))
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
		log.Printf("Error sending logs: %s", resp.Status)
	}

	defer resp.Body.Close()

	return nil
}
