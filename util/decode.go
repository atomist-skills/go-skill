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

package util

import "olympos.io/encoding/edn"

// Decode returns a subscription result payload as a struct of specified
// generic type P
func Decode[P interface{}](event edn.RawMessage) P {
	ednboby, _ := edn.Marshal(event)
	var decoded P
	edn.Unmarshal(ednboby, &decoded)
	return decoded
}
