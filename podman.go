// Binary podman is a gokrazy wrapper program that runs the bundled podman
// executable in /usr/local/bin/podman after doing any necessary runtime system
// setup.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

const defaultDockerSocket = "/var/run/docker.sock"

func isMounted(mountpoint string) (bool, error) {
	b, err := ioutil.ReadFile("/proc/self/mountinfo")
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // platform does not have /proc/self/mountinfo, fall back to not verifying
		}
		return false, err
	}

	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		parts := strings.Fields(line)
		if len(parts) < 5 {
			continue
		}
		if parts[4] == mountpoint {
			return true, nil
		}
	}

	return false, nil
}

func makeWritable(dir string) error {
	mounted, err := isMounted(dir)
	if err != nil {
		return err
	}
	if mounted {
		// Nothing to do, directory is already mounted.
		return nil
	}

	// Read all regular files in this directory.
	regularFiles := make(map[string]string)
	fis, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		b, err := os.ReadFile(filepath.Join(dir, fi.Name()))
		if err != nil {
			return err
		}
		regularFiles[fi.Name()] = string(b)
	}

	if err := syscall.Mount("tmpfs", dir, "tmpfs", 0, ""); err != nil {
		return fmt.Errorf("tmpfs on %s: %v", dir, err)
	}

	// Write all regular files from memory back to new tmpfs.
	for name, contents := range regularFiles {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0644); err != nil {
			return err
		}
	}

	return nil
}

func runService(socketPath string) {
	uri := "unix://" + socketPath

	if err := os.MkdirAll(filepath.Dir(socketPath), 0755); err != nil {
		log.Fatalf("creating socket directory: %v", err)
	}

	if socketPath != defaultDockerSocket {
		if _, err := os.Lstat(defaultDockerSocket); err == nil {
			log.Fatal(defaultDockerSocket + " already exists, refusing to overwrite")
		}
		if err := os.Symlink(socketPath, defaultDockerSocket); err != nil {
			log.Fatalf("creating symlink %s -> %s: %v", defaultDockerSocket, socketPath, err)
		}
	}

	cmd := exec.Command("/usr/local/bin/podman", "system", "service", "--time=0", uri)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "DOCKER_HOST="+uri)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	if err := cmd.Start(); err != nil {
		log.Fatalf("starting podman service: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case sig := <-sigCh:
		cmd.Process.Signal(sig)
		<-done
		os.Exit(0)
	case err := <-done:
		if err != nil {
			log.Fatalf("podman service exited: %v", err)
		}
		os.Exit(0)
	}
}

func main() {
	socketPath := flag.String("socket", "", "Run as API daemon on given socket path; this also creates a symlink from /var/run/docker.sock")
	flag.Parse()

	// netavark invokes nft and pipes rules via /dev/stdin, which doesn't
	// exist on gokrazy's minimal /dev.
	if _, err := os.Lstat("/dev/stdin"); os.IsNotExist(err) {
		os.Symlink("/proc/self/fd/0", "/dev/stdin")
	}

	if *socketPath != "" {
		runService(*socketPath)
	} else {
		args := append([]string{"/usr/local/bin/podman"}, flag.Args()...)
		if err := syscall.Exec("/usr/local/bin/podman", args, os.Environ()); err != nil {
			log.Fatal(err)
		}
	}
}
