package main

import (
//	"os"
	"io"
	"fmt"
	"net"
	"bytes"
	"net/http"
	"html/template"
	"encoding/json"
)

type Server struct {
	name string
	conn net.Conn
	files Dir
}

type File struct {
	name string
	id int
	parent *Dir
}

type Dir struct {
	name string
	children []interface{}
	parent *Dir
}

func main() {
	fmt.Println("Listening.")
	
	go srvListen()
	httpListen()
}

var srvs []*Server = make([]*Server, 0)

// HTTP Frontend stuff

func httpListen() {
	http.HandleFunc("/", mainMenu)

	http.ListenAndServe(":3424", nil)
}

func mainMenu(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("foo").Parse(`<html><body><h1>Bootleg</h1>{{.}}</body></html>`)
	if err != nil { panic(err); }

	buf := bytes.Buffer{}

	type ServerArg struct {
		Name string
		Body template.HTML
	}

	for _, srv := range srvs {
		t1, err1 := template.New("bar").Parse(`<h2>{{.Name}}</h2><ul>{{.Body}}</ul>`)
		if err1 != nil { panic(err1); }
		
		buf1 := bytes.Buffer{}
		formatDir(srv.files, &buf1)
		t1.Execute(&buf, ServerArg{srv.name, template.HTML(buf1.String())})
	}

	t.Execute(w, template.HTML(buf.String()))
}

func formatDir(d Dir, b io.Writer) {
	type DirArg struct {
		Name string
		Files template.HTML
	}

	t, err := template.New("foo").Parse(`<li>{{.Name}}</li><ul>{{.Files}}</ul></li>`)
	if err != nil {
		fmt.Printf("formatDir err: %s\n", err)
		panic(err)
	}

	buf := bytes.Buffer{}

	for _, fd := range d.children {
		switch fd.(type) {
		case Dir:
			formatDir(fd.(Dir), &buf)
		case File:
			formatFile(fd.(File), &buf)
		}
	}

	t.Execute(b, DirArg{d.name, template.HTML(buf.String())})
}

func formatFile(f File, b io.Writer) {
	type FileArg struct {
		Id int
		Name string
	}
	
	t, err := template.New("foo").Parse(`<li><a href="/id/{{.Id}}">{{.Name}}</a></li>`)
	if err != nil {
		panic(err)
	}

	t.Execute(b, FileArg{f.id, f.name})
}

// Server-manager stuff

func srvListen() {
	ln, err := net.Listen("tcp", ":3425")
	if err != nil {
		fmt.Println("Nope.")
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Failure to connect.")
			continue
		}
		go server(conn)
	}
}

func server(c net.Conn) {
	dec := json.NewDecoder(c)

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Connection to server failed: %s\n", err)
			return
		}
	}()
	
	srv := readServer(c, dec)

	srvs = append(srvs, &srv)
	fmt.Printf("Received %s\n", srv)

	var res map[string]interface{}
	for {
		if err := dec.Decode(&res); err != nil {
			fmt.Printf("Server comms failed: %s\n", err)
			break
		}
		switch res["type"] {
		case "add":
			// TODO handle add
		case "remove":
			// TODO handle remove
		}
	}
}

func readServer(c net.Conn, dec *json.Decoder) Server {
	var res map[string]interface{}

	if err := dec.Decode(&res); err != nil {
		fmt.Printf("Dun goof: %s\n", err)
		panic(err)
	}

	return Server{res["name"].(string),  c, readFiles(res["tree"].(map[string]interface{}), nil).(Dir)}
}

func readFiles(jf map[string]interface{}, parent *Dir) interface{} {
	switch jf["type"] {
	case "dir":
		ret := Dir{jf["name"].(string), []interface{}{}, parent}

		children := []interface{}{}
		for _, v := range jf["children"].([]interface{}) {
			children = append(children, readFiles(v.(map[string]interface{}), &ret))
		}
		ret.children = children
		return ret
	case "file":
		return File{jf["name"].(string), int(jf["id"].(float64)), parent}
	default:
		panic("Invalid type")
	}
}
