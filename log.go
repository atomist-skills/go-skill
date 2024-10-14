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
	"net/http"
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

type CreateLogger func(ctx context.Context, labels map[string]string) *Logger

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

	if v, ok := os.LookupEnv("ATOMIST_LOG_FORMAT"); ok && v == "json" {
		Log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		Log.SetFormatter(&logrus.TextFormatter{
			DisableTimestamp: true,
			PadLevelText:     true,
			ForceColors:      runtime.GOOS != "windows",
		})
	}

	if v, ok := os.LookupEnv("ATOMIST_LOG_GCP"); !ok || v == "true" {
		// try to obtain the GCP project id
		if _, ok := os.LookupEnv("K_SERVICE"); ok {
			projectID, _ = metadata.ProjectID()
			instanceID, _ = metadata.InstanceID()
		}
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

func createCommonLabels(event EventIncoming, headers http.Header) map[string]string {
	labels := make(map[string]string)

	labels["correlation_id"] = event.ExecutionId
	labels["name"] = NameFromEvent(event)
	labels["organization"] = event.Organization
	labels["skill_id"] = fmt.Sprintf("%s/%s@%s", event.Skill.Namespace, event.Skill.Name, event.Skill.Version)
	labels["skill_name"] = event.Skill.Name
	labels["skill_namespace"] = event.Skill.Namespace
	labels["skill_version"] = event.Skill.Version
	labels["workspace_id"] = event.WorkspaceId
	labels["instance_id"] = instanceID

	labels["request_id"] = headers.Get("X-Request-ID")
	labels["forwarded_host"] = headers.Get("X-Forwarded-Host")
	labels["original_forwarded_for"] = headers.Get("X-Original-Forwarded-For")
	labels["cloud_trace_context"] = headers.Get("X-Cloud-Trace-Context")
	labels["trace_parent"] = headers.Get("traceparent")

	return labels
}

func createGcpLogger(ctx context.Context, labels map[string]string) *Logger {
	if projectID == "" {
		return nil
	}

	var gcpLogger *logging.Logger
	var client *logging.Client

	client, _ = logging.NewClient(ctx, projectID)
	gcpLogger = client.Logger("skill_logging")

	if gcpLogger != nil {
		return nil
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

	logger := Logger{
		Debug: func(msg string) {
			doGcpLog(msg, internal.Debug)
		},
		Debugf: func(format string, a ...any) {
			doGcpLog(fmt.Sprintf(format, a...), internal.Debug)
		},
		Info: func(msg string) {
			doGcpLog(msg, internal.Info)
		},
		Infof: func(format string, a ...any) {
			doGcpLog(fmt.Sprintf(format, a...), internal.Info)
		},
		Warn: func(msg string) {
			doGcpLog(msg, internal.Warn)
		},
		Warnf: func(format string, a ...any) {
			doGcpLog(fmt.Sprintf(format, a...), internal.Warn)
		},
		Error: func(msg string) {
			doGcpLog(msg, internal.Error)
		},
		Errorf: func(format string, a ...any) {
			doGcpLog(fmt.Sprintf(format, a...), internal.Error)
		},
		Close: func() {
			if client != nil {
				_ = client.Close()
			}
		},
	}

	return &logger
}

func createDefaultLogger(ctx context.Context, labels map[string]string) *Logger {
	localLabels := make(map[string]interface{})
	for k, v := range labels {
		localLabels[k] = v
	}

	logger := Logger{
		Debug: func(msg string) {
			Log.WithFields(localLabels).Debug(msg)
		},
		Debugf: func(format string, a ...any) {
			Log.WithFields(localLabels).Debugf(format, a...)
		},
		Info: func(msg string) {
			Log.WithFields(localLabels).Info(msg)
		},
		Infof: func(format string, a ...any) {
			Log.WithFields(localLabels).Infof(format, a...)
		},
		Warn: func(msg string) {
			Log.WithFields(localLabels).Warn(msg)
		},
		Warnf: func(format string, a ...any) {
			Log.WithFields(localLabels).Warnf(format, a...)
		},
		Error: func(msg string) {
			Log.WithFields(localLabels).Error(msg)
		},
		Errorf: func(format string, a ...any) {
			Log.WithFields(localLabels).Errorf(format, a...)
		},
		Close: func() {
		},
	}

	return &logger
}

func createLogger(ctx context.Context, event EventIncoming, headers http.Header, loggerCreator CreateLogger) Logger {
	labels := createCommonLabels(event, headers)

	loggerCreators := []CreateLogger{
		createGcpLogger,
	}

	if loggerCreator != nil {
		loggerCreators = append(loggerCreators, loggerCreator)
	} else {
		loggerCreators = append(loggerCreators, createDefaultLogger)
	}

	loggers := []Logger{}
	for _, creator := range loggerCreators {
		l := creator(ctx, labels)
		if l != nil {
			loggers = append(loggers, *l)
		}
	}

	logger := Logger{}
	logger.Debug = func(msg string) {
		for _, l := range loggers {
			l.Debug(msg)
		}
	}
	logger.Debugf = func(format string, a ...any) {
		a = expandFuncs(a, logrus.DebugLevel)
		for _, l := range loggers {
			l.Debugf(format, a...)
		}
	}
	logger.Info = func(msg string) {
		for _, l := range loggers {
			l.Info(msg)
		}
	}
	logger.Infof = func(format string, a ...any) {
		for _, l := range loggers {
			l.Infof(format, a...)
		}
	}
	logger.Warn = func(msg string) {
		for _, l := range loggers {
			l.Warn(msg)
		}
	}
	logger.Warnf = func(format string, a ...any) {
		for _, l := range loggers {
			l.Warnf(format, a...)
		}
	}
	logger.Error = func(msg string) {
		for _, l := range loggers {
			l.Error(msg)
		}
	}
	logger.Errorf = func(format string, a ...any) {
		for _, l := range loggers {
			l.Errorf(format, a...)
		}
	}
	logger.Close = func() {
		for _, l := range loggers {
			l.Close()
		}
	}

	debugInfo(logger, event)

	return logger
}

func expandFuncs(a []any, level logrus.Level) []any {
	for i, v := range a {
		if f, ok := v.(func() interface{}); ok && Log.Level >= level {
			a[i] = f()
		}
	}

	return a
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
