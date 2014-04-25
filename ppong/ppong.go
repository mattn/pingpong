package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
	"time"
)

var interval = flag.Uint("i", 1, "Check interval")
var timeout = flag.Int64("t", 60, "Default timeout")
var addr = flag.String("s", ":5678", "Server address")

type Task struct {
	cmd     *exec.Cmd
	Name    string   `json:"name"`
	Args    []string `json:"args"`
	Timeout int64    `json:"timeout"`
	pong    time.Time
}

var tasks = map[string]*Task{}
var re = regexp.MustCompile("^/([^/]+)/(ping|kill)$")
var cwd string
var mutex sync.Mutex

func (t *Task) Terminate() error {
	var err error
	if runtime.GOOS == "windows" {
		err = t.cmd.Process.Kill()
		if err != nil {
			return err
		}
	} else {
		err = t.cmd.Process.Signal(os.Interrupt)
		if err != nil {
			return err
		}
		err = t.cmd.Wait()
		if err != nil {
			return err
		}
	}
	return nil
}

func ping(w http.ResponseWriter, r *http.Request, name string) {
	mutex.Lock()
	defer mutex.Unlock()

	task, ok := tasks[name]
	if ok {
		task.pong = time.Now()
		w.WriteHeader(http.StatusNoContent)
		return
	}

	conf := filepath.Join(cwd, name+".json")
	f, err := os.Open(conf)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	task.cmd = exec.Command(task.Name, task.Args...)
	err = task.cmd.Start()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	task.pong = time.Now()
	if task.Timeout == 0 {
		task.Timeout = *timeout
	}
	tasks[name] = task
	log.Printf("%q started.", name)
	go func(name string, cmd *exec.Cmd) {
		cmd.Wait()
		if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			mutex.Lock()
			defer mutex.Unlock()
			if _, ok := tasks[name]; ok {
				delete(tasks, name)
				log.Printf("%q exited.", name)
			}
		}
	}(name, task.cmd)
	w.WriteHeader(http.StatusNoContent)
}

func kill(w http.ResponseWriter, r *http.Request, name string) {
	mutex.Lock()
	defer mutex.Unlock()

	task, ok := tasks[name]
	if !ok {
		http.NotFound(w, r)
		return
	}

	err := task.Terminate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("%q stopped.", name)
	w.WriteHeader(http.StatusNoContent)
}

func loggerHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func main() {
	flag.Parse()

	var err error
	cwd, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	var t *time.Timer
	t = time.AfterFunc(time.Duration(*interval)*time.Second, func() {
		mutex.Lock()
		defer mutex.Unlock()

		n := time.Now().Unix()
		for name, task := range tasks {
			if task.Timeout >= 0 && task.pong.Unix()+task.Timeout < n {
				task.Terminate()
				delete(tasks, name)
				log.Printf("%q had terminated because timeout.", name)
			}
		}
		t.Reset(time.Duration(*interval) * time.Second)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		m := re.FindStringSubmatch(r.RequestURI)
		if len(m) == 0 {
			http.NotFound(w, r)
			return
		}
		name, method := m[1], m[2]
		switch {
		case r.Method == "GET" && method == "ping":
			ping(w, r, name)
		case r.Method == "POST" && method == "kill":
			kill(w, r, name)
		}
	})
	http.ListenAndServe(*addr, loggerHandler(http.DefaultServeMux))
}
