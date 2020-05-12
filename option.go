package issh

import (
	"time"
)

// Option is Command struct option
type Option interface {
	Apply(*Command)
}

// NoCheckReturnCodeOption is option whether add ";echo $?" and check return code after command
func WithNoCheckReturnCodeOption() *withNoCheckReturnCode {
	opt := withNoCheckReturnCode(false)
	return &opt
}

type withNoCheckReturnCode bool

func (w *withNoCheckReturnCode) Apply(c *Command) {
	c.ReturnCodeCheck = false
}

// WithOutputLevelOption is option of command log print
func WithOutputLevelOption(v OutputLevel) *withOutputLevel {
	opt := withOutputLevel(v)
	return &opt
}

type withOutputLevel OutputLevel

func (w *withOutputLevel) Apply(c *Command) {
	c.OutputLevel = OutputLevel(*w)
}

// WithTimeoutOption is option time.Duration to command timeout
func WithTimeoutOption(v time.Duration) *withTimeout {
	opt := withTimeout(v)
	return &opt
}

type withTimeout time.Duration

func (w *withTimeout) Apply(c *Command) {
	c.Timeout = time.Duration(*w)
}

// WithCallbackOption is option function called after command is finished
func WithCallbackOption(v func(c *Command) (bool, error)) *withCallback {
	opt := withCallback(v)
	return &opt
}

type withCallback func(c *Command) (bool, error)

func (w *withCallback) Apply(c *Command) {
	c.Callback = *w
}

// WithNextCommandOption is option function called after Callback func return true
func WithNextCommandOption(v func(c *Command) *Command) *withNextCommand {
	opt := withNextCommand(v)
	return &opt
}

type withNextCommand func(c *Command) *Command

func (w *withNextCommand) Apply(c *Command) {
	c.NextCommand = *w
}
