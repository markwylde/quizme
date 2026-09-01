package questionnaire

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// validate applies the document-wide rules that can only be checked once every
// question has been read: identity, option counts, bounds, conditional
// references, and the validity of any answers already recorded in the file.
func validate(doc *Document, errs *Errors) {
	seen := map[string]int{}
	for i, q := range doc.Questions {
		if q.ID == "" {
			continue
		}
		if first, dup := seen[q.ID]; dup {
			errs.addf("question %q at questions[%d] repeats the id already used at questions[%d]; ids must be unique", q.ID, i, first)
			continue
		}
		seen[q.ID] = i
	}

	for _, q := range doc.Questions {
		validateQuestion(q, errs)
	}

	validateShowIfRefs(doc, seen, errs)
	validateShowIfCycles(doc, errs)
}

func validateQuestion(q *Question, errs *Errors) {
	where := fmt.Sprintf("question %q", q.ID)

	if q.Type.NeedsOptions() {
		switch {
		case len(q.Options) == 0:
			errs.addf("%s: type %s requires `options`", where, q.Type)
		case len(q.Options) < 2:
			errs.addf("%s: type %s requires at least two options, found %d", where, q.Type, len(q.Options))
		}
		if dup := firstDuplicate(q.Options); dup != "" {
			errs.addf("%s: option %q is listed more than once", where, dup)
		}
	}

	if q.Min != nil && q.Max != nil && *q.Min > *q.Max {
		errs.addf("%s: `min` (%g) is greater than `max` (%g)", where, *q.Min, *q.Max)
	}

	if q.Type == TypeScale {
		lo, hi := q.ScaleBounds()
		if lo >= hi {
			errs.addf("%s: scale needs min below max, got %d and %d", where, lo, hi)
		}
	}

	if q.Answer != nil {
		normalized, err := NormalizeAnswer(q, q.Answer)
		if err != nil {
			errs.addf("%s: recorded `answer` is not valid: %v", where, err)
		} else {
			q.Answer = normalized
		}
	}
}

func validateShowIfRefs(doc *Document, ids map[string]int, errs *Errors) {
	for _, q := range doc.Questions {
		for _, c := range q.ShowIf {
			if _, ok := ids[c.Question]; !ok {
				errs.addf("question %q: `show_if` refers to unknown question %q", q.ID, c.Question)
				continue
			}
			if c.Question == q.ID {
				errs.addf("question %q: `show_if` refers to itself", q.ID)
			}
		}
	}
}

// validateShowIfCycles reports any circular chain of show_if references. A
// cycle can never resolve, so evaluation would either loop or pick an arbitrary
// answer; rejecting at load keeps visibility deterministic.
func validateShowIfCycles(doc *Document, errs *Errors) {
	const (
		white = 0 // unvisited
		grey  = 1 // on the current path
		black = 2 // fully explored
	)
	colour := map[string]int{}
	var path []string
	var reported = map[string]bool{}

	var walk func(id string)
	walk = func(id string) {
		q := doc.Question(id)
		if q == nil {
			return
		}
		colour[id] = grey
		path = append(path, id)
		for _, c := range q.ShowIf {
			switch colour[c.Question] {
			case white:
				walk(c.Question)
			case grey:
				cycle := cycleFrom(path, c.Question)
				key := canonicalCycle(cycle)
				if !reported[key] {
					reported[key] = true
					errs.addf("circular `show_if` chain: %s", strings.Join(append(cycle, c.Question), " -> "))
				}
			}
		}
		path = path[:len(path)-1]
		colour[id] = black
	}

	for _, q := range doc.Questions {
		if colour[q.ID] == white {
			walk(q.ID)
		}
	}
}

func cycleFrom(path []string, start string) []string {
	for i, id := range path {
		if id == start {
			return append([]string(nil), path[i:]...)
		}
	}
	return append([]string(nil), path...)
}

// canonicalCycle gives a rotation-independent key so the same cycle is only
// reported once however it was entered.
func canonicalCycle(cycle []string) string {
	sorted := append([]string(nil), cycle...)
	sort.Strings(sorted)
	return strings.Join(sorted, "\x00")
}

func firstDuplicate(values []string) string {
	seen := map[string]bool{}
	for _, v := range values {
		if seen[v] {
			return v
		}
		seen[v] = true
	}
	return ""
}

