package skeletonkit

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/mod/modfile"
)

func ModInit(path string) (string, error) {
	var mf modfile.File
	if err := mf.AddModuleStmt(path); err != nil {
		return "", fmt.Errorf("create go.mod: %w", err)
	}

	gov, err := goVersion()
	if err != nil {
		return "", fmt.Errorf("create go.mod: %w", err)
	}

	if err := mf.AddGoStmt(gov); err != nil {
		return "", fmt.Errorf("create go.mod: %w", err)
	}

	b, err := mf.Format()
	if err != nil {
		return "", fmt.Errorf("create go.mod: %w", err)
	}

	return string(b), nil
}

func goVersion() (string, error) {
	var stdout bytes.Buffer
	cmd := exec.Command("go", "env", "GOVERSION")
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return strings.TrimPrefix(strings.TrimSpace(stdout.String()), "go"), nil
}
