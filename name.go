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

// nameFromEvent extracts the name of either a subscription or
// webhook from the incoming payload
func nameFromEvent(event EventIncoming) string {
	var name string
	if event.Context.Subscription.Name != "" {
		name = event.Context.Subscription.Name
	} else if event.Context.Webhook.Name != "" {
		name = event.Context.Webhook.Name
		for _, v := range event.Context.Webhook.Request.Tags {
			if v.Name == "parameter-name" {
				name = v.Value.(string)
			}
		}
	}
	return name
}
