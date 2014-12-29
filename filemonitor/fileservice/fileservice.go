package fileservice

import (
    "fmt"
    "log"
    "os"
    "time"
    "strconv"
    "strings"
    "regexp"
    "github.com/emmasteimann/fsmonitor"
)

const logfile string = "/tmp/foo/mylog.log"
const watchdirectory string = "/tmp/foo"

func writeLine(line string, path string) error {
  file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
  if err != nil {
    return err
  }
  defer file.Close()

  if _, err = file.WriteString(strings.TrimSpace(line) + "\n"); err != nil {
    panic(err)
  }

  return err
}

func createLogFile() {
  if _, err := os.Stat(logfile); os.IsNotExist(err) {
      fmt.Printf("creating a log file: %s", logfile)
      f, err := os.Create(logfile)

      if err != nil {
        log.Fatal(err)
      }

      defer f.Close()
      return
  }
}

func NewWatcher() {
  createLogFile()

  watcher, err := fsmonitor.NewWatcher()

    if err != nil {
        log.Fatal(err)
    }

    err = watcher.Watch(watchdirectory)

    if err != nil {
        log.Fatal(err)
    }

    for {
        select {
        case event := <-watcher.Event:

            if f, err := os.Stat(event.Name); err == nil {
              match, _ := regexp.MatchString("^_(.*)$", f.Name())

              if event.IsCreate() && match {
                timestamp := strconv.FormatInt(time.Now().Unix(), 10)
                logstring := timestamp + "," + event.Name + "," + f.Name()

                if err := writeLine(logstring, logfile); err != nil {
                  log.Fatalf("writeLine: %s", err)
                }
              }
            }

            fmt.Println("event:", event.Name)

        case err := <-watcher.Error:

            fmt.Println("error:", err)
        }
    }
}
