package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/samverrall/polybin/config"
)

const (
	Watch = "watch"
	Run   = "run"
)

func Polybin(userConfig *config.Config, project string) error {
	foundProject := userConfig.FindProjectByName(project)
	switch {
	case foundProject == nil:
		return fmt.Errorf("-> supplied project does not exist in config: %s", project)

	case len(foundProject.Services) == 0:
		return fmt.Errorf("-> supplied project has no services: %s", project)
	}

	for _, service := range foundProject.Services {
		switch service.Type {
		case Watch:
			go watchBinary(service.Dir, *service.Binary, service.Args...)

		case Run:
			go run(service.Dir, service.Args[0], service.Args[1:]...)
		}
	}

	return nil
}

func run(dir, name string, args ...string) *exec.Cmd {
	if strings.HasSuffix(name, ".bat") || strings.HasSuffix(name, ".exe") {
		exeName := strings.Replace(name, ".bat", ".exe", -1)

		fmt.Printf("-> Ending existing processes for %q...\n", exeName)

		for _, n := range []string{exeName, exeName + "~"} {
			killProcressByName(n)
		}

		name = filepath.Join(dir, name)

		time.Sleep(1 * time.Second)
	}

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Println(err)
	}

	return cmd
}

func watchBinary(dir, name string, args ...string) {
	src := filepath.Join(dir, name)

	copyAndRun(dir, src, args...)

	var lastModified time.Time
	for {
		fi, err := os.Stat(src)
		if err != nil {
			log.Println(err)
			continue
		}

		if !lastModified.IsZero() && fi.ModTime() != lastModified {
			copyAndRun(dir, src, args...)
		}
		lastModified = fi.ModTime()

		time.Sleep(1 * time.Second)
	}
}

// killProcressByName attempts to kill a process by using the appropriate kill command
// for the runtime system OS.
func killProcressByName(name string) {
	var err error
	switch {
	case strings.Contains(runtime.GOOS, "windows"):
		err = exec.Command("taskkill", "/im", name, "/f").Run()

	default: // linux, mac etc
		err = exec.Command("pkill", "-9", name).Run()
	}

	if err != nil {
		fmt.Printf("-> Failed to kill process for %q, got error: %s\n", name, err.Error())
	}
}

func copyAndRun(dir, srcBin string, args ...string) *exec.Cmd {
	originalName := filepath.Base(srcBin)

	fmt.Printf("-> Ending existing processes for %q...\n", originalName)

	for _, n := range []string{originalName} {
		killProcressByName(n)
	}

	fmt.Printf("-> Starting %q...\n", originalName)

	cmd := exec.Command(args[0], args[1:]...) // #nosec G204
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Println(err)
	}

	return cmd
}
