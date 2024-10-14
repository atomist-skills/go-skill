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
	"context"
	"net/http"
	"os"
	"time"

	"github.com/atomist-skills/go-skill/internal"

	"olympos.io/encoding/edn"
)

func NewCompletedStatus(reason string) Status {
	return Status{
		State:  Completed,
		Reason: reason,
	}
}

func NewFailedStatus(reason string) Status {
	return Status{
		State:  Failed,
		Reason: reason,
	}
}

func NewRetryableStatus(reason string) Status {
	return Status{
		State:  retryable,
		Reason: reason,
	}
}

func NewRunningStatus(reason string) Status {
	return Status{
		State:  running,
		Reason: reason,
	}
}

func SendStatus(ctx context.Context, req RequestContext, status Status) error {
	// Don't send the status when evaluating policies locally
	if os.Getenv("SCOUT_LOCAL_POLICY_EVALUATION") == "true" {
		return nil
	}
	bs, err := edn.MarshalPPrint(internal.StatusBody{
		Status: status,
	}, nil)
	if err != nil {
		return err
	}

	req.Log.Debugf("Sending status: %s", string(bs))
	client := http.DefaultClient
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, req.Event.Urls.Execution, bytes.NewBuffer(bs))
	httpReq.Header.Set("Authorization", "Bearer "+req.Event.Token)
	httpReq.Header.Set("Content-Type", "application/edn")
	if err != nil {
		return err
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		time.Sleep(time.Millisecond * 100)
		resp, err = client.Do(httpReq)
		if err != nil {
			return err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 202 {
		Log.Warnf("Error sending status: %s", resp.Status)
	}

	return nil
}
