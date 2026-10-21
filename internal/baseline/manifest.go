package baseline

import (
	"bytes"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/bbak/proviso/internal/lint"
)

// ModelVersion is the parameter model_version: the version of the Proviso
// Model this binary reads and writes.
const ModelVersion = 0

var modelLine = fmt.Sprintf("model %d", ModelVersion)

// CheckModel reports whether the first line of a manifest names the model
// version this binary reads. It reads nothing else of the manifest.
func CheckModel(data []byte) error {
	first, _, _ := bytes.Cut(data, []byte("\n"))
	if string(first) == modelLine {
		return nil
	}
	if v, ok := strings.CutPrefix(string(first), "model "); ok && v != "" {
		return fmt.Errorf("the spec tree needs model version %s, as the first line of %s says, and this proviso reads model version %d", v, "manifest.sum", ModelVersion)
	}
	return fmt.Errorf("the first line of manifest.sum names no model version, so the version the spec tree needs is unknown; this proviso reads model version %d", ModelVersion)
}

// Write returns the manifest of the elements: the model line, then one
// line per part, sorted byte by byte.
func Write(elements []lint.Element) []byte {
	var lines []string
	for _, e := range elements {
		for _, p := range Parts(e) {
			lines = append(lines, e.ID+"\t"+p.Key+"\t"+p.FP)
		}
	}
	slices.Sort(lines)
	var b bytes.Buffer
	b.WriteString(modelLine + "\n")
	for _, l := range lines {
		b.WriteString(l + "\n")
	}
	return b.Bytes()
}

// Baseline is what a manifest records: the fingerprint of each part, by
// element identifier and part.
type Baseline map[string]map[string]string

var (
	hash       = `h1:[0-9a-f]{64}`
	manifestID = regexp.MustCompile(`^[a-z][a-z0-9-]*(\.[a-z][a-z0-9-]*)*$`)
	partKey    = regexp.MustCompile(`^(heading|body|field [^\t\n]+|section [^\t\n]+|step (Given|When|Provided|Then|Otherwise)( #[1-9][0-9]*)?|row ` + hash + `( #[1-9][0-9]*)?)$`)
	fpValue    = regexp.MustCompile(`^` + hash + `$`)
)

// Parse reads a manifest whose first line lint has accepted. It returns
// what it records, or a message for each line after the first that is not
// as manifest_file describes.
func Parse(data []byte) (Baseline, []string) {
	b := Baseline{}
	var bad []string
	text := string(data)
	if !strings.HasSuffix(text, "\n") {
		// The last line lacks its line feed; it is named below.
		text += "\x00"
	}
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	for i, line := range lines[1:] {
		n := i + 2
		f := strings.Split(line, "\t")
		switch {
		case strings.HasSuffix(line, "\x00"):
			bad = append(bad, fmt.Sprintf("line %d, %q, does not end with a line feed", n, strings.TrimSuffix(line, "\x00")))
		case len(f) != 3 || !manifestID.MatchString(f[0]) || !partKey.MatchString(f[1]) || !fpValue.MatchString(f[2]):
			bad = append(bad, fmt.Sprintf("line %d, %q, is not an identifier, a part and a fingerprint separated by tabs", n, line))
		case b[f[0]][f[1]] != "":
			bad = append(bad, fmt.Sprintf("line %d, %q, records a part that an earlier line records already", n, line))
		default:
			if b[f[0]] == nil {
				b[f[0]] = map[string]string{}
			}
			b[f[0]][f[1]] = f[2]
		}
	}
	return b, bad
}
