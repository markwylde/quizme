package questionnaire

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// keyRange locates a single `key: value` entry in the source, from the line the
// key sits on through the last line of its value.
type keyRange struct {
	present   bool
	startLine int // 1-based, inclusive
	endLine   int // 1-based, inclusive
	indent    int // spaces before the key
}

// documentPos locates the parts of the top-level mapping that may be rewritten.
type documentPos struct {
	totalLines   int
	status       keyRange
	submittedAt  keyRange
	questions    keyRange
	insertAtLine int // where a missing top-level key should be inserted
}

// questionPos locates one question's block and the two keys within it that
// belong to this tool.
type questionPos struct {
	startLine int
	endLine   int
	indent    int
	answer    keyRange
	comment   keyRange
}

// lines splits src into lines without their terminators. A trailing newline
// does not produce a final empty element, so len(lines) is the line count a
// text editor would report.
func lines(src []byte) []string {
	s := strings.ReplaceAll(string(src), "\r\n", "\n")
	s = strings.TrimSuffix(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

// detectLineEnding reports the dominant line terminator in src so rewritten
// files keep the convention they arrived with.
func detectLineEnding(src []byte) string {
	crlf := strings.Count(string(src), "\r\n")
	lf := strings.Count(string(src), "\n") - crlf
	if crlf > lf {
		return "\r\n"
	}
	return "\n"
}

// isFiller reports whether a line is blank or contains only a comment. Filler
// at the end of a block belongs to whatever comes next, not to the block, so
// it is excluded from computed ranges and survives rewriting untouched.
func isFiller(line string) bool {
	t := strings.TrimSpace(line)
	return t == "" || strings.HasPrefix(t, "#")
}

// trimFiller walks back from rawEnd to the last line that is not filler, never
// going above floor.
func trimFiller(src []string, floor, rawEnd int) int {
	end := rawEnd
	if end > len(src) {
		end = len(src)
	}
	for end > floor && isFiller(src[end-1]) {
		end--
	}
	return end
}

// mapEntries returns the key/value node pairs of a mapping node.
func mapEntries(m *yaml.Node) [][2]*yaml.Node {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	out := make([][2]*yaml.Node, 0, len(m.Content)/2)
	for i := 0; i+1 < len(m.Content); i += 2 {
		out = append(out, [2]*yaml.Node{m.Content[i], m.Content[i+1]})
	}
	return out
}

// mapGet finds a key in a mapping, returning its index among the pairs along
// with the key and value nodes.
func mapGet(m *yaml.Node, key string) (int, *yaml.Node, *yaml.Node) {
	for i, kv := range mapEntries(m) {
		if kv[0].Value == key {
			return i, kv[0], kv[1]
		}
	}
	return -1, nil, nil
}

// mapKeyRange computes the source range of one entry of a mapping.
//
// The end of a value is derived from where the next key starts rather than
// from the value node itself, because a node's own extent is unknowable for
// block scalars and folded strings. hardEnd bounds the last entry, which has
// no following key to lean on.
func mapKeyRange(src []string, m *yaml.Node, key string, hardEnd int) keyRange {
	idx, keyNode, _ := mapGet(m, key)
	if idx < 0 {
		return keyRange{}
	}
	entries := mapEntries(m)
	rawEnd := hardEnd
	if idx+1 < len(entries) {
		rawEnd = entries[idx+1][0].Line - 1
	}
	return keyRange{
		present:   true,
		startLine: keyNode.Line,
		endLine:   trimFiller(src, keyNode.Line, rawEnd),
		indent:    keyNode.Column - 1,
	}
}

// seqItemRanges computes the source range of every item in a sequence node.
func seqItemRanges(src []string, seq *yaml.Node, hardEnd int) [][2]int {
	if seq == nil || seq.Kind != yaml.SequenceNode {
		return nil
	}
	out := make([][2]int, len(seq.Content))
	for i, item := range seq.Content {
		rawEnd := hardEnd
		if i+1 < len(seq.Content) {
			rawEnd = seq.Content[i+1].Line - 1
		}
		out[i] = [2]int{item.Line, trimFiller(src, item.Line, rawEnd)}
	}
	return out
}
