package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/gokrazy/gokrazy"
)

// version is set at build time using -ldflags
// Example: go build -ldflags "-X main.version=$(git describe --tags --always --dirty)"
var version = "vx.x.x"

func podman(args ...string) error {
	podman := exec.Command("/usr/local/bin/podman", args...)
	podman.Env = expandPath(os.Environ())
	podman.Env = append(podman.Env, "TMPDIR=/tmp")
	podman.Stdin = os.Stdin
	podman.Stdout = os.Stdout
	podman.Stderr = os.Stderr
	if err := podman.Run(); err != nil {
		return fmt.Errorf("%v: %v", podman.Args, err)
	}
	return nil
}

func getLogLevelFromEnv() slog.Level {
	levelStr := os.Getenv("LOG_LEVEL")

	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func postgres() error {
	// Ensure we have an up-to-date clock, which in turn also means that
	// networking is up. This is relevant because podman takes what’s in
	// /etc/resolv.conf (nothing at boot) and holds on to it, meaning your
	// container will never have working networking if it starts too early.
	gokrazy.WaitForClock()

	slog.Info("Stopping existing postgres container if running")
	if err := podman("kill", "postgres"); err != nil {
		slog.Info("Could not kill postgres container (may not be running)", "error", err)
	}

	slog.Info("Removing existing postgres container if present")
	if err := podman("rm", "postgres"); err != nil {
		slog.Info("Could not remove postgres container (may not exist)", "error", err)
	}

	// You could podman pull here.

	if err := podman("run",
		"-td",
		"-v", "/perm/postgres_docker:/var/lib/postgresql/data:Z",
		"-p", "5432:5432",
		"--network", "host",
		"-e POSTGRES_PASSWORD",
		"--name", "postgres",
		"docker.io/library/postgres",
		"postgres"); err != nil {
		return err
	}

	return nil
}

func getGitTag() string {
	return version
}

func main() {
	_ = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: getLogLevelFromEnv(),
	}))

	// Get git tag information
	gitTag := getGitTag()
	slog.Info("Application version", "git_tag", gitTag)

	if err := postgres(); err != nil {
		slog.Error("Failed to start postgres container", "error", err)
		os.Exit(1)
	}
	slog.Info("Postgres container started successfully")
}

// expandPath returns env, but with PATH= modified or added
// such that both /user and /usr/local/bin are included, which podman needs.
func expandPath(env []string) []string {
	extra := "/user:/usr/local/bin"
	found := false
	for idx, val := range env {
		parts := strings.Split(val, "=")
		if len(parts) < 2 {
			continue // malformed entry
		}
		key := parts[0]
		if key != "PATH" {
			continue
		}
		val := strings.Join(parts[1:], "=")
		env[idx] = fmt.Sprintf("%s=%s:%s", key, extra, val)
		found = true
	}
	if !found {
		const busyboxDefaultPATH = "/usr/local/sbin:/sbin:/usr/sbin:/usr/local/bin:/bin:/usr/bin"
		env = append(env, fmt.Sprintf("PATH=%s:%s", extra, busyboxDefaultPATH))
	}
	return env
}
