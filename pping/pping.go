package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"
)

var interval = flag.Uint("p", 10, "Ping interval")
var name = flag.String("n", "", "Task name")
var server = flag.String("s", "http://127.0.0.1:5678", "Ping server")

func ping() {
	res, err := http.Get(*server + "/" + *name + "/ping")
	if err == nil {
		res.Body.Close()
	} else {
		log.Println(err)
	}
}

func main() {
	flag.Parse()

	if *name == "" {
		fmt.Fprintln(os.Stderr, "Task name must be specified with '-n'.")
		os.Exit(1)
	}

	var t *time.Timer
	t = time.AfterFunc(time.Duration(*interval) * time.Second, func() {
		ping()
		t.Reset(time.Duration(*interval) * time.Second)
	})
	ping()

	if flag.NArg() == 0 {
		return
	}

	cmd := exec.Command(flag.Arg(0), flag.Args()[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		if cmd.ProcessState != nil {
			if status, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		os.Exit(1)
	}
}
