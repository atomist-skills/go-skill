package test_util

import "github.com/atomist-skills/go-skill"

func CreateEmptyLogger() skill.Logger {
	return skill.Logger{
		Debug:  func(msg string) {},
		Debugf: func(format string, a ...any) {},
		Info:   func(msg string) {},
		Infof:  func(format string, a ...any) {},
		Warn:   func(msg string) {},
		Warnf:  func(format string, a ...any) {},
		Error:  func(msg string) {},
		Errorf: func(format string, a ...any) {},
	}
}
