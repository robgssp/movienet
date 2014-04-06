package main

import (
	"fmt"
	"net"
//	"bufio"
//	"html/template"
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
	ln, err := net.Listen("tcp", ":3425")
	if err != nil {
		fmt.Println("Nope.")
		return
	}

	fmt.Println("Listening.")
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
	// rd := bufio.NewReader(c)
	// wr := bufio.NewWriter(c)
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Connection to server failed: %s\n", err)
			return
		}
	}()
	
	srv := readServer(c)
	

	fmt.Printf("Received %s\n", srv)
	c.Close()
}

func readServer(c net.Conn) Server {
	dec := json.NewDecoder(c)
	var res map[string]interface{}

	if err := dec.Decode(&res); err != nil {
		fmt.Printf("Dun goof: %s\n", err)
		panic("asdf")
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

