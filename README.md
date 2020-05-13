[![GoReportCard](https://goreportcard.com/badge/github.com/jlandowner/go-interactive-ssh)](https://goreportcard.com/report/github.com/jlandowner/go-interactive-ssh)

# go-interactive-ssh

Go interactive ssh client. 

You can use each standard outputs in your Callback function and check command's output is expected. 

## Install

```bash
go get -u "github.com/jlandowner/go-interactive-ssh"
```

## Usage

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
You cam generate a Command by NewCommand()

Command struct is here. 

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

## Expect

You can use command output in Callback.
Set callback function by NewCommand options

- __WithCallbackOption(v func(c *Command) (bool, error))__
	- Augument "c" is previous Command pointer. You can get command output in stdout by c.Result.Output.
	- Returned bool is whether run WithNextCommandOption function or not.
	- If err is not nil, issh.Run will exit.

- __WithNextCommandOption(v v func(c *Command) *Command)__
	- Augument "c" is previous Command pointer.
	- Returned Command is executed after previous Command in remote host.
	- It's useful when you have to type sequentially in stdin. (Example see [SwitchUser](https://github.com/jlandowner/go-interactive-ssh/blob/master/commands.go#L28))


### Example
```go
func commands() []*issh.Command {
	return []*issh.Command{
        issh.SwitchUser("pi2", "password", issh.DefaultPrompt),
        issh.CheckUser("pi2"),
		issh.ChangeDirectory("/tmp"),
		issh.NewCommand("./my-script", issh.WithOutputLevelOption(issh.Output)),
		issh.NewCommand("ls -l my-script.log",
			issh.WithCallbackOption(
				func(c *issh.Command) (bool, error) {
					if c.Result.ReturnCode != 0 {
						// TODO when return code is not zero.
						return false, issh.ErrReturnCodeNotZero
					}
					// TODO when return code is zero.
					lines := c.Result.Output
					for _, line := range lines {
						row := strings.Fields(line)
						// use output
					}
					return false, nil
				},
			)),
	}
}
```
