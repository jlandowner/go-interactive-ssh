package interactive-ssh

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultTimeOut archeve, Command is canneled. You can change by WithTimeoutOption
	DefaultTimeOut = time.Second * 5
)

var (
	// DefaultCallback is called after Command and just sleep in a second. You can change by WithCallbackOption
	DefaultCallback = func(c *Command) (bool, error) {
		time.Sleep(time.Second)
		return true, nil
	}
	// ErrReturnCodeNotZero is error in command exit with non zero
	ErrReturnCodeNotZero = errors.New("return code is not 0")
)

// Command has Input config and Output in remote host.
// Input is line of command execute in remote host.
// Callback is called after input command is finished. You can check whether Output is exepected in this function.
// NextCommand is called after Callback and called only Callback returns "true". NextCommand cannot has another NextCommand.
// DefaultNextCommand is called after Callback and called only Callback returns "false".
// ReturnCodeCheck is "true", Input is added ";echo $?" and check after Output is 0. Also you can manage retrun code in Callback.
// OutputLevel is logging level of command. Secret command should be set Silent
// Result is Command Output. You can use this in Callback, NextCommand, DefaultNextCommand functions.
type Command struct {
	Input              string
	Callback           func(c *Command) (bool, error)
	NextCommand        func(c *Command) *Command
	DefaultNextCommand func(c *Command) *Command
	ReturnCodeCheck    bool
	OutputLevel        OutputLevel
	Timeout            time.Duration
	Result             *CommandResult
}

// CommandResult has command output and return code in remote host
type CommandResult struct {
	Output     []string
	Lines      int
	ReturnCode int
}

// OutputLevel set logging level of command
type OutputLevel int

const (
	// Silent logs nothing
	Silent OutputLevel = iota
	// Info logs only start and end of command
	Info
	// Output logs command output in remote host
	Output
)

// NewCommand return Command with given options
func NewCommand(input string, options ...Option) *Command {
	c := &Command{
		Input:           input,
		OutputLevel:     Info,
		Timeout:         DefaultTimeOut,
		Callback:        DefaultCallback,
		ReturnCodeCheck: true,
		Result:          &CommandResult{},
	}

	for _, opt := range options {
		opt.Apply(c)
	}

	if c.ReturnCodeCheck {
		c.Input += ";echo $?"
	}
	return c
}

func (c *Command) wait(ctx context.Context, out <-chan string) error {
	timeout, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	for {
		select {
		case v := <-out:
			c.Result.Output = strings.Split(v, "\r\n")
			c.Result.Lines = len(c.Result.Output)

			if c.ReturnCodeCheck {
				if len([]byte(v)) == 0 {
					return errors.New("output is 0 byte")
				}

				ok, err := c.checkReturnCode()
				if err != nil {
					return err
				}
				if !ok {
					return ErrReturnCodeNotZero
				}
			}
			return nil
		case <-timeout.Done():
			msg := fmt.Sprintf("[%v] is canceled by timeout or by parent", c.Input)
			return errors.New(msg)
		}
	}
}

func (c *Command) checkReturnCode() (bool, error) {
	if c.Result.Lines-2 < 0 {
		return false, errors.New("Couldn't check return code")
	}

	returnCode, err := strconv.Atoi(c.Result.Output[c.Result.Lines-2])
	if err != nil {
		return false, err
	}

	c.Result.ReturnCode = returnCode
	if c.Result.ReturnCode == 0 {
		return true, nil
	}
	return false, nil
}

func (c *Command) output() ([]string, bool) {
	if c.OutputLevel != Output {
		return nil, false
	}
	var output []string
	for i := 0; i < c.Result.Lines-1; i++ {
		if c.Result.Output[i] == "0" {
			break
		}
		output = append(output, c.Result.Output[i])
	}
	return output, true
}
