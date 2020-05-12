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

	Client := issh.NewClient(config, "raspberrypi.local", "22", []issh.Prompt{issh.DefaultPrompt})

	err := Client.Run(ctx, commands())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("OK")
}

func commands() []*issh.Command {
	return []*issh.Command{
		issh.CheckUser("pi"),
		issh.NewCommand("id", issh.WithOutputLevelOption(issh.Output)),
		issh.ChangeDirectory("/tmp"),
		issh.NewCommand("ls -l"),
	}
}
