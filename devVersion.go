package halfpipe

import "github.com/blang/semver"

var DevVersion = semver.Version{
	Major: 0,
	Minor: 0,
	Patch: 0,
	Pre:   []semver.PRVersion{semver.PRVersion{VersionStr: "DEV"}},
}