package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// GetMostNarrowPath will return the most narrow match in _paths_
// of the _descendant_ (of which is exactly on or beneath a path).
//
// NOTE: Since many the _descendant_ may be decendant on many paths,
// the longest one will be returned, hence the most narrow.
//
// If no match, `false` and -1 is returned.
func GetMostNarrowPath(descendant string, paths []string) (pos int, ok bool) {

	pos = -1
	longest := -1
	for i := range paths {

		yes, err := IsSubPath(paths[i], descendant)

		if err != nil {
			panic(err)
		}

		if !yes {
			continue
		}

		if len(paths[i]) > longest {

			longest = len(paths[i])
			pos = i
			ok = true

		}

	}

	return
}

// IsSubPath checks if _subpath_ is a subpath to _path_.
//
// CAUTION: Make sure that both _path_ and _subpath_ is absolute
// path. Use `filepath.Abs(path)` (though may be a bit unsure, see docs)
// if possible.
func IsSubPath(path, subpath string) (bool, error) {
	up := ".." + string(os.PathSeparator)

	rel, err := filepath.Rel(path, subpath)

	if err != nil {
		return false, err
	}

	if !strings.HasPrefix(rel, up) && rel != ".." {
		return true, nil
	}

	return false, nil

}
