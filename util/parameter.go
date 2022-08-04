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

import (
	"errors"
	"github.com/atomist-skills/go-skill"
)

// GetParameterValue locates a configuration parameter value by name and
// converts the value to type T
func GetParameterValue[T interface{}](name string, cfg skill.Configuration) (T, error) {
	var value T
	var err error
	for _, v := range cfg.Parameters {
		if v.Name == name {
			v, ok := v.Value.(T)
			if ok {
				value = v
			} else {
				err = errors.New("failed to convert type")
			}
		}
	}
	return value, err
}
