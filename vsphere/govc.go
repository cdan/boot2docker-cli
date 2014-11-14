package vsphere

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"
)

func init() {
}

var (
	ErrGOVCNotFound = errors.New("govc not found")
)

func govc(args ...string) error {
	cmd := exec.Command(cfg.Govc, args...)
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		log.Printf("executing: %v %v", cfg.Govc, strings.Join(args, " "))
	}
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.Error); ok && ee == exec.ErrNotFound {
			return ErrGOVCNotFound
		}
		return err
	}
	return nil
}

func govcOutErr(args ...string) (string, string, error) {
	cmd := exec.Command(cfg.Govc, args...)
	if verbose {
		log.Printf("executing: %v %v", cfg.Govc, strings.Join(args, " "))
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		if ee, ok := err.(*exec.Error); ok && ee == exec.ErrNotFound {
			err = ErrGOVCNotFound
		}
	}
	return stdout.String(), stderr.String(), err
}
