package questionnaire

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// This file writes answers back into the document they came from.
//
// It deliberately does not re-serialize the questionnaire. No Go YAML library
// round-trips with byte fidelity: comments survive, but indentation, quoting
// and line breaks are reflowed, so every save would churn a file a human reads
// and edits. Instead the parser's line positions are used to replace or insert
// only the entries this tool owns -- `answer` and `comment` on each question,
// and `status` and `submitted_at` at the top level -- leaving every other byte
// exactly as it was found.

// ownedKeys are the per-question entries the tool may create, change, or
// remove. Everything else in a question block belongs to its author.
var ownedKeys = []string{"answer", "comment"}

// edit replaces the source text from start (inclusive) to end (exclusive) with
// lines. Positions are column-precise so an answer can be spliced inside a
// one-line flow mapping as readily as appended as a block key.
//
// An edit whose start equals its end inserts without removing anything.
type edit struct {
	start pt
	end   pt
	lines []string
}

// replaceLines builds an edit covering whole lines [from, to] inclusive.
func replaceLines(from, to int, lines []string) edit {
	return edit{start: pt{from, 1}, end: pt{to + 1, 1}, lines: lines}
}

// lineBlock reports whether an edit works in whole lines rather than inside
// one. Flow-mapping edits always sit past a brace, so the two never blur.
func (e edit) lineBlock() bool { return e.start.col == 1 && e.end.col == 1 }

// insertLines builds an edit inserting whole lines before line.
func insertLines(line int, lines []string) edit {
	return edit{start: pt{line, 1}, end: pt{line, 1}, lines: lines}
}

// Render produces the document's new contents with the given status and the
// answers currently held on its questions.
//
// A dismissed questionnaire records only the status: the responder discarded
// this session's work, so nothing they typed is written, and any answers from
// an earlier session are left alone.
func (d *Document) Render(status Status, at time.Time) ([]byte, error) {
	if !status.Valid() {
		return nil, fmt.Errorf("unknown status %q", status)
	}

	src := lines(d.source)
	edits, err := d.metaEdits(status, at)
	if err != nil {
		return nil, err
	}

	if status != StatusDismissed {
		vis := d.Evaluate()
		for _, q := range d.Questions {
			qe, err := d.questionEdits(q, vis.Visible(q.ID))
			if err != nil {
				return nil, err
			}
			edits = append(edits, qe...)
		}
	}

	out, err := applyEdits(src, edits)
	if err != nil {
		return nil, err
	}

	joined := strings.Join(out, d.lineEnd)
	if len(d.source) == 0 || bytes.HasSuffix(d.source, []byte("\n")) {
		joined += d.lineEnd
	}
	return []byte(joined), nil
}

// metaEdits sets the top-level status and timestamp. When neither key is
// present they are added together, immediately before `questions`, which keeps
// the document's metadata in one block at the head of the file.
func (d *Document) metaEdits(status Status, at time.Time) ([]edit, error) {
	timestamp := ""
	if status != StatusPending {
		timestamp = at.Format(time.RFC3339)
	}

	var edits []edit
	var inserts []string

	add := func(kr keyRange, key string, value any) {
		if kr.present {
			edits = append(edits, replaceLines(kr.startLine, kr.endLine, renderEntry(key, value, kr.indent)))
			return
		}
		inserts = append(inserts, renderEntry(key, value, 0)...)
	}

	add(d.pos.status, "status", string(status))
	if timestamp == "" {
		// Keep an existing submitted_at explicit rather than deleting the key,
		// so a questionnaire reset to pending still reads as deliberately empty.
		if kr := d.pos.submittedAt; kr.present {
			line := strings.Repeat(" ", kr.indent) + "submitted_at: null"
			edits = append(edits, replaceLines(kr.startLine, kr.endLine, []string{line}))
		}
	} else {
		add(d.pos.submittedAt, "submitted_at", timestamp)
	}

	if len(inserts) > 0 {
		edits = append(edits, insertLines(d.pos.insertAtLine, inserts))
	}
	return edits, nil
}

// questionEdits writes, updates, or clears the two owned keys of one question.
//
// A question that is not visible records nothing: its answer would refer to a
// question the responder cannot see, which is a trap for whoever reads the file
// next.
func (d *Document) questionEdits(q *Question, visible bool) ([]edit, error) {
	values := map[string]any{}
	if visible {
		if q.HasAnswer() {
			normalized, err := NormalizeAnswer(q, q.Answer)
			if err != nil {
				return nil, fmt.Errorf("question %q: %w", q.ID, err)
			}
			values["answer"] = normalized
		}
		if strings.TrimSpace(q.Comment) != "" {
			values["comment"] = q.Comment
		}
	}

	if q.pos.flow != nil {
		return flowQuestionEdits(q, values), nil
	}
	return blockQuestionEdits(q, values), nil
}

