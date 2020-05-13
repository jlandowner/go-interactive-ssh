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

	client := issh.NewClient(config, "raspberrypi.local", "22", []issh.Prompt{issh.DefaultPrompt})

	err := client.Run(ctx, commands())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("OK")
}

func commands() []*issh.Command {
	return []*issh.Command{
		issh.CheckUser("pi"),
		issh.NewCommand("pwd", issh.WithOutputLevelOption(issh.Output)),
		issh.ChangeDirectory("/tmp"),
		issh.NewCommand("ls -l"),
	}
}
```

## Expect output

You can use command output in Callback.
NewCommand has some options.

- __WithCallbackOption(v func(c *Command) (bool, error))__
	- "c" is previous Command. You can get command output in stdout by c.Result.Output.
	- Returned bool is whether run WithNextCommandOption function or not.
	- If err is not nil, issh.Run will exit.

- __WithNextCommandOption(v v func(c *Command) *Command)__
	- "c" is previous Command.
	- Returned Command is executed in remote host.
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
						return false, issh.ErrReturnCodeNotZero
					}
					lines := c.Result.Output
					for _, line := range lines {
						row := strings.Fields(line)
						// TODO ...
					}
					return false, nil
				},
			)),
	}
}
```
