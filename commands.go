package interactive-ssh

import (
	"errors"
	"fmt"
	"time"
)

// ChangeDirectory run "cd" command in remote host
func ChangeDirectory(tgtdir string) *Command {
	cd := fmt.Sprintf("cd %v;pwd", tgtdir)

	callback := func(c *Command) (bool, error) {
		time.Sleep(time.Second)
		// fmt.Println(c.Result.Output[c.Result.Lines-2])
		if c.Result.Output[c.Result.Lines-2] != tgtdir {
			msg := fmt.Sprintf(
				"wrong output in cd expect %v got %v", tgtdir, c.Result.Output[c.Result.Lines-2])
			return false, errors.New(msg)
		}
		return true, nil
	}
	c := NewCommand(cd, WithCallbackOption(callback),
		WithNoCheckReturnCodeOption(), WithOutputLevelOption(Output))
	return c
}

// SwitchUser run "su - xxx" command in remote host
func SwitchUser(user, password string, newUserPrompt Prompt) *Command {
	su := "su - " + user

	callback := func(c *Command) (bool, error) {
		time.Sleep(time.Second)
		expect := "パスワード:" //TODO support "Password:" or other prompt pattern
		if c.Result.Output[c.Result.Lines-1] != expect {
			msg := fmt.Sprintf(
				"wrong output in su expect %v got %v", expect, c.Result.Output[c.Result.Lines-1])
			return false, errors.New(msg)
		}
		return true, nil
	}

	nextCommand := func(c *Command) *Command {
		nextcallback := func(c *Command) (bool, error) {
			time.Sleep(time.Second * 1)
			got := c.Result.Output[c.Result.Lines-1]
			if got[len(got)-newUserPrompt.SufixPosition] != newUserPrompt.SufixPattern {
				fmt.Println(got)
				msg := fmt.Sprintf(
					"wrong output in su expect %v(%v) got %v(%v) RootPassword may be invalid",
					string(newUserPrompt.SufixPattern), newUserPrompt.SufixPattern, string(got[len(got)-2]), got[len(got)-2])
				return false, errors.New(msg)
			}
			return true, nil
		}
		return NewCommand(password, WithCallbackOption(nextcallback),
			WithNoCheckReturnCodeOption(), WithOutputLevelOption(Silent))
	}

	return NewCommand(su, WithCallbackOption(callback),
		WithNextCommandOption(nextCommand), WithNoCheckReturnCodeOption())
}

// CheckUser check current login user is expected in remote host
func CheckUser(expectUser string) *Command {
	whoami := "whoami"
	callback := func(c *Command) (bool, error) {
		if c.Result.Lines-3 < 0 {
			return false, errors.New("user is not expected")
		}
		user := c.Result.Output[c.Result.Lines-3]
		if user != expectUser {
			msg := fmt.Sprintf("user is invalid expected %v got %v", expectUser, user)
			return false, errors.New(msg)
		}
		return true, nil
	}
	return NewCommand(whoami, WithCallbackOption(callback), WithOutputLevelOption(Output))
}

// Exit run "exit" command in remote host
func Exit() *Command {
	return NewCommand("exit", WithNoCheckReturnCodeOption())
}