// blockQuestionEdits updates the owned keys of a question written in block
// style, appending any that are new to the end of its block.
func blockQuestionEdits(q *Question, values map[string]any) []edit {
	var edits []edit
	var inserts []string

	for _, key := range ownedKeys {
		kr := q.pos.answer
		if key == "comment" {
			kr = q.pos.comment
		}
		value, wanted := values[key]

		switch {
		case wanted && kr.present:
			edits = append(edits, replaceLines(kr.startLine, kr.endLine, renderEntry(key, value, kr.indent)))
		case wanted:
			inserts = append(inserts, renderEntry(key, value, q.pos.indent)...)
		case kr.present:
			// Cleared, or the question is no longer applicable: drop the key
			// rather than leave a stale value behind.
			edits = append(edits, replaceLines(kr.startLine, kr.endLine, nil))
		}
	}

	if len(inserts) > 0 {
		edits = append(edits, insertLines(q.pos.endLine+1, inserts))
	}
	return edits
}

// flowQuestionEdits updates the owned keys of a question written as a flow
// mapping, splicing between the braces so the author's one-line style survives.
func flowQuestionEdits(q *Question, values map[string]any) []edit {
	fp := *q.pos.flow
	var edits []edit
	var additions []string

	for _, key := range ownedKeys {
		value, wanted := values[key]
		entry, present := fp.entry(key)

		switch {
		case wanted && present:
			text := key + ": " + marshalFlowValue(value)
			edits = append(edits, edit{start: entry.start, end: entry.end, lines: []string{text}})
		case wanted:
			additions = append(additions, key+": "+marshalFlowValue(value))
		case present:
			edits = append(edits, removeFlowEntry(fp, key))
		}
	}

	if len(additions) > 0 {
		at := fp.close
		text := ", " + strings.Join(additions, ", ")
		if len(fp.entries) == 0 {
			text = strings.Join(additions, ", ")
		} else {
			at = fp.entries[len(fp.entries)-1].end
		}
		edits = append(edits, edit{start: at, end: at, lines: []string{text}})
	}
	return edits
}

// removeFlowEntry deletes one entry and the comma that joins it to its
// neighbour, so the remaining mapping is still well formed.
func removeFlowEntry(fp flowPos, key string) edit {
	i := fp.indexOf(key)
	entry := fp.entries[i]
	switch {
	case len(fp.entries) == 1:
		return edit{start: entry.start, end: entry.end}
	case i+1 < len(fp.entries):
		// Take the following separator with it.
		return edit{start: entry.start, end: fp.entries[i+1].start}
	default:
		// The last entry: take the separator that precedes it.
		return edit{start: fp.entries[i-1].end, end: entry.end}
	}
}

// applyEdits splices edits into src. Edits are applied from the bottom of the
// file upwards so earlier positions stay valid as the text changes shape.
func applyEdits(src []string, edits []edit) ([]string, error) {
	if len(edits) == 0 {
		return src, nil
	}

	// Stable sort descending by start. Several inserts can share a position;
	// keeping their relative order means they come out as written.
	sort.SliceStable(edits, func(i, j int) bool { return edits[j].start.before(edits[i].start) })

	out := append([]string(nil), src...)
	lowest := pt{len(src) + 2, 1}
	for _, e := range edits {
		if e.start.line < 1 || e.start.line > len(out)+1 {
			return nil, fmt.Errorf("internal error: edit at line %d is outside the document", e.start.line)
		}
		if e.start != e.end && lowest.before(e.end) {
			return nil, fmt.Errorf("internal error: overlapping edits at line %d", e.end.line)
		}
		lowest = e.start

		head := out[:e.start.line-1]
		var middle []string
		var tail []string

		if e.lineBlock() {
			// Whole lines: the replacement stands on its own, and the line the
			// edit ends at begins the remainder untouched.
			middle = e.lines
			tail = out[min(e.end.line-1, len(out)):]
		} else {
			prefix := sliceLine(out, e.start.line, 1, e.start.col)
			suffix := lineSuffix(out, e.end.line, e.end.col)
			switch len(e.lines) {
			case 0:
				middle = []string{prefix + suffix}
			case 1:
				middle = []string{prefix + e.lines[0] + suffix}
			default:
				middle = append(middle, prefix+e.lines[0])
				middle = append(middle, e.lines[1:len(e.lines)-1]...)
				middle = append(middle, e.lines[len(e.lines)-1]+suffix)
			}
			tail = out[min(e.end.line, len(out)):]
		}

		next := make([]string, 0, len(head)+len(middle)+len(tail))
		next = append(next, head...)
		next = append(next, middle...)
		next = append(next, tail...)
		out = next
	}
	return out, nil
}

