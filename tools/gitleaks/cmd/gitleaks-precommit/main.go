package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const cfgKeyEnable = "gitleaks.precommit.enable"

func main() {
	repoRoot := mustRepoRoot()
	_ = os.Chdir(repoRoot) // guarantee CWD=repo root

	enabled, _ := gitConfigBool(cfgKeyEnable, true)
	if !enabled {
		info("pre-commit disabled via git config (%s=false), skip", cfgKeyEnable)
		os.Exit(0)
	}

	// ensure gitleaks exists or install
	if !binaryExists("gitleaks") && !binaryExists(filepath.Join(repoRoot, "bin", exeName("gitleaks"))) {
		info("gitleaks not found — trying auto install")
		if err := autoInstallGitleaks(repoRoot); err != nil {
			fail("failed to install gitleaks: %v\nhint: install manually and retry", err)
		}
	}
	// put repoRoot/bin into PATH so fresh install is visible
	os.Setenv("PATH", filepath.Join(repoRoot, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))

	// prefer local bin/gitleaks if present
	gitleaksPath := "gitleaks"
	local := filepath.Join(repoRoot, "bin", exeName("gitleaks"))
	if fileExists(local) {
		gitleaksPath = local
	}

	configPath := filepath.Join(repoRoot, "tools", "gitleaks", "config.toml")
	args := []string{"protect", "--staged", "--redact", "--exit-code", "1", "--config", configPath}

	cmd := exec.Command(gitleaksPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err == nil {
		success("no leaks detected")
		return
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		fail("❌ gitleaks found potential secrets in staged changes. Commit rejected.")
	}
	fail("gitleaks run failed: %v", err)
}

func mustRepoRoot() string {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		fail("not a git repo? %v", err)
	}
	return strings.TrimSpace(string(out))
}

func gitConfigBool(key string, def bool) (bool, error) {
	out, err := exec.Command("git", "config", "--bool", "--get", key).Output()
	if err != nil {
		return def, nil
	}
	v := strings.TrimSpace(strings.ToLower(string(out)))
	switch v {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	default:
		return def, nil
	}
}

func binaryExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func fileExists(path string) bool {
	if st, err := os.Stat(path); err == nil && !st.IsDir() {
		return true
	}
	return false
}

func exeName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func autoInstallGitleaks(repoRoot string) error {
	// 1) brew (mac/linux)
	if binaryExists("brew") {
		if err := runSilent("brew", "install", "gitleaks"); err == nil {
			return nil
		}
	}
	// 2) apt (linux)
	if runtime.GOOS == "linux" && binaryExists("apt-get") {
		_ = run("sudo", "apt-get", "update")
		if err := run("sudo", "apt-get", "install", "-y", "gitleaks"); err == nil {
			return nil
		}
	}
	// 3) choco (windows)
	if runtime.GOOS == "windows" && binaryExists("choco") {
		if err := run("choco", "install", "-y", "gitleaks"); err == nil {
			return nil
		}
	}
	// 4) go toolchain
	if binaryExists("go") {
		if err := run("go", "install", "github.com/gitleaks/gitleaks/v8@latest"); err == nil {
			info("installed via go install — ensure $GOBIN/$GOPATH/bin is in PATH")
			return nil
		}
	}
	// 5) fallback: curl | sh to repoRoot/bin
	script := filepath.Join(repoRoot, "tools", "gitleaks", "scripts", "install_gitleaks.sh")
	if _, err := os.Stat(script); err == nil {
		env := os.Environ()
		env = append(env, "INSTALL_DIRECTORY="+filepath.Join(repoRoot, "bin"))
		cmd := exec.Command("bash", script)
		cmd.Env = env
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		}
	}
	return fmt.Errorf("auto-install methods exhausted")
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runSilent(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b
	return cmd.Run()
}

func info(msg string, a ...any)   { fmt.Fprintf(os.Stderr, "[INFO] "+msg+"\n", a...) }
func success(msg string, a ...any){ fmt.Fprintf(os.Stderr, "[OK] "+msg+"\n", a...) }
func fail(msg string, a ...any)   { fmt.Fprintf(os.Stderr, msg+"\n", a...); os.Exit(1) }
