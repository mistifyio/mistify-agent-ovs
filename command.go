package ovs

import (
	"bytes"
	"io"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
)

type command struct {
	command string
	stdin   io.Reader
	stdout  io.Writer
}

func (c *command) Run(arg ...string) ([][]string, error) {
	cmd := exec.Command(c.command, arg...)

	// Set up command stdin, stdout, stderr
	var stdout, stderr bytes.Buffer

	cmd.Stderr = &stderr

	if c.Stdout != nil {
		cmd.Stdout = c.Stdout
	} else {
		cmd.Stdout = &stdout
	}

	if c.Stdin != nil {
		cmd.Stdin = c.Stdin
	}

	logFields := log.Fields{
		"command": cmd.Path,
		"args":    cmd.Args,
	}
	log.WithFields(logFields).Debug("running command")

	if err := cmd.Run(); err != nil {
		e := &CmdError{
			err:    err,
			stderr: stderr.String(),
		}
		log.WithFields(logFields).WithField("error", e).Error("command failed")
		return nil, e
	}

	// If stdout was specified, processing is left entirely up to the caller
	if c.Stdout != nil {
		return nil, nil
	}

	// Split output by lines then spaces
	lines := strings.Split(stdout.String(), "\n")
	// Remove empty last line
	lines = lines[:len(lines)-1]
	output := make([][]string, len(lines))
	for i, line := range lines {
		output[i] = strings.Fields(line)
	}

	return output, nil
}
