// Copyright © 2019 Jose Riguera <jriguera@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package runner

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-cmd/cmd"

	log "confinit/pkg/log"
)

// Executor
type Runner struct {
	command *cmd.Cmd
	show    bool
	Env     map[string]string
	Cmd     []string
	Timeout int
	Dir     string
	log     log.Logger
	status  *Status
}

type Status cmd.Status

// New is the contructor
func NewRunner(l log.Logger) *Runner {
	p := Runner{
		show:    true,
		Env:     make(map[string]string),
		log:     l,
		Timeout: 60,
		status:  nil,
	}
	return &p
}

func (p *Runner) SetEnv(env map[string]string) {
	p.Env = env
}

func (p *Runner) SetTimeout(t int) {
	p.Timeout = t
}

func (p *Runner) SetDir(d string) {
	p.Dir = d
}

func (p *Runner) Command(command []string) {
	bin := command[0]
	args := []string{}
	if len(command) > 1 {
		args = command[1:]
	}
	p.Cmd = command
	cmdOptions := cmd.Options{
		Buffered:  false,
		Streaming: true,
	}
	p.command = cmd.NewCmdOptions(cmdOptions, bin, args...)
	p.command.Dir = p.Dir
	for key, value := range p.Env {
		p.command.Env = append(p.command.Env, fmt.Sprintf("%s=%s", key, value))
	}
	p.show = true
}

func (p *Runner) run() *Status {
	statusChan := p.command.Start()
	if p.Timeout > 0 {
		// Stop command after timeout
		go func() {
			<-time.After(time.Duration(p.Timeout) * time.Second)
			p.log.Errorf("Timeout (%d s). Killing process", p.Timeout)
			p.command.Stop()
		}()
	}
	// Print STDOUT and STDERR lines streaming from Cmd
	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		// Done when both channels have been closed
		// https://dave.cheney.net/2013/04/30/curious-channels
		for p.command.Stdout != nil || p.command.Stderr != nil {
			select {
			case line, open := <-p.command.Stdout:
				if !open {
					p.command.Stdout = nil
					continue
				}
				p.print(line, false, false)
			case line, open := <-p.command.Stderr:
				if !open {
					p.command.Stderr = nil
					continue
				}
				p.print(line, true, false)
			}
		}
	}()
	finalStatus := <-statusChan
	if finalStatus.Exit == 0 {
		p.print("Exit", false, true)
	} else {
		errtxt := "Error"
		if finalStatus.Error != nil {
			errtxt = finalStatus.Error.Error()
		}
		p.print(errtxt, true, true)
	}
	status := Status(p.command.Status())
	return &status
}

func (p *Runner) Run() (int, error) {
	status := p.run()
	if status.Exit != 0 {
		if status.Error == nil {
			status.Error = fmt.Errorf("Error!")
		}
	}
	p.status = status
	return status.Exit, status.Error
}

func (p *Runner) Status() *Status {
	return p.status
}

func (p *Runner) String() string {
	return strings.Join(p.Cmd, " ")
}

func (p *Runner) print(out string, err, rc bool) {
	status := p.command.Status()
	if p.show {
		p.log.Debugf("Running (pid %d): %s", status.PID, strings.Join(p.Cmd, " "))
		p.show = false
	}
	// finished
	if rc {
		if err {
			p.log.Errorf("Done (pid %d): %s %d", status.PID, out, status.Exit)
		} else {
			p.log.Debugf("Done (pid %d): %s %d", status.PID, out, status.Exit)
		}
	} else {
		if err {
			p.log.Errorf("%d: %s", status.PID, out)
		} else {
			p.log.Debugf("%d: %s", status.PID, out)
		}
	}
}
