// Command interrogate presents a YAML questionnaire as a desktop form and
// writes the answers back into the file it came from.
//
// Usage:
//
//	interrogate path/to/questionnaire.yaml
//	interrogate --validate path/to/questionnaire.yaml
//
// The exit code reports the outcome, so a caller can branch on it without
// parsing anything:
//
//	0  submitted   every visible required question was answered
//	1  error       the questionnaire could not be read, shown, or written
//	2  dismissed   the responder discarded their answers
//	3  saved       answers were kept with required questions outstanding
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/markwylde/interrogate/internal/questionnaire"
	"github.com/markwylde/interrogate/internal/ui"
)

// Exit codes. Each responder outcome is distinct from the others and from a
// failure, which is the point: branching should not require reading stdout.
const (
	exitSubmitted = 0
	exitError     = 1
	exitDismissed = 2
	exitSaved     = 3
)

// presenter shows a questionnaire and reports how the responder left it. It is
// a parameter so the command can be tested without a display.
type presenter func(*questionnaire.Document) (questionnaire.Status, error)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, present))
}

// present is the real form: check for a desktop session, then open the window.
func present(doc *questionnaire.Document) (questionnaire.Status, error) {
	if err := ui.CheckDisplay(); err != nil {
		return "", err
	}
	return ui.Run(doc)
}

func run(args []string, stdout, stderr io.Writer, show presenter) int {
	opts, err := parseArgs(args, stderr)
	if err != nil {
		return exitError
	}
	path := opts.path

	doc, err := questionnaire.Load(path)
	if err != nil {
		fmt.Fprintf(stderr, "interrogate: %s could not be read as a questionnaire:\n%v\n", path, err)
		return exitError
	}

	if opts.validate {
		// Loading is the whole check, and it has already happened. Nothing is
		// presented and nothing is written, so the file is left as found.
		fmt.Fprintf(stderr, "interrogate: %s is a valid questionnaire (%d questions)\n", path, len(doc.Questions))
		return exitSubmitted
	}

	status, err := show(doc)
	if err != nil {
		fmt.Fprintf(stderr, "interrogate: %v\n", err)
		return exitError
	}
	if !status.Valid() || status == questionnaire.StatusPending {
		fmt.Fprintf(stderr, "interrogate: the form returned an unusable outcome %q\n", status)
		return exitError
	}

	if err := doc.Save(status, time.Now()); err != nil {
		fmt.Fprintf(stderr, "interrogate: %v\n", err)
		return exitError
	}

	// The answers go to stdout only once they are safely on disk, so a caller
	// that trusts stdout is never ahead of the file.
	if err := doc.Result(status).WriteJSON(stdout); err != nil {
		fmt.Fprintf(stderr, "interrogate: could not write the answers to stdout: %v\n", err)
		return exitError
	}

	switch status {
	case questionnaire.StatusSubmitted:
		return exitSubmitted
	case questionnaire.StatusSaved:
		return exitSaved
	default:
		return exitDismissed
	}
}

// options is one parsed invocation.
type options struct {
	path string
	// validate checks the questionnaire and stops, without presenting it. It is
	// for a caller checking a file it has just written -- a plain run already
	// validates before it opens anything.
	validate bool
}

func parseArgs(args []string, stderr io.Writer) (options, error) {
	fs := flag.NewFlagSet("interrogate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { fmt.Fprint(stderr, usage) }

	var opts options
	fs.BoolVar(&opts.validate, "validate", false, "check the questionnaire and exit, without opening a window")

	if err := fs.Parse(args); err != nil {
		return options{}, err
	}
	switch fs.NArg() {
	case 1:
		opts.path = fs.Arg(0)
		return opts, nil
	case 0:
		fmt.Fprint(stderr, "interrogate: no questionnaire given\n\n"+usage)
		return options{}, errors.New("no questionnaire given")
	default:
		fmt.Fprintf(stderr, "interrogate: expected one questionnaire, got %d\n\n%s", fs.NArg(), usage)
		return options{}, errors.New("too many arguments")
	}
}

const usage = `usage: interrogate [--validate] <questionnaire.yaml>

Opens the questionnaire as a desktop form. On submit or save the answers are
written back into the same file and printed to stdout as JSON.

  --validate   Check the questionnaire and exit without opening a window, and
               without writing to the file. A plain run already validates
               before it presents anything, so this is for checking a
               questionnaire you have just written.

Exit codes:
  0  submitted   every visible required question was answered
  1  error       the questionnaire could not be read, shown, or written
  2  dismissed   the responder discarded their answers
  3  saved       answers were kept with required questions outstanding
`