// NormalizeAnswer checks a raw answer against its question's type and converts
// it to the canonical Go representation for that type:
//
//	select, text, textarea  string
//	multiselect, rank       []string
//	number                  float64
//	boolean                 bool
//	scale                   int
//
// A blank answer normalizes to nil, so clearing a control and never touching it
// are the same thing.
func NormalizeAnswer(q *Question, raw any) (any, error) {
	if isBlank(raw) {
		return nil, nil
	}

	switch q.Type {
	case TypeSelect:
		s, ok := scalarString(raw)
		if !ok {
			return nil, fmt.Errorf("expected a single value, got %T", raw)
		}
		if !containsString(q.Options, s) {
			return nil, fmt.Errorf("%q is not one of the options (%s)", s, strings.Join(q.Options, ", "))
		}
		return s, nil

	case TypeMultiselect:
		items, err := stringList(raw)
		if err != nil {
			return nil, err
		}
		if dup := firstDuplicate(items); dup != "" {
			return nil, fmt.Errorf("%q is selected more than once", dup)
		}
		for _, it := range items {
			if !containsString(q.Options, it) {
				return nil, fmt.Errorf("%q is not one of the options (%s)", it, strings.Join(q.Options, ", "))
			}
		}
		return items, nil

	case TypeRank:
		items, err := stringList(raw)
		if err != nil {
			return nil, err
		}
		if err := checkPermutation(items, q.Options); err != nil {
			return nil, err
		}
		return items, nil

	case TypeText, TypeTextarea:
		s, ok := raw.(string)
		if !ok {
			var alt bool
			if s, alt = scalarString(raw); !alt {
				return nil, fmt.Errorf("expected text, got %T", raw)
			}
		}
		return s, nil

	case TypeNumber:
		f, ok := numeric(raw)
		if !ok {
			return nil, fmt.Errorf("expected a number, got %v", raw)
		}
		if q.Min != nil && f < *q.Min {
			return nil, fmt.Errorf("%g is below the minimum of %g", f, *q.Min)
		}
		if q.Max != nil && f > *q.Max {
			return nil, fmt.Errorf("%g is above the maximum of %g", f, *q.Max)
		}
		return f, nil

	case TypeBoolean:
		b, ok := raw.(bool)
		if !ok {
			return nil, fmt.Errorf("expected true or false, got %v", raw)
		}
		return b, nil

	case TypeScale:
		f, ok := numeric(raw)
		if !ok {
			return nil, fmt.Errorf("expected a whole number, got %v", raw)
		}
		n := int(f)
		if float64(n) != f {
			return nil, fmt.Errorf("expected a whole number, got %g", f)
		}
		lo, hi := q.ScaleBounds()
		if n < lo || n > hi {
			return nil, fmt.Errorf("%d is outside the scale range %d to %d", n, lo, hi)
		}
		return n, nil
	}

	return nil, fmt.Errorf("unknown type %q", q.Type)
}

// checkPermutation verifies that a rank answer covers every option exactly
// once, which is the only ordering that is meaningful to read back.
func checkPermutation(got, options []string) error {
	if len(got) != len(options) {
		return fmt.Errorf("ranking must list all %d options, got %d", len(options), len(got))
	}
	if dup := firstDuplicate(got); dup != "" {
		return fmt.Errorf("%q appears more than once in the ranking", dup)
	}
	// Report a value that does not belong before one that is missing: it names
	// what the responder actually did, rather than a consequence of it.
	for _, g := range got {
		if !containsString(options, g) {
			return fmt.Errorf("%q is not one of the options", g)
		}
	}
	for _, o := range options {
		if !containsString(got, o) {
			return fmt.Errorf("option %q is missing from the ranking", o)
		}
	}
	return nil
}

// scalarString renders a YAML scalar as the string used to match it against an
// option. Unquoted options like `3` or `true` decode as int and bool, so
// comparison has to happen on their textual form.
func scalarString(v any) (string, bool) {
	switch t := v.(type) {
	case string:
		return t, true
	case int:
		return strconv.Itoa(t), true
	case int64:
		return strconv.FormatInt(t, 10), true
	case float64:
		return strconv.FormatFloat(t, 'g', -1, 64), true
	case bool:
		return strconv.FormatBool(t), true
	}
	return "", false
}

func stringList(v any) ([]string, error) {
	switch t := v.(type) {
	case []string:
		return append([]string(nil), t...), nil
	case []any:
		out := make([]string, 0, len(t))
		for _, item := range t {
			s, ok := scalarString(item)
			if !ok {
				return nil, fmt.Errorf("expected a list of simple values, found %T", item)
			}
			out = append(out, s)
		}
		return out, nil
	}
	return nil, fmt.Errorf("expected a list, got %T", v)
}

func numeric(v any) (float64, bool) {
	switch t := v.(type) {
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case float64:
		return t, true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return f, err == nil
	}
	return 0, false
}

func containsString(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}
