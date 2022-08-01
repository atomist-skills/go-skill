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
	"runtime/debug"
	"time"

	"olympos.io/encoding/edn"
)

const (
	Debug edn.Keyword = "debug"
	Info              = "info"
	Warn              = "warn"
	Error             = "error"
)

type Logger struct {
	Debug  func(msg string)
	Debugf func(format string, a ...any)

	Info  func(msg string)
	Infof func(format string, a ...any)

	Warn  func(msg string)
	Warnf func(format string, a ...any)

	Error  func(msg string)
	Errorf func(format string, a ...any)
}

type LogEntry struct {
	Timestamp string      `edn:"timestamp"`
	Level     edn.Keyword `edn:"level"`
	Text      string      `edn:"text"`
}

type LogBody struct {
	Logs []LogEntry `edn:"logs"`
}

func createLogger(ctx context.Context, event EventIncoming) Logger {
	logger := Logger{}

	var doLog = func(msg string, level edn.Keyword) {
		// Print on console as well for now
		log.Print(msg)

		bs, err := edn.MarshalIndent(LogBody{Logs: []LogEntry{{
			Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05.999Z"),
			Level:     level,
			Text:      msg,
		}}}, "", " ")
		if err != nil {
			log.Panicf("Failed to marshal log message: %s", err)
		}

		client := &http.Client{}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, event.Urls.Logs, bytes.NewBuffer(bs))
		req.Header.Set("Authorization", "Bearer "+event.Token)
		req.Header.Set("Content-Type", "application/edn")
		if err != nil {
			log.Printf("Failed to send log message: %s", err)
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Failed to execute log http request: %s", err)
		}
		if resp.StatusCode != 202 {
			log.Printf("Error sending logs: %s\n%s", resp.Status, string(bs))
		}
		defer resp.Body.Close()
	}

	logger.Debug = func(msg string) {
		doLog(msg, Debug)
	}
	logger.Debugf = func(format string, a ...any) {
		doLog(fmt.Sprintf(format, a...), Debug)
	}
	logger.Info = func(msg string) {
		doLog(msg, Info)
	}
	logger.Infof = func(format string, a ...any) {
		doLog(fmt.Sprintf(format, a...), Info)
	}
	logger.Warn = func(msg string) {
		doLog(msg, Warn)
	}
	logger.Warnf = func(format string, a ...any) {
		doLog(fmt.Sprintf(format, a...), Warn)
	}
	logger.Error = func(msg string) {
		doLog(msg, Error)
	}
	logger.Errorf = func(format string, a ...any) {
		doLog(fmt.Sprintf(format, a...), Error)
	}

	debugInfo(logger, event)

	return logger
}

func debugInfo(logger Logger, event EventIncoming) {
	if bi, ok := debug.ReadBuildInfo(); ok {
		goVersion := bi.GoVersion

		var skillDep *debug.Module
		for _, v := range bi.Deps {
			if v.Path == "github.com/atomist-skills/go-skill" {
				skillDep = v
			}
		}
		var revision string
		for _, v := range bi.Settings {
			if v.Key == "vcs.revision" {
				revision = v.Value[0:7]
			}
		}
		if skillDep != nil && revision != "" {
			logger.Debugf("Starting %s/%s:%s '%s' (%s) %s:%s %s", event.Skill.Namespace, event.Skill.Name, event.Skill.Version, nameFromEvent(event), revision, skillDep.Path, skillDep.Version, goVersion)
		}
	}
}
