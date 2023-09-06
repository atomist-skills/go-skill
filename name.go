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

// NameFromEvent extracts the name of either a subscription or
// webhook from the incoming payload
func NameFromEvent(event EventIncoming) string {
	switch event.Type {
	case "subscription":
		return event.Context.Subscription.Name
	case "webhook":
		name := event.Context.Webhook.Name
		for _, v := range event.Context.Webhook.Request.Tags {
			if v.Name == "parameter-name" {
				name = v.Value.(string)
			}
		}
		return name
	case "query-result":
		return event.Context.AsyncQueryResult.Name
	case "event":
		return event.Context.Event.Name
	case "sync-request":
		return event.Context.SyncRequest.Name
	}
	return ""
}
