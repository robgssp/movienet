package main

import (
    "fmt"
    "net/http"
    "io/ioutil"
    "os"
    "strconv"
    "strings"
    "io"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Printf("New %s request to %s:\n", r.Method, r.URL.Path)
    headers := r.Header
    //for h := range(headers) {
    //    fmt.Printf("%s: %s\n", h, headers[h])
    //}

    if r.URL.Path == "/test.webm"  && r.Method == "GET" {
        stats, err := os.Stat("test3.webm")
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
        //fmt.Sscanf(headers["Range"][0], "bytes=%d-", startLoc)

        fmt.Printf("Start: %d\nEnd: %d\n", startLoc, endLoc)

        w.Header()["Accept-Ranges"] = []string{"bytes"}
        w.Header()["Content-Length"] = []string{strconv.Itoa(int(stats.Size() - startLoc))}
        w.Header()["Content-Range"] = []string{"bytes " + strconv.Itoa(int(startLoc)) + "-" + strconv.Itoa(int(endLoc)) + "/" + strconv.Itoa(int(endLoc) + 1)}

        file, err := os.Open("test3.webm")
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
        //bufLen := 16
        //buf := make([]byte, bufLen)
        //n, err := file.ReadAt(buf, startLoc)
        //for err == nil {
        //    w.Write(buf[0:n])
        //    //fmt.Fprintf(w, "%s", buf[0:n])
        //    n, err = file.Read(buf)
        //}
    }
}

func servePage(w http.ResponseWriter, r *http.Request) {
    body, err := ioutil.ReadFile("mn.html")
    if err != nil {
        fmt.Fprintf(w, "error!")
    } else {
        fmt.Fprintf(w, "%s", body)
    }
}

func servePage2(w http.ResponseWriter, r *http.Request) {
    body, err := ioutil.ReadFile("video-js.css")
    if err != nil {
        fmt.Fprintf(w, "error!")
    } else {
        fmt.Fprintf(w, "%s", body)
    }
}

func servePage3(w http.ResponseWriter, r *http.Request) {
    body, err := ioutil.ReadFile("video.js")
    if err != nil {
        fmt.Fprintf(w, "error!")
    } else {
        fmt.Fprintf(w, "%s", body)
    }
}

func main() {
    http.HandleFunc("/", handler)
    http.HandleFunc("/mn.html", servePage)
    http.HandleFunc("/video-js.css", servePage2)
    http.HandleFunc("/video.js", servePage3)
    fmt.Println("Listening on port 8080 for requests")
    http.ListenAndServe(":8080", nil)
}
