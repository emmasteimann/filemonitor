package main

import (
    "fmt"
    "log"
    "os"
    "bufio"
    "time"
    "strconv"
    "strings"
    "net/http"
    "regexp"
    "sort"
    "encoding/json"
    "github.com/emmasteimann/fsmonitor"
)

const logfile string = "/tmp/foo/mylog.log"

type Response struct {
    Files []string `json:"files"`
    MedianLength int `json:"medianlength"`
}

type ByLength []string
func (s ByLength) Len() int {
    return len(s)
}
func (s ByLength) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}
func (s ByLength) Less(i, j int) bool {
    return len(s[i]) < len(s[j])
}

func handler(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path[1:]
    match, _ := regexp.MatchString("^\\d+$", path)
    if match {
      currentTime := time.Now().Unix()
      secondsFromNow, _ := strconv.ParseInt(path, 10, 64)
      maxTime := currentTime - secondsFromNow
      respondezvous := &Response{}
      respondezvous.getFilesWithMaxTime(maxTime)
      fmt.Println(respondezvous)
      marshalledjson, _ := json.Marshal(respondezvous)
      w.Header().Set("Content-Type", "application/json")
      fmt.Println(marshalledjson)
      w.Write(marshalledjson)
    }
}

func (u *Response) getFilesWithMaxTime(maxTime int64){
  lines, err := readLines(logfile)
  if err != nil {
      fmt.Println("Error: %s\n", err)
      return
  }
  fullpathlist := []string{}
  filenames := []string{}
  for _, line := range lines {
    s := strings.Split(line, ",")
    timestamp, path, name := s[0], s[1], s[2]
    timefromline, _ := strconv.ParseInt(timestamp, 10, 64)
    if timefromline >= maxTime {
      fullpathlist = append(fullpathlist, path)
      filenames = append(filenames, name)
    }
  }
  filenameslength := len(filenames)
  medianlength := 0
  if filenameslength > 0 {
    sort.Sort(ByLength(filenames))
    median := filenameslength / 2
    medianlength = len(filenames[median])
  }
  u.Files = fullpathlist
  u.MedianLength = medianlength
  return
}

func readLines(path string) ([]string, error) {
  file, err := os.Open(path)
  if err != nil {
    return nil, err
  }
  defer file.Close()

  var lines []string
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    lines = append(lines, scanner.Text())
  }
  return lines, scanner.Err()
}

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

func CreateLogFile() {
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

func newWatcher() {
  watcher, err := fsmonitor.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    err = watcher.Watch("/tmp/foo")
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

func main() {
    CreateLogFile()
    http.HandleFunc("/", handler)
    go func() {
      http.ListenAndServe(":8888", nil)
    }()
    newWatcher()
}
