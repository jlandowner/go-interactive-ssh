package issh

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func TestRun(t *testing.T) {
	h := os.Getenv("SSH_HOST_IP")
	u := os.Getenv("SSH_LOGIN_USER")
	p := os.Getenv("SSH_LOGIN_PASS")
	if h == "" {
		h = "raspberrypi.local"
	}
	if u == "" {
		u = "pi"
	}
	if p == "" {
		p = "raspberry"
	}

	ctx := context.Background()

	config := &ssh.ClientConfig{
		User:            u,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(p),
		},
	}

	Client := NewClient(config, h, "22", []Prompt{DefaultPrompt, DefaultRootPrompt})

	err := Client.Run(ctx, testSwitchUser())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("OK")
}

func testSwitchUser() []*Command {
	files := []string{"go-interactive-ssh-test-1.txt", "go-interactive-ssh-test-2.txt", "go-interactive-ssh-test-3.txt"}

	var createFiles []*Command
	for _, file := range files {
		createFiles = append(createFiles, NewCommand("touch "+file))
	}

	filetest := NewCommand("ls -l go-interactive-ssh-test*.txt", WithCallbackOption(func(c *Command) (bool, error) {
		if c.Result.ReturnCode != 0 {
			fmt.Println("/tmp/go-interactive-ssh-test*.txt not found")
			return false, nil
		}
		var file string
		for _, output := range c.Result.Output {
			row := strings.Fields(output)
			if len(row) == 9 {
				file = row[8]
				fmt.Println(file)
			}
		}
		return true, nil

	}), WithOutputLevelOption(Output))

	var commands []*Command
	commands = append(commands,
		CheckUser("pi"),
		SwitchUser("dummy", "dummy", DefaultRootPrompt), //TODO change user to switch
		NewCommand("id", WithOutputLevelOption(Output)),
		ChangeDirectory("/tmp"))

	commands = append(commands, createFiles...)

	commands = append(commands,
		filetest,
		NewCommand("sleep 5", WithTimeoutOption(time.Second*10)))

	commands = append(commands, cleanup(files)...)
	commands = append(commands,
		Exit(),
		CheckUser("pi"),
	)

	return commands
}

func cleanup(files []string) []*Command {
	var deleteFiles []*Command
	for _, file := range files {
		f := file
		rm := NewCommand("ls -l "+f,
			WithCallbackOption(func(c *Command) (bool, error) {
				if c.Result.ReturnCode != 0 {
					fmt.Printf("%v not found\n", file)
					return false, nil
				}
				return true, nil
			}), WithNextCommandOption(func(c *Command) *Command {
				return NewCommand("rm -f " + f)
			}))
		deleteFiles = append(deleteFiles, rm)
	}
	return deleteFiles
}
