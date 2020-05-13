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
		issh.ChangeDirectory("/tmp"),
		issh.NewCommand("sleep 2"),
		issh.NewCommand("ls -l", issh.WithOutputLevelOption(issh.Output)),
	}
}
