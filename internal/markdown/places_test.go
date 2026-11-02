package markdown

import (
	"fmt"
	"slices"
	"testing"
)

func TestPlaces(t *testing.T) {
	cases := []struct {
		src  string
		want []string // kind:line:value, each checked against the source bytes
	}{
		{"## <a id=\"a.b\"></a> Rule: B\n", []string{"0:1:a.b"}},
		{"## <a class=\"c\" id='a.b'></a> B\n", []string{"0:1:a.b"}},
		{"# T\n\nSee [x](../r.md#a.b) and [y](#c.d \"t\").\n", []string{"1:3:a.b", "1:3:c.d"}},
		{"# T\n\nSee [#a.b](r.md#a.b).\n", []string{"1:3:a.b"}},
		{"# T\n\nSee `[x](r.md#a.b)` and <https://x.org/#a.b>.\n", []string{"1:3:a.b"}},
		{"# T\n\n```\n[x](r.md#a.b)\n```\n", nil},
		{"# T\n\n[x][r]\n\n[r]: r.md#a.b\n", []string{"3:5:a.b"}},
		{"# T\n\n[r]: <r.md#a.b> \"see #a.b\"\n[s]:\n  ../s.md#c.d\n\n- [t]: #e.f\n", []string{"3:3:a.b", "3:5:c.d", "3:7:e.f"}},
		{"# T\n\n```\n[r]: r.md#a.b\n```\n\n`[r]: r.md#a.b`\n", nil},
		{"# T\n\n![x](r.md#a.b)\n", nil},
		{"# T\n\n| a |\n| --- |\n| [x](r.md#a.b) |\n", []string{"1:5:a.b"}},
		{"# T\n\n- [a.b] Should it? (ask: X)\n- [c.d](x.md) is a link\n- no tag\n", []string{"2:3:a.b"}},
		{"\xef\xbb\xbf## <a id=\"a.b\"></a> B\n", []string{"0:1:a.b"}},
	}
	for _, c := range cases {
		var got []string
		for _, p := range Places([]byte(c.src)) {
			if c.src[p.Start:p.End] != p.Value {
				t.Errorf("%q: %+v does not stand at its bytes", c.src, p)
			}
			got = append(got, fmt.Sprintf("%d:%d:%s", p.Kind, p.Line, p.Value))
		}
		if !slices.Equal(got, c.want) {
			t.Errorf("%q: %v, want %v", c.src, got, c.want)
		}
	}
}
