package utils

import (
	"net/url"
	"path"
)

func JoinURLPath(base, u *url.URL) (joinedPath, rawJoinedPath string) {
	joinedPath = path.Join(base.Path, u.Path)

	if u.RawPath != "" {
		rawJoinedPath = path.Join(base.Path, u.RawPath)
	} else {
		rawJoinedPath = joinedPath
	}

	if base.Path == "" && len(joinedPath) > 0 && joinedPath[0] != '/' {
		joinedPath = "/" + joinedPath
	}

	if base.Path == "" && len(rawJoinedPath) > 0 && rawJoinedPath[0] != '/' {
		rawJoinedPath = "/" + rawJoinedPath
	}

	return joinedPath, rawJoinedPath
}
