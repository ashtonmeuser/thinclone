package thinclone

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"regexp"
	"time"
)

const alphanumericSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func randAlphanumeric(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = alphanumericSet[seededRand.Intn(len(alphanumericSet))]
	}
	return string(b)
}

func validate(s string, r *regexp.Regexp) error {
	if !r.MatchString(s) {
		return errors.New("Failed regex match")
	}
	return nil
}

// ExecContext defines the command execution context
// Outside of testing, os/exec.Command should be used
type ExecContext = func(name string, arg ...string) *exec.Cmd

// SelectiveClone clones a repo only pulling files of specified extension
// Repo is cloned with a depth of one
// A sparse checkout is used to filter files to be pulled
func SelectiveClone(repoURI string, extension string) error {
	if err := validate(extension, regexp.MustCompile(`^\w+$`)); err != nil {
		return fmt.Errorf("Invalid extension: %s", extension)
	}
	if err := validate(repoURI, regexp.MustCompile(`^(https:\/\/|.+@)([\w.]+\.\w+)(:[\d]+)?(\/.+)*\.git$`)); err != nil {
		return fmt.Errorf("Invalid repository URI: %s", extension)
	}
	repoDir := fmt.Sprintf("repo-%s", randAlphanumeric(16)) // Name of dir in which the repo is cloned
	cwd, err := os.Getwd()
	if err != nil {
		return errors.New("Failed to get current working directory")
	}
	commands := [...]*exec.Cmd{
		exec.Command("mkdir", "-p", repoDir),
		exec.Command("git", "clone", repoURI, "--no-checkout", "--depth", "1", "."),
		exec.Command("git", "config", "core.sparsecheckout", "true"),
		exec.Command("git", "sparse-checkout", "set", fmt.Sprintf("*.%s", extension)),
		exec.Command("git", "checkout", "--"),
	}
	for i, c := range commands {
		if i > 0 {
			c.Dir = path.Join(cwd, repoDir) // Only repo dir creation should be executed at CWD level
		}
		if err := c.Run(); err != nil {
			exec.Command("rm", "-rf", repoDir).Run()
			return fmt.Errorf(`Command "%s" failed with output "%s"`, c.String(), err.Error())
		}
	}
	return nil
}
