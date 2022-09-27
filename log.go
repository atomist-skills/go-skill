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
	"net/http"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/atomist-skills/go-skill/internal"

	log "github.com/sirupsen/logrus"
	"olympos.io/encoding/edn"
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

func createLogger(ctx context.Context, event EventIncoming) Logger {
	logger := Logger{}

	var doLog = func(msg string, level edn.Keyword) {
		bs, err := edn.MarshalPPrint(internal.LogBody{Logs: []internal.LogEntry{{
			Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05.999Z"),
			Level:     level,
			Text:      msg,
		}}}, nil)
		if err != nil {
			log.Panicf("Failed to marshal log message: %s", err)
		}

		client := &http.Client{}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, event.Urls.Logs, bytes.NewBuffer(bs))
		req.Header.Set("Authorization", "Bearer "+event.Token)
		req.Header.Set("Content-Type", "application/edn")
		if err != nil {
			log.Warnf("Failed to send log message: %s", err)
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Warnf("Failed to execute log http request: %s", err)
		}
		if resp.StatusCode != 202 {
			log.Warnf("Error sending logs: %s\n%s", resp.Status, string(bs))
		}
		defer resp.Body.Close()
	}

	logger.Debug = func(msg string) {
		log.Debug(msg)
		doLog(msg, internal.Debug)
	}
	logger.Debugf = func(format string, a ...any) {
		log.Debugf(format, a...)
		doLog(fmt.Sprintf(format, a...), internal.Debug)
	}
	logger.Info = func(msg string) {
		log.Info(msg)
		doLog(msg, internal.Info)
	}
	logger.Infof = func(format string, a ...any) {
		log.Infof(format, a...)
		doLog(fmt.Sprintf(format, a...), internal.Info)
	}
	logger.Warn = func(msg string) {
		log.Warn(msg)
		doLog(msg, internal.Warn)
	}
	logger.Warnf = func(format string, a ...any) {
		log.Warnf(format, a...)
		doLog(fmt.Sprintf(format, a...), internal.Warn)
	}
	logger.Error = func(msg string) {
		log.Error(msg)
		doLog(msg, internal.Error)
	}
	logger.Errorf = func(format string, a ...any) {
		log.Errorf(format, a...)
		doLog(fmt.Sprintf(format, a...), internal.Error)
	}

	debugInfo(logger, event)

	return logger
}

// SanitizeEvent removes any sensitive information from the incoming payload structure
func sanitizeEvent(incoming string) string {
	re, _ := regexp.Compile(`:([a-z\.\/-]*)\s*"(.*?)"`)
	res := re.FindAllStringSubmatchIndex(incoming, -1)
	for i := range res {
		name := incoming[res[i][2]:res[i][3]]
		match, _ := regexp.MatchString("(?i)token|password|jwt|url|secret|authorization|key|cert|pass|user|address|email|pat", name)
		if match {
			value := incoming[res[i][4]:res[i][5]]
			newValue := value[0:1] + strings.Repeat("*", len(value)-2) + value[len(value)-1:]
			incoming = incoming[0:res[i][4]] + newValue + incoming[res[i][5]:]
		}
	}
	return incoming
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
