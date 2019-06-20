package main

import (
    "flag"
    "log"
    "net/http"
    "os"
    "strings"
    "github.com/gorilla/handlers"
    "fmt"
)

var VERSION string
var BUILD string
var BUILDDATE string

// containsDotFile reports whether name contains a path element starting with a period.
// The name is assumed to be a delimited by forward slashes, as guaranteed
// by the http.FileSystem interface.
func containsDotFile(name string) bool {
    parts := strings.Split(name, "/")
    for _, part := range parts {
        if strings.HasPrefix(part, ".") {
            return true
        }
    }
    return false
}

// dotFileHidingFile is the http.File use in dotFileHidingFileSystem.
// It is used to wrap the Readdir method of http.File so that we can
// remove files and directories that start with a period from its output.
type dotFileHidingFile struct {
    http.File
}

// Readdir is a wrapper around the Readdir method of the embedded File
// that filters out all files that start with a period in their name.
func (f dotFileHidingFile) Readdir(n int) (fis []os.FileInfo, err error) {
    files, err := f.File.Readdir(n)
    for _, file := range files { // Filters out the dot files
        if !strings.HasPrefix(file.Name(), ".") {
            fis = append(fis, file)
        }
    }
    return
}

// dotFileHidingFileSystem is an http.FileSystem that hides
// hidden "dot files" from being served.
type dotFileHidingFileSystem struct {
    http.FileSystem
}

// Open is a wrapper around the Open method of the embedded FileSystem
// that serves a 403 permission error when name has a file or directory
// with whose name starts with a period in its path.
func (fs dotFileHidingFileSystem) Open(name string) (http.File, error) {
    if containsDotFile(name) { // If dot file, return 403 response
        return nil, os.ErrPermission
    }

    file, err := fs.FileSystem.Open(name)
    if err != nil {
        return nil, err
    }
    return dotFileHidingFile{file}, err
}

func progInfo() {
    fmt.Print("goStatic v", VERSION," - simple static webserver\n")
    fmt.Print("Copyright (C) 2019  Benjamin Kübler <b.kuebler@kuebler-it.de>\n\n")
    fmt.Println("Build:", BUILD)
    fmt.Print("Date:  ", BUILDDATE, "\n\n")
    fmt.Println("This program is free software: you can redistribute it and/or modify")
    fmt.Println("it under the terms of the GNU General Public License as published by")
    fmt.Println("the Free Software Foundation, either version 3 of the License, or")
    fmt.Print("any later version.\n\n")
    fmt.Println("This program is distributed in the hope that it will be useful,")
    fmt.Println("but WITHOUT ANY WARRANTY; without even the implied warranty of")
    fmt.Println("MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the")
    fmt.Print("GNU General Public License for more details.\n\n")
    fmt.Println("You should have received a copy of the GNU General Public License")
    fmt.Println("along with this program.  If not, see <https://www.gnu.org/licenses/>.")
    os.Exit(1)
}

func main() {

    gs_default_root_dir, _ := os.Getwd()
    gs_root_dir := flag.String("d", gs_default_root_dir, "root directory for static files")
    gs_http_port := flag.String("p", "8080", "port to listen on")
    gs_http_ip := flag.String("b", "127.0.0.1", "bind server to ip")
    printUsage := flag.Bool("h", false, "print this help")
    printVersion := flag.Bool("v", false, "print version information")
    quiet := flag.Bool("q", false, "quiet mode - only output access log")
    flag.Parse()

    if *printUsage {
        fmt.Print("goStatic v", VERSION," - simple static webserver\n\n")
        flag.PrintDefaults()
        os.Exit(1)
    }

    if *printVersion {
        progInfo()
    }
    if !*quiet {
        log.Print("Starting goStatic v", VERSION," - simple static webserver...")
        log.Println("Server Root Dir:", *gs_root_dir)
        log.Print("You can access the server on http://", *gs_http_ip, ":", *gs_http_port)
    }

    fs := dotFileHidingFileSystem{http.Dir(*gs_root_dir)}
    http.Handle("/", http.FileServer(fs))

    if !*quiet {
        fmt.Println("=== access log output starts (combined log format) ===")
    }

    err := http.ListenAndServe(*gs_http_ip + ":" + *gs_http_port, handlers.CombinedLoggingHandler(os.Stdout, http.DefaultServeMux))
    if err != nil && !*quiet {
        log.Fatal(err)
    }

}
