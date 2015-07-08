package ovs

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
)

type (
	command struct {
		command string
		stdin   io.Reader
		stdout  io.Writer
	}

	// Error is an error which is returned when the `zfs` or `zpool` shell
	// commands return with a non-zero exit code.
	cmdError struct {
		Err    error
		Stderr string
	}
)

// Error returns the string representation of an Error.
func (ce cmdError) Error() string {
	return fmt.Sprintf("%s: %s", ce.Err, ce.Stderr)
}

func (c *command) Run(arg ...string) ([]string, error) {
	cmd := exec.Command(c.command, arg...)

	// Set up command stdin, stdout, stderr
	var stdout, stderr bytes.Buffer

	cmd.Stderr = &stderr

	if c.stdout != nil {
		cmd.Stdout = c.stdout
	} else {
		cmd.Stdout = &stdout
	}

	if c.stdin != nil {
		cmd.Stdin = c.stdin
	}

	logFields := log.Fields{
		"command": cmd.Path,
		"args":    cmd.Args,
	}
	log.WithFields(logFields).Debug("running command")

	if err := cmd.Run(); err != nil {
		e := &cmdError{
			Err:    err,
			Stderr: stderr.String(),
		}
		log.WithFields(logFields).WithField("error", e).Error("command failed")
		return nil, e
	}

	// If stdout was specified, processing is left entirely up to the caller
	if c.stdout != nil {
		return nil, nil
	}

	// Split output by lines then spaces
	lines := strings.Split(stdout.String(), "\n")
	// Remove empty last line
	return lines[:len(lines)-1], nil
}
