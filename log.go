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
	"fmt"
	"log"
	"net/http"
	"olympos.io/encoding/edn"
)

type Logger struct {
	Print  func(msg string) error
	Printf func(format string, a ...any) error
}

type LogBody struct {
	Logs []string `edn:"logs"`
}

func CreateLogger(url string, token string) Logger {
	logger := Logger{}

	logger.Print = func(msg string) error {
		// Print on console as well for now
		log.Print(msg)

		client := &http.Client{}

		bs, err := edn.MarshalIndent(LogBody{Logs: []string{msg}}, "", " ")
		if err != nil {
			return err
		}

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bs))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/edn")
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		return nil
	}

	logger.Printf = func(format string, a ...any) error {
		return logger.Print(fmt.Sprintf(format, a...))
	}

	return logger
}
