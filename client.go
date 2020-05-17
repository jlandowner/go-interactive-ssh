package issh

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"

	"golang.org/x/crypto/ssh"
)

// Client has configuration to interact a host
type Client struct {
	Sshconfig *ssh.ClientConfig
	Host      string
	Port      string
	Prompt    []Prompt
}

// NewClient return new go-interactive-ssh client
// prompts []Prompt is list of all expected prompt pettern.
// for example if you use normal user and switch to root user, '$' and '#' prompts must be given.
func NewClient(sshconfig *ssh.ClientConfig, host string, port string, prompts []Prompt) *Client {
	return &Client{
		Sshconfig: sshconfig,
		Host:      host,
		Port:      port,
		Prompt:    prompts,
	}
}

// Run execute given commands in remote host
func (c *Client) Run(ctx context.Context, cmds []*Command) error {
	url := c.Host + ":" + c.Port
	client, err := ssh.Dial("tcp", url, c.Sshconfig)
	if err != nil {
		return fmt.Errorf("error in ssh.Dial to %v %w", url, err)
	}

	defer client.Close()
	session, err := client.NewSession()

	if err != nil {
		return fmt.Errorf("error in client.NewSession to %v %w", url, err)
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return fmt.Errorf("error in session.RequestPty to %v %w", url, err)
	}

	w, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("error in session.StdinPipe to %v %w", url, err)
	}
	r, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error in session.StdoutPipe to %v %w", url, err)
	}
	in, out := listener(w, r, c.Prompt)
	if err := session.Start("/bin/sh"); err != nil {
		return fmt.Errorf("error in session.Start to %v %w", url, err)
	}

	<-out // ignore login output
	for _, cmd := range cmds {
		select {
		case <-ctx.Done():
			return errors.New("canceled by context")

		default:
			logf(cmd.OutputLevel, "[%v]: cmd [%v] starting...", c.Host, cmd.Input)

			in <- cmd
			err := cmd.wait(ctx, out)
			if err != nil {
				return fmt.Errorf("[%v]: Error in cmd [%v] at waiting %w", c.Host, cmd.Input, err)
			}

			if outputs, ok := cmd.output(); ok {
				for _, output := range outputs {
					fmt.Println(output)
				}
			}

			doNext, err := cmd.Callback(cmd)
			if err != nil {
				return fmt.Errorf("[%v]: Error in cmd [%v] Callback %w", c.Host, cmd.Input, err)
			}

			if doNext && cmd.NextCommand != nil {
				nextCmd := cmd.NextCommand(cmd)

				logf(nextCmd.OutputLevel, "[%v]:   next cmd [%v] starting...", c.Host, nextCmd.Input)

				in <- nextCmd
				err = nextCmd.wait(ctx, out)
				if err != nil {
					return fmt.Errorf("[%v]: Error in next cmd [%v] at waiting %w", c.Host, cmd.Input, err)
				}

				if outputs, ok := nextCmd.output(); ok {
					for _, output := range outputs {
						fmt.Println(output)
					}
				}

				_, err := nextCmd.Callback(nextCmd)
				if err != nil {
					return fmt.Errorf("[%v]: Error in next cmd [%v] Callback %w", c.Host, nextCmd.Input, err)
				}

				logf(nextCmd.OutputLevel, "[%v]:   next cmd [%v] done", c.Host, nextCmd.Input)

			}

			logf(cmd.OutputLevel, "[%v]: cmd [%v] done", c.Host, cmd.Input)
		}
	}
	session.Close()

	return nil
}

func listener(w io.Writer, r io.Reader, prompts []Prompt) (chan<- *Command, <-chan string) {
	in := make(chan *Command, 1)
	out := make(chan string, 1)
	var wg sync.WaitGroup
	wg.Add(1) //for the shell itself
	go func() {
		for cmd := range in {
			wg.Add(1)
			w.Write([]byte(cmd.Input + "\n"))
			wg.Wait()
		}
	}()
	go func() {
		var (
			buf [65 * 1024]byte
			t   int
		)
		for {
			n, err := r.Read(buf[t:])
			if err != nil {
				close(in)
				close(out)
				return
			}
			t += n
			if t < 2 {
				continue
			}
			if buf[t-1] == ':' {
				out <- string(buf[:t])
				t = 0
				wg.Done()
				continue
			}

			for _, p := range prompts {
				// fmt.Print(string(p.SufixPattern))
				if buf[t-p.SufixPosition] == p.SufixPattern {
					out <- string(buf[:t])
					t = 0
					wg.Done()
					break
				}
			}
		}
	}()
	return in, out
}

func logf(level OutputLevel, msg string, v ...interface{}) {
	format := "go-interactive-ssh: " + msg
	if level != Silent {
		log.Printf(format, v...)
	}
}
