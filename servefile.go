package main

import (
    "fmt"
    "net/http"
    "io/ioutil"
    "os"
    "strconv"
    "strings"
    "io"
    "code.google.com/p/go.net/websocket"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Printf("New %s request to %s:\n", r.Method, r.URL.Path)
    headers := r.Header
    //for h := range(headers) {
    //    fmt.Printf("%s: %s\n", h, headers[h])
    //}

    if r.URL.Path[len(r.URL.Path) - 5:] == ".webm"  && r.Method == "GET" {
        filepath := r.URL.Path[1:]
        stats, err := os.Stat(filepath)
        if err != nil {
            fmt.Println("Error getting file stats")
            return
        }

        var startLoc, endLoc int64
        endLoc = stats.Size() - 1
        rangeString := headers["Range"][0]
        rangeString = rangeString[6:]
        intHolder, _ := strconv.Atoi(rangeString[0:strings.Index(rangeString, "-")])
        startLoc = int64(intHolder)

        fmt.Printf("Start: %d\nEnd: %d\n", startLoc, endLoc)

        w.Header()["Accept-Ranges"] = []string{"bytes"}
        w.Header()["Content-Length"] = []string{strconv.Itoa(int(stats.Size() - startLoc))}
        w.Header()["Content-Range"] = []string{"bytes " + strconv.Itoa(int(startLoc)) + "-" + strconv.Itoa(int(endLoc)) + "/" + strconv.Itoa(int(endLoc) + 1)}

        file, err := os.Open(filepath)
        if err != nil {
            fmt.Println("Error reading video file")
            return
        }
        defer file.Close()

        w.WriteHeader(206)

        _, err = file.Seek(startLoc, 0)
        if err != nil {
            fmt.Println("Error seeking")
        }
        n, err := io.Copy(w, file)
        if err != nil {
            fmt.Printf("Error copying. %d bytes written\n", n)
            return
        }
    } else {
        path := r.URL.Path[1:]
        fmt.Println("Giving the file " + path)
        body, err := ioutil.ReadFile(path)
        if err != nil {
            fmt.Fprintf(w, "error!")
        } else {
            fmt.Fprintf(w, "%s", body)
        }
    }
}

func WebSocket(ws *websocket.Conn) {
    io.Copy(ws, ws)
}

func main() {
    http.HandleFunc("/", handler)
    http.Handle("/ws", websocket.Handler(WebSocket))
    fmt.Println("Listening on port 8080 for requests")
    http.ListenAndServe(":8080", nil)
}
