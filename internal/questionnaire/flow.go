package questionnaire

import (
	"bytes"
	"strings"

	"gopkg.in/yaml.v3"
)

// A question may be written as a flow mapping on a single line:
//
//	- {id: q, type: text, prompt: "Name?"}
//
// Block keys cannot be appended to one of those, so answers are spliced inside
// the braces instead. That keeps the author's compact style rather than
// silently expanding their question over five lines.

// pt is a 1-based position in the source.
type pt struct {
	line int
	col  int
}

func (p pt) before(o pt) bool {
	return p.line < o.line || (p.line == o.line && p.col < o.col)
}

func (p pt) valid() bool { return p.line > 0 && p.col > 0 }

// flowEntry is one `key: value` pair of a flow mapping, spanning from the first
// character of the key to just past the last character of the value.
type flowEntry struct {
	key   string
	start pt
	end   pt
}

// flowPos locates a flow mapping and its entries.
type flowPos struct {
	open    pt
	close   pt
	entries []flowEntry
}

func (f flowPos) entry(key string) (flowEntry, bool) {
	for _, e := range f.entries {
		if e.key == key {
			return e, true
		}
	}
	return flowEntry{}, false
}

func (f flowPos) indexOf(key string) int {
	for i, e := range f.entries {
		if e.key == key {
			return i
		}
	}
	return -1
}

// scanFlow maps out a flow mapping starting at the given node.
func scanFlow(src []string, node *yaml.Node) (flowPos, bool) {
	open := pt{node.Line, node.Column}
	closePt, ok := matchFlowClose(src, open)
	if !ok {
		return flowPos{}, false
	}

	fp := flowPos{open: open, close: closePt}
	entries := mapEntries(node)
	for i, kv := range entries {
		start := pt{kv[0].Line, kv[0].Column}
		// An entry ends where the next one begins, less the separating comma
		// and any padding around it.
		limit := closePt
		if i+1 < len(entries) {
			limit = pt{entries[i+1][0].Line, entries[i+1][0].Column}
		}
		fp.entries = append(fp.entries, flowEntry{
			key:   kv[0].Value,
			start: start,
			end:   trimBackTo(src, start, limit),
		})
	}
	return fp, true
}

// matchFlowClose finds the brace closing the flow collection that opens at
// start, skipping over quoted scalars, comments, and nested collections.
func matchFlowClose(src []string, start pt) (pt, bool) {
	if start.line < 1 || start.line > len(src) {
		return pt{}, false
	}
	depth := 0
	for line := start.line; line <= len(src); line++ {
		text := src[line-1]
		col := 1
		if line == start.line {
			col = start.col
		}
		for ; col <= len(text); col++ {
			switch ch := text[col-1]; ch {
			case '#':
				if depth == 0 {
					return pt{}, false
				}
				col = len(text) // a comment runs to the end of the line
			case '\'', '"':
				end, ok := skipQuoted(text, col, ch)
				if !ok {
					return pt{}, false // an unterminated scalar; give up
				}
				col = end
			case '{', '[':
				depth++
			case '}', ']':
				depth--
				if depth == 0 {
					return pt{line, col}, true
				}
			}
		}
	}
	return pt{}, false
}

// skipQuoted returns the column of the closing quote of the scalar starting at
// col. Single quotes escape by doubling; double quotes by backslash.
func skipQuoted(text string, col int, quote byte) (int, bool) {
	for i := col + 1; i <= len(text); i++ {
		ch := text[i-1]
		if quote == '"' && ch == '\\' {
			i++
			continue
		}
		if ch != quote {
			continue
		}
		if quote == '\'' && i < len(text) && text[i] == '\'' {
			i++ // a doubled quote is a literal one
			continue
		}
		return i, true
	}
	return 0, false
}

// trimBackTo walks back from limit over whitespace and one separating comma, so
// the returned position is just past the previous entry's value.
func trimBackTo(src []string, floor, limit pt) pt {
	p := limit
	back := func() bool {
		if p.col > 1 {
			p.col--
			return true
		}
		if p.line > floor.line {
			p.line--
			p.col = len(src[p.line-1]) + 1
			return true
		}
		return false
	}
	at := func() byte {
		if p.line < 1 || p.line > len(src) || p.col < 2 {
			return 0
		}
		text := src[p.line-1]
		if p.col-2 >= len(text) {
			return ' '
		}
		return text[p.col-2]
	}

	for floor.before(p) {
		switch at() {
		case ' ', '\t', 0:
			if !back() {
				return p
			}
		case ',':
			if !back() {
				return p
			}
			// Trailing whitespace before the comma belongs to this entry's gap
			// too, but only one comma may be consumed.
			for floor.before(p) {
				if c := at(); c == ' ' || c == '\t' {
					if !back() {
						break
					}
					continue
				}
				break
			}
			return p
		default:
			return p
		}
	}
	return p
}

// marshalFlowValue renders a value as a single line, safe to place inside a
// flow mapping. A multi-line answer becomes a quoted string with escapes,
// because block scalars cannot appear between braces.
func marshalFlowValue(v any) string {
	node := &yaml.Node{}
	if err := node.Encode(v); err != nil {
		return quoteFlow(strings.TrimSpace(strings.Join(marshalValue(v), " ")))
	}
	forceFlow(node)

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(node); err != nil {
		enc.Close()
		return quoteFlow(strings.TrimSpace(buf.String()))
	}
	enc.Close()

	out := strings.TrimRight(buf.String(), "\n")
	if strings.Contains(out, "\n") {
		// Nothing multi-line may survive here.
		return quoteFlow(flatten(v))
	}
	return out
}

// forceFlow rewrites a node tree to flow style, and any multi-line scalar to a
// double-quoted one.
func forceFlow(n *yaml.Node) {
	switch n.Kind {
	case yaml.SequenceNode, yaml.MappingNode:
		n.Style = yaml.FlowStyle
	case yaml.ScalarNode:
		if strings.ContainsAny(n.Value, "\n") {
			n.Style = yaml.DoubleQuotedStyle
		} else if n.Style == yaml.LiteralStyle || n.Style == yaml.FoldedStyle {
			n.Style = yaml.DoubleQuotedStyle
		}
	}
	for _, c := range n.Content {
		forceFlow(c)
	}
}

func flatten(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return strings.Join(marshalValue(v), " ")
}

func quoteFlow(s string) string {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	_ = enc.Encode(&yaml.Node{Kind: yaml.ScalarNode, Style: yaml.DoubleQuotedStyle, Value: s})
	enc.Close()
	return strings.TrimRight(buf.String(), "\n")
}
