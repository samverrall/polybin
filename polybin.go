package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func main() {
	var service string
	flag.StringVar(&service, "service", "", "The name of the group of processes to run")
	flag.Parse()

	if service == "" {
		log.Fatal("The -service flag must be provided")
	}

	switch service {
	case "service:tht":
		go watchBinary("/Users/samverrall/projects/tht-api", "./the_horse_trust", "./run.sh")

	case "service:auth-api-v2":
		go watchBinary("/Users/samverrall/projects/auth-api-v2", "./auth-api-v2", "./run.sh")

	default:
		fmt.Printf("Unsupported project supplied: %s", service)
		os.Exit(1)
	}

	stopChan := make(chan os.Signal, 2)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	<-stopChan
}

func run(dir, name string, args ...string) *exec.Cmd {
	if strings.HasSuffix(name, ".bat") || strings.HasSuffix(name, ".exe") {
		exeName := strings.Replace(name, ".bat", ".exe", -1)

		fmt.Printf("-> Ending existing processes for %q...\n", exeName)

		for _, n := range []string{exeName, exeName + "~"} {
			exec.Command("pkill", n).Run()
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
	newName := fmt.Sprintf("%s_polybin", name)
	src := filepath.Join(dir, name)
	dst := filepath.Join(dir, newName)

	copyAndRun(dir, src, dst, args...)

	var lastModified time.Time
	for {
		fi, err := os.Stat(src)
		if err != nil {
			log.Println(err)
			continue
		}

		if !lastModified.IsZero() && fi.ModTime() != lastModified {
			copyAndRun(dir, src, newName, args...)
		}
		lastModified = fi.ModTime()

		time.Sleep(1 * time.Second)
	}
}

func copyAndRun(dir, srcBin, dstBin string, args ...string) *exec.Cmd {
	originalName := filepath.Base(srcBin)
	newName := filepath.Base(dstBin)

	fmt.Printf("-> Ending existing processes for %q...\n", originalName)

	for _, n := range []string{originalName, newName, newName + "~"} {
		cmd := exec.Command("pkill", "-9", n)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("-> Failed to kill process for %q, got error: %s\n", n, err.Error())
		}
	}

	fmt.Printf("-> Copying %q as %q...\n", originalName, newName)
	time.Sleep(1 * time.Second)

	err := func() error {
		src, err := os.Open(srcBin)
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := os.OpenFile(dstBin, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			return err
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return err
		}

		return nil
	}()
	if err != nil {
		log.Println(err)
		return nil
	}

	fmt.Printf("-> Starting %q...\n", newName)
	time.Sleep(1 * time.Second)

	cmd := exec.Command(args[0])
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Println(err)
	}

	return cmd
}
