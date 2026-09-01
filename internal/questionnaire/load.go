package questionnaire

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads and validates a questionnaire from disk.
func Load(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadBytes(path, data)
}

// LoadBytes parses and validates a questionnaire held in memory. path is
// recorded on the document and used in error messages, but is not read.
func LoadBytes(path string, data []byte) (*Document, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	body := documentBody(&root)
	if body == nil {
		return nil, Errors{fmt.Errorf("%s: expected a YAML mapping at the top level", path)}
	}

	src := lines(data)
	doc := &Document{
		Path:    path,
		source:  data,
		lineEnd: detectLineEnding(data),
		Status:  StatusPending,
	}
	doc.pos.totalLines = len(src)

	var errs Errors
	decodeString(body, "title", &doc.Title, &errs)
	decodeString(body, "intro", &doc.Intro, &errs)

	if _, _, v := mapGet(body, "status"); v != nil {
		doc.Status = Status(v.Value)
		if !doc.Status.Valid() {
			errs.addf("line %d: unknown status %q (expected pending, submitted, saved, or dismissed)", v.Line, v.Value)
		}
	}
	if _, _, v := mapGet(body, "submitted_at"); v != nil && v.Tag != "!!null" {
		doc.SubmittedAt = v.Value
	}

	doc.pos.status = mapKeyRange(src, body, "status", len(src))
	doc.pos.submittedAt = mapKeyRange(src, body, "submitted_at", len(src))
	doc.pos.questions = mapKeyRange(src, body, "questions", len(src))

	_, _, questionsNode := mapGet(body, "questions")
	if questionsNode == nil {
		errs.addf("%s: missing required key `questions`", path)
		return nil, errs
	}
	if questionsNode.Kind != yaml.SequenceNode {
		errs.addf("line %d: `questions` must be a sequence", questionsNode.Line)
		return nil, errs
	}
	if len(questionsNode.Content) == 0 {
		errs.addf("line %d: `questions` is empty; a questionnaire needs at least one question", questionsNode.Line)
		return nil, errs
	}

	// New top-level keys go immediately before `questions`, which keeps the
	// metadata together at the head of the file where an author expects it.
	doc.pos.insertAtLine = doc.pos.questions.startLine
	if doc.pos.insertAtLine == 0 {
		doc.pos.insertAtLine = len(src) + 1
	}

	ranges := seqItemRanges(src, questionsNode, doc.pos.questions.endLine)
	for i, item := range questionsNode.Content {
		q := decodeQuestion(src, item, i, ranges[i], &errs)
		if q != nil {
			doc.Questions = append(doc.Questions, q)
		}
	}

	validate(doc, &errs)
	if err := errs.ErrOrNil(); err != nil {
		return nil, err
	}
	return doc, nil
}

// documentBody unwraps the document node to reach the top-level mapping.
func documentBody(root *yaml.Node) *yaml.Node {
	n := root
	if n.Kind == yaml.DocumentNode {
		if len(n.Content) == 0 {
			return nil
		}
		n = n.Content[0]
	}
	if n.Kind != yaml.MappingNode {
		return nil
	}
	return n
}

func decodeString(m *yaml.Node, key string, dst *string, errs *Errors) {
	_, _, v := mapGet(m, key)
	if v == nil || v.Tag == "!!null" {
		return
	}
	if v.Kind != yaml.ScalarNode {
		errs.addf("line %d: `%s` must be a string", v.Line, key)
		return
	}
	*dst = v.Value
}

// decodeQuestion reads one question. It reports every problem it finds rather
// than returning at the first, and still returns the question when the parts it
// could read are usable, so later checks (duplicate ids, dangling show_if) can
// see it.
func decodeQuestion(src []string, item *yaml.Node, index int, span [2]int, errs *Errors) *Question {
	where := fmt.Sprintf("questions[%d] (line %d)", index, item.Line)
	if item.Kind != yaml.MappingNode {
		errs.addf("line %d: each entry of `questions` must be a mapping", item.Line)
		return nil
	}

	q := &Question{}
	q.pos = questionPos{
		startLine: span[0],
		endLine:   span[1],
		indent:    item.Column - 1,
		answer:    mapKeyRange(src, item, "answer", span[1]),
		comment:   mapKeyRange(src, item, "comment", span[1]),
	}

	decodeString(item, "id", &q.ID, errs)
	if q.ID == "" {
		errs.addf("%s: missing required key `id`", where)
	} else {
		where = fmt.Sprintf("question %q (line %d)", q.ID, item.Line)
	}

	decodeString(item, "prompt", &q.Prompt, errs)
	if q.Prompt == "" {
		errs.addf("%s: missing required key `prompt`", where)
	}
	decodeString(item, "help", &q.Help, errs)

	var typeName string
	decodeString(item, "type", &typeName, errs)
	if typeName == "" {
		errs.addf("%s: missing required key `type`", where)
	} else {
		q.Type = Type(typeName)
		if !q.Type.Valid() {
			errs.addf("%s: unknown type %q (expected one of %s)", where, typeName, typeList())
		}
	}

	if _, _, v := mapGet(item, "options"); v != nil {
		if v.Kind != yaml.SequenceNode {
			errs.addf("%s: `options` must be a sequence", where)
		} else {
			for _, o := range v.Content {
				if o.Kind != yaml.ScalarNode {
					errs.addf("%s: every entry of `options` must be a simple value", where)
					continue
				}
				q.Options = append(q.Options, o.Value)
			}
		}
	}

	q.Min = decodeFloat(item, "min", where, errs)
	q.Max = decodeFloat(item, "max", where, errs)

	if _, _, v := mapGet(item, "required"); v != nil {
		if err := v.Decode(&q.Required); err != nil {
			errs.addf("%s: `required` must be true or false", where)
		}
	}

	q.ShowIf = decodeShowIf(item, where, errs)

	if _, _, v := mapGet(item, "answer"); v != nil && v.Tag != "!!null" {
		var raw any
		if err := v.Decode(&raw); err != nil {
			errs.addf("%s: could not read the recorded `answer`: %v", where, err)
		} else {
			q.Answer = raw
		}
	}
	decodeString(item, "comment", &q.Comment, errs)

	return q
}

func decodeFloat(m *yaml.Node, key, where string, errs *Errors) *float64 {
	_, _, v := mapGet(m, key)
	if v == nil || v.Tag == "!!null" {
		return nil
	}
	var f float64
	if err := v.Decode(&f); err != nil {
		errs.addf("%s: `%s` must be a number", where, key)
		return nil
	}
	return &f
}

// decodeShowIf reads the conditional-visibility clauses, preserving the order
// they were written in so error messages and evaluation are predictable.
func decodeShowIf(item *yaml.Node, where string, errs *Errors) []Condition {
	_, _, v := mapGet(item, "show_if")
	if v == nil || v.Tag == "!!null" {
		return nil
	}
	if v.Kind != yaml.MappingNode {
		errs.addf("%s: `show_if` must be a mapping of question id to expected value", where)
		return nil
	}
	var out []Condition
	for _, kv := range mapEntries(v) {
		var expected any
		if err := kv[1].Decode(&expected); err != nil {
			errs.addf("%s: could not read the `show_if` value for %q", where, kv[0].Value)
			continue
		}
		out = append(out, Condition{Question: kv[0].Value, Expected: expected})
	}
	return out
}

func typeList() string {
	out := ""
	for i, t := range AllTypes {
		if i > 0 {
			out += ", "
		}
		out += string(t)
	}
	return out
}
