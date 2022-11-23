package main

import (
	"fmt"
	"runtime/debug"
)

const unknown = "unknown"

var (
	version = unknown
	commit  = unknown
	date    = unknown
)

func Commit() string {
	if commit == unknown {
		if val, ok := getBuildInfo("vcs.revision"); ok {
			commit = val
		}
		if val, ok := getBuildInfo("vcs.modified"); ok {
			if val == "true" {
				commit = fmt.Sprintf("%s-ditry", commit)
			}
		}
	}
	return commit
}

func Timestamp() string {
	if date == unknown {
		if val, ok := getBuildInfo("vcs.time"); ok {
			date = val
		}
	}
	return date
}

func Version() string {
	return version
}

func getBuildInfo(key string) (string, bool) {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == key {
				return setting.Value, true
			}
		}
	}

	return "", false
}
