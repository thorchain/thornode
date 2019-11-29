package thorchain

import "github.com/blang/semver"

func getVersion(sversion semver.Version, prefix dbPrefix) semver.Version {
	switch prefix {
	case prefixNodeAccount:
		return getNodeAccountVersion(sversion)
	default:
		return semver.MustParse("0.1.0") // default
	}
}

func getNodeAccountVersion(sversion semver.Version) semver.Version {
	return semver.MustParse("0.1.0") // default
}
