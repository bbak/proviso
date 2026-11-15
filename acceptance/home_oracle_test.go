package acceptance

// The oracle of lint.identifiers.home, computed from the test tree alone.

import (
	"regexp"
	"strings"
)

// wellFormed is the pattern of a [well-formed] identifier.
var wellFormed = regexp.MustCompile(`^[a-z][a-z0-9-]*(\.[a-z][a-z0-9-]*)*$`)

// capabilityIDs gives the identifier of every capability of tr, by the
// folder that holds it; "" where its capability.md is missing or holds no
// identifier.
func capabilityIDs(tr tree) map[string]string {
	out := map[string]string{}
	for p := range tr {
		m := capabilityDir.FindStringSubmatch(p)
		if m == nil {
			continue
		}
		capFile := "proviso/capabilities/" + m[1] + "/capability.md"
		out[m[1]] = ""
		for _, e := range treeElements(tree{capFile: tr[capFile]}) {
			out[m[1]] = e[0]
			break
		}
	}
	return out
}

// homeSegments gives the [home segment]s of tr (system.home-segment): system
// and the first segment, up to its first ".", of the identifier of every
// capability of tr, well-formed or not.
func homeSegments(tr tree) map[string]bool {
	out := map[string]bool{"system": true}
	for _, id := range capabilityIDs(tr) {
		if id != "" {
			first, _, _ := strings.Cut(id, ".")
			out[first] = true
		}
	}
	return out
}

// unhomed gives the identifiers of tr at which lint.identifiers.home reports
// findings.identifier-unhomed, each with its file, in the order of
// treeElements.
func unhomed(tr tree) [][2]string {
	caps := capabilityIDs(tr)
	homes := homeSegments(tr)
	systemID := ""
	for _, e := range treeElements(tree{systemPath: tr[systemPath]}) {
		systemID = e[0]
		break
	}
	var out [][2]string
	for _, e := range treeElements(tr) {
		id, p := e[0], e[1]
		if !wellFormed.MatchString(id) || (p == systemPath && id == systemID) {
			continue
		}
		if m := capabilityDir.FindStringSubmatch(p); m != nil && !wellFormed.MatchString(caps[m[1]]) {
			continue
		}
		first, _, _ := strings.Cut(id, ".")
		if !homes[first] {
			out = append(out, e)
		}
	}
	return out
}
