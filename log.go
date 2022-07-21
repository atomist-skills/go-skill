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
	"fmt"
	"log"
	"net/http"
	"olympos.io/encoding/edn"
	"runtime/debug"
	"time"
)

const (
	Debug edn.Keyword = "debug"
	Info              = "info"
	Warn              = "warn"
	Error             = "error"
)

type Logger struct {
	Print  func(msg string) error
	Printf func(format string, a ...any) error
}

type LogEntry struct {
	Timestamp string      `edn:"timestamp"`
	Level     edn.Keyword `edn:"level"`
	Text      string      `edn:"text"`
}

type LogBody struct {
	Logs []LogEntry `edn:"logs"`
}

func createLogger(ctx context.Context, url string, token string) Logger {
	logger := Logger{}

	logger.Print = func(msg string) error {
		// Print on console as well for now
		log.Print(msg)

		bs, err := edn.MarshalIndent(LogBody{Logs: []LogEntry{{
			Timestamp: time.Now().Format(time.RFC3339),
			Level:     Info,
			Text:      msg,
		}}}, "", " ")
		if err != nil {
			return err
		}

		client := &http.Client{}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bs))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/edn")
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != 202 {
			log.Printf("Error sending logs: %s\n%s", resp.Status, string(bs))
		}

		defer resp.Body.Close()

		return nil
	}

	logger.Printf = func(format string, a ...any) error {
		return logger.Print(fmt.Sprintf(format, a...))
	}

	debugInfo(logger)

	return logger
}

func debugInfo(logger Logger) {
	if bi, ok := debug.ReadBuildInfo(); ok {
		goVersion := bi.GoVersion
		path := bi.Main.Path
		version := bi.Main.Version

		var skillDep *debug.Module
		for _, v := range bi.Deps {
			if v.Path == "github.com/atomist-skills/go-skill" {
				skillDep = v
			}
		}
		var revision debug.BuildSetting
		for _, v := range bi.Settings {
			if v.Key == "vcs.revision" {
				revision = v
			}
		}
		logger.Printf("Starting http listener %s:%s (%s) %s:%s %s", path, version, revision.Value[0:7], skillDep.Path, skillDep.Version, goVersion)
	}
}
