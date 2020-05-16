[![GoReportCard](https://goreportcard.com/badge/github.com/jlandowner/go-interactive-ssh)](https://goreportcard.com/report/github.com/jlandowner/go-interactive-ssh)

# go-interactive-ssh

Go interactive ssh client. 

You can use each standard outputs in your Callback function and check command's output is expected. 

this client support for windows, macos, and linux.

## Install

```bash
go get -u "github.com/jlandowner/go-interactive-ssh"
```

## Example

```go:example/main.go
package main

import (
	"context"
	"log"

	"golang.org/x/crypto/ssh"

	issh "github.com/jlandowner/go-interactive-ssh"
)

func main() {
	ctx := context.Background()

	config := &ssh.ClientConfig{
		User:            "pi",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password("raspberry"),
		},
	}

	// create client
	client := issh.NewClient(config, "raspberrypi.local", "22", []issh.Prompt{issh.DefaultPrompt})

	// give Commands to client and Run
	err := client.Run(ctx, commands())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("OK")
}

// make Command structs executed sequentially in remote host.  
func commands() []*issh.Command {
	return []*issh.Command{
		issh.CheckUser("pi"),
		issh.NewCommand("pwd", issh.WithOutputLevelOption(issh.Output)),
		issh.ChangeDirectory("/tmp"),
		issh.NewCommand("ls -l", issh.WithOutputLevelOption(issh.Output)),
	}
}
```

## Usage

### Setup Client

Setup Client by NewClient().

```go:client.go
func NewClient(sshconfig *ssh.ClientConfig, host string, port string, prompts []Prompt) *Client 
```
SSH client settings is the same as standard `ssh.ClientConfig`.

As the last argument, `[]Prompt` is a list of `Prompt` struct.

`Prompt` is used to confirm whether command execution is completed in the SSH login shell.

```go:prompt.go
type Prompt struct {
    SufixPattern  byte
    SufixPosition int
}
```

Normally, a prompt is like `pi@raspberrypi: ~ $`, so the '$' and '#' prompts are predefined in the package.

Client wait until each command outputs match this.

```go: prompt.go
var (
	// DefaultPrompt is prompt pettern like "pi @ raspberrypi: ~ $"
	DefaultPrompt = Prompt {
		SufixPattern: '$',
		SufixPosition: 2,
	}
	// DefaultRootPrompt is prompt pettern like "pi @ raspberrypi: ~ $"
	DefaultRootPrompt = Prompt {
		SufixPattern: '#',
		SufixPosition: 2,
	}
)
```

### Run and Command struct

All you have to do is just give Run() of this client a list of commands you want to execute in a remote host.

```go:client.go
func (c *Client) Run(ctx context.Context, cmds []*Command) error
```

The command is passed as a Command struct that contains the expected result, callback function, and so on.

```go:command.go
// Command has Input config and Output in remote host.
// Input is line of command execute in remote host.
// Callback is called after input command is finished. You can check whether Output is exepected in this function.
// NextCommand is called after Callback and called only Callback returns "true". NextCommand cannot has another NextCommand.
// ReturnCodeCheck is "true", Input is added ";echo $?" and check after Output is 0. Also you can manage retrun code in Callback.
// OutputLevel is logging level of command. Secret command should be set Silent
// Result is Command Output. You can use this in Callback, NextCommand, DefaultNextCommand functions.
type Command struct {
	Input           string
	Callback        func(c *Command) (bool, error)
	NextCommand     func(c *Command) *Command
	ReturnCodeCheck bool
	OutputLevel     OutputLevel
	Timeout         time.Duration
	Result          *CommandResult
}
```

You can generate a Command by NewCommand()

```go: command.go
func NewCommand(input string, options ...Option) *Command
```

Other than Input, set it with a function that implements the Option interface starting with `With...`.


```go: option.go
func WithNoCheckReturnCodeOption() *withNoCheckReturnCode
func WithOutputLevelOption(v OutputLevel) *withOutputLevel
func WithTimeoutOption(v time.Duration) *withTimeout
func WithCallbackOption(v func (c *Command) (bool, error)) *withCallback
func WithNextCommandOption(v func (c *Command) * Command) *withNextCommand
```

By default, "; echo $?" is added to the command passed to all Input and executed, and the return code is checked.
If you do not want to check the return code, such as when command has a standard input, add `WithNoCheckReturnCodeOption()`.


## Expect

As a function of Expect, you can use command output in Callback.

Set callback function by NewCommand options

```go:option.go
// WithCallbackOption is option function called after command is finished
func WithCallbackOption(v func(c *Command) (bool, error)) *withCallback
```

In the callback function, the Command struct with the stdout is received as an argument, so you can describe your own processing from the command execution result.

You can refer to the command execution result with `c.Result` in callback.

```go:command.go
type Command struct {
	...
	Result          *CommandResult
}

type CommandResult struct {
	Output     []string
	Lines      int
	ReturnCode int
}
```

In addition, you can add a command that can be executed only when the callback function returns `true` by `WithNextCommandOption()`.

```go:option.go
// WithNextCommandOption is option function called after Callback func return true
func WithNextCommandOption(v func(c *Command) *Command) *withNextCommand
```

The summary is as follows.

- __WithCallbackOption(v func(c *Command) (bool, error))__
	- Augument "c" is previous Command pointer. You can get command output in stdout by c.Result.Output.
	- Returned bool is whether run WithNextCommandOption function or not.
	- If err is not nil, issh.Run will exit.
	(Example see [CheckUser](https://github.com/jlandowner/go-interactive-ssh/blob/master/commands.go#L64))

- __WithNextCommandOption(v v func(c *Command) *Command)__
	- Augument "c" is previous Command pointer.
	- Returned Command is executed after previous Command in remote host.
	- It's useful when you have to type sequentially in stdin. (Example see [SwitchUser](https://github.com/jlandowner/go-interactive-ssh/blob/master/commands.go#L28))
