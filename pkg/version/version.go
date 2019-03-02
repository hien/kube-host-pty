package version

import (
	"runtime"

	"gopkg.in/yaml.v2"
)

var (
	branch  string
	commit  string
	version string
)

type info struct {
	VersionName string `json:"version" yaml:"version"`
	Branch      string `json:"branch" yaml:"branch"`
	Commit      string `json:"commit" yaml:"commit"`
	OS          string `json:"os" yaml:"os"`
	GoVersion   string `json:"go" yaml:"go"`
}

func Info() string {
	data, _ := yaml.Marshal(&info{
		VersionName: version,
		Branch:      branch,
		Commit:      commit,
		GoVersion:   GoVersion(),
		OS:          BuildSys(),
	})
	return string(data)
}

// Branch of the git repo built at
func Branch() string { return branch }

// Commit of the git repo built with
func Commit() string { return commit }

// Version the name of git tag
func Version() string { return version }

// GoVersion the version of go used to build
func GoVersion() string { return runtime.Version() }

// BuildSys the system environment
func BuildSys() string { return runtime.GOOS + "/" + runtime.GOARCH }
