package workflow

import "errors"

type Step string

type EventType string

const (
	StepValidate   Step = "validate"
	StepDetectGame Step = "detect_game"
	StepVersion    Step = "version"
	StepDownload   Step = "download"
	StepSymlink    Step = "symlink"
	StepDone       Step = "done"
)

const (
	EventStepStart   EventType = "step_start"
	EventStepEnd     EventType = "step_end"
	EventDownload    EventType = "download"
	EventError       EventType = "error"
	EventSymlinkDone EventType = "symlink_done"
)

var ErrInvalidPath = errors.New("invalid path")

var ErrUnknownGame = errors.New("unknown game")

var ErrSymlinkFailed = errors.New("symlink failed")

type Event struct {
	Type     EventType
	Step     Step
	Message  string
	FileName string
	Index    int
	Total    int
	Err      error
}
