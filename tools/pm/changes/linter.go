package changes

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/coreos/go-semver/semver"
)

type CheckMode int

const (
	CheckStandard CheckMode = 0 + iota
	CheckPreRelease
	CheckRelease
)

type Linter struct {
	r    io.Reader
	mode CheckMode
}

type Failure struct {
	Line    int
	Message string
}

type Failures []Failure

func (fs Failures) String() string {
	buf := &strings.Builder{}
	for i, f := range fs {
		if i > 0 {
			_, _ = fmt.Fprint(buf, "\n")
		}
		_, _ = fmt.Fprintf(buf, " * Line %d: %s", f.Line, f.Message)
	}
	return buf.String()
}

type Error struct {
	Failures
}

func (e *Error) Error() string {
	return fmt.Sprintf("Change log linter check failed:\n%s", e.Failures.String())
}

func NewLinter(r io.Reader, mode CheckMode) *Linter {
	return &Linter{r, mode}
}

type checkStatus struct {
	previousVersion *semver.Version
	previousDate    string
	previousLine    int

	previousLineWasBlank  bool
	previousLineWasBullet bool

	Failures
}

func (s *checkStatus) Fail(
	lineNumber int,
	msg string,
) {
	if s.Failures == nil {
		s.Failures = Failures{}
	}
	s.Failures = append(s.Failures, Failure{lineNumber, msg})
}

func (s *checkStatus) Failf(
	lineNumber int,
	f string,
	args ...any,
) {
	s.Fail(lineNumber, fmt.Sprintf(f, args...))
}

func (l *Linter) Check() error {
	status := checkStatus{}

	s := bufio.NewScanner(l.r)
	n := 0
	for s.Scan() {
		n++
		l.checkLine(n, s.Text(), &status)
	}

	if len(status.Failures) > 0 {
		return &Error{status.Failures}
	} else {
		return nil
	}
}

var (
	versionHeading      = regexp.MustCompile(`^v(\d\S+) {2}(20\d\d-\d\d-\d\d)$`)
	logLineStart        = regexp.MustCompile(`^ \* (.*)$`)
	logLineContinuation = regexp.MustCompile(`^ {3}(.*)$`)
	blankLine           = regexp.MustCompile(`^$`)
	whitespaceLine      = regexp.MustCompile(`^\s+$`)
)

func (l *Linter) checkLine(
	lineNumber int,
	line string,
	status *checkStatus,
) {
	lineIsBlank := false
	lineIsBullet := false

	defer func() {
		status.previousLineWasBlank = lineIsBlank
		status.previousLineWasBullet = lineIsBullet
	}()

	if line == "WIP" || line == "WIP  TBD" {
		if lineNumber > 1 {
			status.Fail(lineNumber, "WIP found after line 1")
		}

		if l.mode == CheckRelease {
			status.Fail(lineNumber, "Found WIP line during release")
		}

		status.previousLine = lineNumber

		return
	}

	// we shouldn't get here if the first line is a WIP line
	if l.mode == CheckPreRelease && lineNumber == 1 {
		status.Fail(lineNumber, "WIP not found during pre-release check")
	}

	if m := versionHeading.FindStringSubmatch(line); m != nil {
		ver, date := m[1], m[2]
		version, err := semver.NewVersion(ver)
		if err != nil {
			status.Fail(lineNumber, "Unable to parse version number in heading")

			// this is fatal for this line, checks cannot continue
			status.previousLine = lineNumber
		}

		// version and date are in descending order in a changelog

		if status.previousVersion != nil && status.previousVersion.LessThan(*version) {
			status.Failf(lineNumber, "version error %s < %s from line %d",
				version, status.previousVersion, status.previousLine)
		}

		if status.previousDate != "" && status.previousDate < date {
			status.Failf(lineNumber, "date error %s < %s from line %d",
				date, status.previousDate, status.previousLine)
		}

		if lineNumber != 1 && !status.previousLineWasBlank {
			status.Fail(lineNumber, "version heading line missing blank line before it")
		}

		status.previousVersion = version
		status.previousDate = date
		status.previousLine = lineNumber

		return
	}

	if m := logLineStart.FindStringSubmatch(line); m != nil {
		if status.previousLine == 0 {
			status.Fail(lineNumber, "log bullet before first version heading or WIP")
		}

		if status.previousLine > 0 && lineNumber-1 == status.previousLine {
			status.Fail(lineNumber, "missing blank line before log bullet")
		}

		if status.previousLine > 0 && lineNumber > status.previousLine+2 && status.previousLineWasBlank {
			status.Fail(lineNumber, "extra blank line before log bullet")
		}

		lineIsBullet = true

		return
	}

	if m := logLineContinuation.FindStringSubmatch(line); m != nil {
		if status.previousLineWasBullet {
			lineIsBullet = true
		} else {
			status.Fail(lineNumber, "log line continuation has not bullet to continue")
		}

		return
	}

	if blankLine.MatchString(line) {
		if status.previousLineWasBlank {
			status.Fail(lineNumber, "consecutive blank lines")
		}

		lineIsBlank = true

		return
	}

	if whitespaceLine.MatchString(line) {
		status.Fail(lineNumber, "line looks blank, but has spaces in it")

		return
	}

	status.Fail(lineNumber, "badly formatted line")
}