// sliceLine returns the text of a line between two 1-based columns, clamped to
// the line's length. A line past the end of the document is empty.
func sliceLine(src []string, line, from, to int) string {
	if line < 1 || line > len(src) {
		return ""
	}
	text := src[line-1]
	lo, hi := from-1, to-1
	if lo < 0 {
		lo = 0
	}
	if hi > len(text) {
		hi = len(text)
	}
	if lo > hi {
		return ""
	}
	return text[lo:hi]
}

// lineSuffix returns the text of a line from a 1-based column to its end.
func lineSuffix(src []string, line, col int) string {
	if line < 1 || line > len(src) {
		return ""
	}
	text := src[line-1]
	if col-1 >= len(text) {
		return ""
	}
	if col < 1 {
		col = 1
	}
	return text[col-1:]
}

// renderEntry formats one `key: value` entry at the given indent, choosing an
// inline scalar, a block scalar, or a nested block as the value requires.
func renderEntry(key string, value any, indent int) []string {
	pad := strings.Repeat(" ", indent)
	body := marshalValue(value)
	if len(body) == 0 {
		return []string{pad + key + ": null"}
	}

	if len(body) == 1 {
		return []string{pad + key + ": " + body[0]}
	}

	// A block scalar keeps its indicator on the key line; the encoder has
	// already indented the body two spaces relative to it.
	if strings.HasPrefix(body[0], "|") || strings.HasPrefix(body[0], ">") {
		out := []string{pad + key + ": " + body[0]}
		for _, l := range body[1:] {
			out = append(out, indentLine(pad, l))
		}
		return out
	}

	// A sequence or mapping goes beneath the key, indented one level.
	out := []string{pad + key + ":"}
	for _, l := range body {
		out = append(out, indentLine(pad+"  ", l))
	}
	return out
}

// indentLine prefixes a line unless it is empty; a blank line inside a block
// scalar must stay blank rather than become trailing whitespace.
func indentLine(pad, line string) string {
	if line == "" {
		return ""
	}
	return pad + line
}

// marshalValue renders a single value as YAML lines, with no key and no
// document markers.
func marshalValue(v any) []string {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		// Every value reaching here has been through NormalizeAnswer, so it is
		// one of string, []string, float64, bool, or int.
		enc.Close()
		return []string{fmt.Sprintf("%q", fmt.Sprint(v))}
	}
	enc.Close()
	return lines(buf.Bytes())
}

// writeAll is a seam for testing a failure partway through writing.
var writeAll = func(f *os.File, data []byte) error {
	_, err := f.Write(data)
	return err
}

// Save renders the document and replaces the file on disk.
//
// The write goes to a temporary file in the same directory and is renamed into
// place, so a failure at any point leaves the original intact rather than
// truncated. On success the document's recorded positions are refreshed, so a
// second save works from where the first left the file.
func (d *Document) Save(status Status, at time.Time) error {
	rendered, err := d.Render(status, at)
	if err != nil {
		return err
	}

	perm := os.FileMode(0o644)
	if info, err := os.Stat(d.Path); err == nil {
		perm = info.Mode().Perm()
	}
	if err := atomicWrite(d.Path, rendered, perm); err != nil {
		return err
	}

	d.Status = status
	if status != StatusPending {
		d.SubmittedAt = at.Format(time.RFC3339)
	}
	return d.refreshPositions(rendered)
}

func atomicWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".interrogate-*.tmp")
	if err != nil {
		return fmt.Errorf("could not create a temporary file next to %s: %w", path, err)
	}
	name := tmp.Name()

	discard := func(cause error) error {
		tmp.Close()
		os.Remove(name)
		return cause
	}

	if err := writeAll(tmp, data); err != nil {
		return discard(fmt.Errorf("writing %s: %w", path, err))
	}
	if err := tmp.Sync(); err != nil {
		return discard(fmt.Errorf("writing %s: %w", path, err))
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return fmt.Errorf("writing %s: %w", path, err)
	}
	if err := os.Chmod(name, perm); err != nil {
		os.Remove(name)
		return fmt.Errorf("writing %s: %w", path, err)
	}
	if err := os.Rename(name, path); err != nil {
		os.Remove(name)
		return fmt.Errorf("replacing %s: %w", path, err)
	}
	return nil
}

// refreshPositions re-reads the freshly written bytes so the line numbers held
// on the document match the file again. The questions themselves are left
// alone: callers hold pointers to them and may still be editing.
func (d *Document) refreshPositions(rendered []byte) error {
	reloaded, err := LoadBytes(d.Path, rendered)
	if err != nil {
		return fmt.Errorf("the answers were saved, but re-reading %s failed: %w", d.Path, err)
	}
	d.source = reloaded.source
	d.lineEnd = reloaded.lineEnd
	d.pos = reloaded.pos
	for _, q := range d.Questions {
		if fresh := reloaded.Question(q.ID); fresh != nil {
			q.pos = fresh.pos
		}
	}
	return nil
}
