package questionnaire

import (
	"encoding/json"
	"io"
)

// Output is what a caller reads from stdout once the form closes: the outcome,
// and the answers keyed by question id.
//
// It exists so an agent can act on the result without re-reading and re-parsing
// the questionnaire file, though the file remains the authority if it does.
type Output struct {
	Path        string           `json:"path"`
	Title       string           `json:"title,omitempty"`
	Status      Status           `json:"status"`
	SubmittedAt string           `json:"submitted_at,omitempty"`
	Complete    bool             `json:"complete"`
	Answers     map[string]Entry `json:"answers"`
	Unanswered  []string         `json:"unanswered,omitempty"`
}

// Entry is one question's result. The prompt and type travel with the answer so
// the reader does not have to hold the questionnaire alongside it.
type Entry struct {
	Prompt  string `json:"prompt"`
	Type    Type   `json:"type"`
	Answer  any    `json:"answer,omitempty"`
	Comment string `json:"comment,omitempty"`
}

// Result collects the answers to report for the given outcome.
//
// Only questions the responder could actually see are included: a hidden
// question's answer was never recorded in the file, so reporting one would
// invite an agent to act on something the human never agreed to. A dismissed
// questionnaire reports no answers at all.
func (d *Document) Result(status Status) Output {
	out := Output{
		Path:        d.Path,
		Title:       d.Title,
		Status:      status,
		SubmittedAt: d.SubmittedAt,
		Answers:     map[string]Entry{},
	}
	if status == StatusDismissed {
		return out
	}

	out.Complete = d.IsComplete()
	for _, q := range d.VisibleQuestions() {
		if !q.HasAnswer() && q.Comment == "" {
			continue
		}
		entry := Entry{Prompt: q.Prompt, Type: q.Type, Comment: q.Comment}
		if q.HasAnswer() {
			entry.Answer = q.Answer
		}
		out.Answers[q.ID] = entry
	}
	for _, q := range d.Unanswered() {
		out.Unanswered = append(out.Unanswered, q.ID)
	}
	return out
}

// WriteJSON writes the result as indented JSON followed by a newline.
func (o Output) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(o)
}
