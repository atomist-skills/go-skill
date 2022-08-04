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
	"github.com/atomist-skills/go-skill"
	"reflect"
	"testing"
)

func TestGetParameterValueForString(t *testing.T) {
	cfg := skill.Configuration{
		Name: "default",
		Parameters: []skill.ParameterValue{
			{
				Name:  "foo",
				Value: "bar",
			},
			{
				Name:  "bar",
				Value: "foo",
			},
		},
	}

	value, _ := GetParameterValue[string]("foo", cfg)
	if reflect.TypeOf(value).Kind() != reflect.String {
		t.Errorf("Wrong type")
	}
	if value != "bar" {
		t.Errorf("Wrong value")
	}
}

func TestGetParameterValueForBoolean(t *testing.T) {
	cfg := skill.Configuration{
		Name: "default",
		Parameters: []skill.ParameterValue{
			{
				Name:  "foo",
				Value: "bar",
			},
			{
				Name:  "bar",
				Value: true,
			},
		},
	}

	value, _ := GetParameterValue[bool]("bar", cfg)
	if reflect.TypeOf(value).Kind() != reflect.Bool {
		t.Errorf("Wrong type")
	}
	if value != true {
		t.Errorf("Wrong value")
	}
}

func TestGetParameterValueErrorHandling(t *testing.T) {
	cfg := skill.Configuration{
		Name: "default",
		Parameters: []skill.ParameterValue{
			{
				Name:  "foo",
				Value: "bar",
			},
			{
				Name:  "bar",
				Value: "true",
			},
		},
	}

	_, err := GetParameterValue[bool]("bar", cfg)
	if err == nil {
		t.Errorf("Expected type conversion to fail")
	}
}
