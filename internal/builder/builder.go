package builder

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
)

type Builder struct {
	Command   string
	Directory string
}

func New(command, directory string) *Builder {
	return &Builder{
		Command:   command,
		Directory: directory,
	}
}

func (b *Builder) Build() error {
	cmd := exec.Command("sh", "-c", b.Command)
	cmd.Dir = b.Directory

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	log.Printf("Building with command: %s", b.Command)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Build failed: %s\n", stderr.String())
		return err
	}

	log.Printf("Build succeeded: %s\n", out.String())
	return nil

}
