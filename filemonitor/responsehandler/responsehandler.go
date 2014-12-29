package responsehandler

import (
    "fmt"
    "os"
    "bufio"
    "time"
    "strconv"
    "strings"
    "net/http"
    "regexp"
    "sort"
    "encoding/json"
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

func Handler(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path[1:]
    match, _ := regexp.MatchString("^\\d+$", path)

    if match {
      currentTime := time.Now().Unix()
      secondsFromNow, _ := strconv.ParseInt(path, 10, 64)
      maxTime := currentTime - secondsFromNow

      respondezvous := &Response{}
      respondezvous.getFilelistAndMedian(maxTime)
      marshalledjson, _ := json.Marshal(respondezvous)

      w.Header().Set("Content-Type", "application/json")
      w.Write(marshalledjson)
    }
}

func (u *Response) getFilelistAndMedian(maxTime int64){
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
