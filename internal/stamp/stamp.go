// Package stamp reads the version stamps of installed skills and agent
// definitions, and tells which model version the Proviso version a stamp
// names reads.
package stamp

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/bbak/proviso/internal/model"
)

// Key is the key of a version stamp below metadata:.
const Key = "proviso-version"

// Add returns the agent definition or skill data with a version stamp
// naming version, in a metadata: block at the end of its frontmatter. Data
// without frontmatter is returned as it is.
func Add(data, version string) string {
	rest, ok := strings.CutPrefix(data, "---\n")
	if !ok {
		return data
	}
	i := strings.Index(rest, "\n---\n")
	if i < 0 {
		return data
	}
	i += len("---\n") + 1
	return data[:i] + "metadata:\n  " + Key + ": " + version + "\n" + data[i:]
}

var (
	metadataLine = regexp.MustCompile(`^metadata[ \t]*:$`)
	stampLine    = regexp.MustCompile(`^[ \t]*` + Key + `[ \t]*:[ \t]*(.*)$`)
)

// Read returns the versions of every version stamp in the frontmatter of
// data, as install.version-stamp defines them, in the order they stand.
func Read(data []byte) []string {
	lines := strings.Split(string(data), "\n")
	for i, l := range lines {
		// system.line-ending: a carriage return ends a line only directly
		// before a line feed, and the last line has none after it.
		if i < len(lines)-1 {
			l = strings.TrimSuffix(l, "\r")
		}
		lines[i] = strings.TrimRight(l, " \t")
	}
	if len(lines) == 0 || lines[0] != "---" {
		return nil
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			end = i
			break
		}
	}
	if end < 0 {
		return nil
	}
	var out []string
	below := false
	for _, l := range lines[1:end] {
		if below && (strings.HasPrefix(l, " ") || strings.HasPrefix(l, "\t")) {
			if m := stampLine.FindStringSubmatch(l); m != nil {
				out = append(out, m[1])
			}
			continue
		}
		below = metadataLine.MatchString(l)
	}
	return out
}

// semver matches a version of semantic versioning 2.0.0, the regular
// expression semver.org gives.
var semver = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

// Names reports whether a stamp names a Proviso version: "dev" or a
// version of semantic versioning without a leading "v".
func Names(version string) bool {
	return version == "dev" || semver.MatchString(version)
}

// version is a version of semantic versioning, read for its precedence.
type version struct {
	core [3]int
	pre  []string
}

func parse(s string) (version, bool) {
	m := semver.FindStringSubmatch(s)
	if m == nil {
		return version{}, false
	}
	var v version
	for i := range v.core {
		n, err := strconv.Atoi(m[i+1])
		if err != nil {
			return version{}, false // too large to compare
		}
		v.core[i] = n
	}
	if m[4] != "" {
		v.pre = strings.Split(m[4], ".")
	}
	return v, true
}

// isRelease reports whether s is a release version: three plain numbers,
// with neither a pre-release nor a build part.
func isRelease(s string) bool {
	v, ok := parse(s)
	return ok && v.pre == nil && !strings.Contains(s, "+")
}

// compare orders a and b by the precedence of semantic versioning.
func compare(a, b version) int {
	for i := range a.core {
		if a.core[i] != b.core[i] {
			return cmpInt(a.core[i], b.core[i])
		}
	}
	switch {
	case a.pre == nil && b.pre == nil:
		return 0
	case a.pre == nil:
		return 1
	case b.pre == nil:
		return -1
	}
	for i := 0; i < len(a.pre) && i < len(b.pre); i++ {
		x, y := a.pre[i], b.pre[i]
		nx, ex := strconv.Atoi(x)
		ny, ey := strconv.Atoi(y)
		switch {
		case x == y:
			continue
		case ex == nil && ey == nil:
			return cmpInt(nx, ny)
		case ex == nil:
			return -1
		case ey == nil:
			return 1
		case x < y:
			return -1
		default:
			return 1
		}
	}
	return cmpInt(len(a.pre), len(b.pre))
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// Model returns the model version the Proviso version stamp reads, as
// install.version.model looks it up in model_releases for a binary of the
// Proviso version own, and whether it finds one.
func Model(stamp, own string) (int, bool) {
	return modelIn(model.Releases, stamp, own)
}

func modelIn(rows []model.Release, stamp, own string) (int, bool) {
	if !isRelease(stamp) || len(rows) == 0 {
		return 0, false
	}
	s, _ := parse(stamp)
	if own != "dev" {
		o, ok := parse(own)
		if !ok || compare(s, o) > 0 {
			return 0, false
		}
	}
	found, ok := 0, false
	for _, r := range rows {
		first, _ := parse(r.First)
		if compare(first, s) > 0 {
			break
		}
		found, ok = r.Model, true
	}
	return found, ok
}
