// Command quizme presents a YAML questionnaire as a desktop form and
// writes the answers back into the file it came from.
//
// Usage:
//
//	quizme path/to/questionnaire.yaml
//	quizme --validate path/to/questionnaire.yaml
//	quizme --text-size 130 path/to/questionnaire.yaml
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
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/markwylde/quizme/internal/config"
	"github.com/markwylde/quizme/internal/questionnaire"
	"github.com/markwylde/quizme/internal/ui"
)

// The icon is compiled in, like the fonts, so one binary carries everything it
// needs to look like itself. It lives here rather than beside the form because
// a package can only embed what sits beside it, and the drawing belongs at the
// repository root where the packaging tools look for it too.
//
//go:embed icon.svg
var iconSVG []byte

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
type presenter func(*questionnaire.Document, ui.Options) (questionnaire.Status, error)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, present))
}

// present is the real form: check for a desktop session, then open the window.
func present(doc *questionnaire.Document, opts ui.Options) (questionnaire.Status, error) {
	if err := ui.CheckDisplay(); err != nil {
		return "", err
	}
	return ui.Run(doc, iconSVG, opts)
}

func run(args []string, stdout, stderr io.Writer, show presenter) int {
	opts, err := parseArgs(args, stderr)
	if err != nil {
		return exitError
	}
	path := opts.path

	doc, err := questionnaire.Load(path)
	if err != nil {
		fmt.Fprintf(stderr, "quizme: %s could not be read as a questionnaire:\n%v\n", path, err)
		return exitError
	}

	if opts.validate {
		// Loading is the whole check, and it has already happened. Nothing is
		// presented and nothing is written, so the file is left as found.
		fmt.Fprintf(stderr, "quizme: %s is a valid questionnaire (%d questions)\n", path, len(doc.Questions))
		return exitSubmitted
	}

	status, err := show(doc, displayOptions(opts, stderr))
	if err != nil {
		fmt.Fprintf(stderr, "quizme: %v\n", err)
		return exitError
	}
	if !status.Valid() || status == questionnaire.StatusPending {
		fmt.Fprintf(stderr, "quizme: the form returned an unusable outcome %q\n", status)
		return exitError
	}

	if err := doc.Save(status, time.Now()); err != nil {
		fmt.Fprintf(stderr, "quizme: %v\n", err)
		return exitError
	}

	// The answers go to stdout only once they are safely on disk, so a caller
	// that trusts stdout is never ahead of the file.
	if err := doc.Result(status).WriteJSON(stdout); err != nil {
		fmt.Fprintf(stderr, "quizme: could not write the answers to stdout: %v\n", err)
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

// displayOptions settles the text size the form opens at, and how a change to
// it is remembered.
//
// The responder's config is a convenience, so nothing wrong with it is fatal: a
// problem is a warning on stderr and the form carries on. stdout and the exit
// code belong to the questionnaire and never hear about it.
func displayOptions(opts options, stderr io.Writer) ui.Options {
	warn := func(err error) { fmt.Fprintf(stderr, "quizme: warning: %v\n", err) }

	path, pathErr := config.Path()
	out := ui.Options{TextSize: config.DefaultTextSize}

	switch {
	case opts.textSize != 0:
		// The flag is for this run only, so the saved size is not even read:
		// a broken config has nothing to say about a size nobody asked it for.
		out.TextSize = opts.textSize
	case pathErr != nil:
		warn(pathErr)
	default:
		cfg, warnings := config.Load(path)
		for _, w := range warnings {
			warn(w)
		}
		out.TextSize = cfg.TextSize
	}

	out.OnTextSize = func(percent int) {
		if pathErr != nil {
			warn(pathErr)
			return
		}
		// Loaded fresh each time, so a size saved here never disturbs anything
		// else in the file -- including anything written since the form opened.
		cfg, _ := config.Load(path)
		cfg.TextSize = percent
		if err := config.Save(path, cfg); err != nil {
			warn(fmt.Errorf("the text size could not be remembered: %w", err))
		}
	}
	return out
}

// options is one parsed invocation.
type options struct {
	path string
	// validate checks the questionnaire and stops, without presenting it. It is
	// for a caller checking a file it has just written -- a plain run already
	// validates before it opens anything.
	validate bool
	// textSize is the size to open the form at for this run, or 0 to use the
	// responder's saved size.
	textSize int
}

func parseArgs(args []string, stderr io.Writer) (options, error) {
	fs := flag.NewFlagSet("quizme", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { fmt.Fprint(stderr, usage) }

	var opts options
	fs.BoolVar(&opts.validate, "validate", false, "check the questionnaire and exit, without opening a window")
	fs.Func("text-size", "open the form at this text size, as a percentage", func(v string) error {
		n, err := strconv.Atoi(v)
		if err != nil || !config.ValidTextSize(n) {
			return fmt.Errorf("want a percentage from %d to %d in steps of %d, got %q",
				config.MinTextSize, config.MaxTextSize, config.TextSizeStep, v)
		}
		opts.textSize = n
		return nil
	})

	if err := fs.Parse(args); err != nil {
		return options{}, err
	}
	switch fs.NArg() {
	case 1:
		opts.path = fs.Arg(0)
		return opts, nil
	case 0:
		fmt.Fprint(stderr, "quizme: no questionnaire given\n\n"+usage)
		return options{}, errors.New("no questionnaire given")
	default:
		fmt.Fprintf(stderr, "quizme: expected one questionnaire, got %d\n\n%s", fs.NArg(), usage)
		return options{}, errors.New("too many arguments")
	}
}

const usage = `usage: quizme [--validate] [--text-size <percent>] <questionnaire.yaml>

Opens the questionnaire as a desktop form. On submit or save the answers are
written back into the same file and printed to stdout as JSON.

  --validate   Check the questionnaire and exit without opening a window, and
               without writing to the file. A plain run already validates
               before it presents anything, so this is for checking a
               questionnaire you have just written.

  --text-size  Open the form at this text size for this run only, as a
               percentage from 70 to 200 in steps of 10. Without it the form
               opens at the size last chosen on it, which is remembered in
               ~/.config/quizme/config.yaml.

Exit codes:
  0  submitted   every visible required question was answered
  1  error       the questionnaire could not be read, shown, or written
  2  dismissed   the responder discarded their answers
  3  saved       answers were kept with required questions outstanding
`
