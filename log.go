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
	"context"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/logging"
	"github.com/atomist-skills/go-skill/internal"
	"github.com/sirupsen/logrus"
	"olympos.io/encoding/edn"
)

var (
	Log        *logrus.Logger
	projectID  string
	instanceID string
)

func init() {
	Log = logrus.New()
	Log.SetOutput(os.Stdout)
	if v, ok := os.LookupEnv("ATOMIST_LOG_LEVEL"); ok {
		switch strings.ToLower(v) {
		case "debug":
			Log.SetLevel(logrus.DebugLevel)
		case "info":
			Log.SetLevel(logrus.InfoLevel)
		case "warn":
			Log.SetLevel(logrus.WarnLevel)
		}
	}
	Log.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		PadLevelText:     true,
		ForceColors:      runtime.GOOS != "windows",
	})

	// try to obtain the GCP project id
	if _, ok := os.LookupEnv("K_SERVICE"); ok {
		projectID, _ = metadata.ProjectID()
		instanceID, _ = metadata.InstanceID()
	}
}

type Logger struct {
	Debug  func(msg string)
	Debugf func(format string, a ...any)

	Info  func(msg string)
	Infof func(format string, a ...any)

	Warn  func(msg string)
	Warnf func(format string, a ...any)

	Error  func(msg string)
	Errorf func(format string, a ...any)

	Close func()
}

func createLogger(ctx context.Context, event EventIncoming) Logger {
	logger := Logger{}

	var gcpLogger *logging.Logger
	var client *logging.Client
	labels := make(map[string]string)
	if projectID != "" {
		client, _ = logging.NewClient(ctx, projectID)
		gcpLogger = client.Logger("skill_logging")

		labels["correlation_id"] = event.ExecutionId
		labels["name"] = NameFromEvent(event)
		labels["organization"] = event.Organization
		labels["skill_id"] = fmt.Sprintf("%s/%s@%s", event.Skill.Namespace, event.Skill.Name, event.Skill.Version)
		labels["skill_name"] = event.Skill.Name
		labels["skill_namespace"] = event.Skill.Namespace
		labels["skill_version"] = event.Skill.Version
		labels["workspace_id"] = event.WorkspaceId
		labels["instance_id"] = instanceID
	}

	var doGcpLog = func(msg string, level edn.Keyword) {
		if gcpLogger != nil {
			var severity logging.Severity
			switch level {
			case internal.Debug:
				severity = logging.Debug
			case internal.Info:
				severity = logging.Info
			case internal.Warn:
				severity = logging.Warning
			case internal.Error:
				severity = logging.Error
			}
			gcpLogger.Log(logging.Entry{
				Labels:   labels,
				Severity: severity,
				Payload:  msg,
			})
		}
	}

	logger.Debug = func(msg string) {
		Log.Debug(msg)
		doGcpLog(msg, internal.Debug)
	}
	logger.Debugf = func(format string, a ...any) {
		Log.Debugf(format, a...)
		doGcpLog(fmt.Sprintf(format, a...), internal.Debug)
	}
	logger.Info = func(msg string) {
		Log.Info(msg)
		doGcpLog(msg, internal.Info)
	}
	logger.Infof = func(format string, a ...any) {
		Log.Infof(format, a...)
		doGcpLog(fmt.Sprintf(format, a...), internal.Info)
	}
	logger.Warn = func(msg string) {
		Log.Warn(msg)
		doGcpLog(msg, internal.Warn)
	}
	logger.Warnf = func(format string, a ...any) {
		Log.Warnf(format, a...)
		doGcpLog(fmt.Sprintf(format, a...), internal.Warn)
	}
	logger.Error = func(msg string) {
		Log.Error(msg)
		doGcpLog(msg, internal.Error)
	}
	logger.Errorf = func(format string, a ...any) {
		Log.Errorf(format, a...)
		doGcpLog(fmt.Sprintf(format, a...), internal.Error)
	}
	logger.Close = func() {
		if client != nil {
			_ = client.Close()
		}
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
			var newValue string

			if len(value) < 2 {
				newValue = "*"
			} else {
				newValue = value[0:1] + strings.Repeat("*", len(value)-2) + value[len(value)-1:]
			}

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
			logger.Debugf("Starting %s/%s:%s '%s' (%s) %s:%s %s", event.Skill.Namespace, event.Skill.Name, event.Skill.Version, NameFromEvent(event), revision, skillDep.Path, skillDep.Version, goVersion)
		}
	}
}
