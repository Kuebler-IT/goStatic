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

func main() {

    gs_default_root_dir, _ := os.Getwd()
    gs_root_dir := flag.String("d", gs_default_root_dir, "root directory for static files")
    gs_http_port := flag.String("p", "8080", "port to listen on")
    gs_http_ip := flag.String("b", "127.0.0.1", "bind server to ip")
    printUsage := flag.Bool("h", false, "print this help")
    quiet := flag.Bool("q", false, "quiet mode - only output access log")
    flag.Parse()

    if *printUsage == true {
        flag.PrintDefaults()
        os.Exit(1)
    }

    if *quiet == false {
        log.Println("Starting goStatic 1.0 - simple static webserver...")
        log.Println("Server Root Dir:", *gs_root_dir)
        log.Print("You can access the server on http://", *gs_http_ip, ":", *gs_http_port)
    }

    fs := dotFileHidingFileSystem{http.Dir(*gs_root_dir)}
    http.Handle("/", http.FileServer(fs))

    if *quiet == false {
        fmt.Println("=== access log output starts (combined log format) ===")
    }

    err := http.ListenAndServe(*gs_http_ip + ":" + *gs_http_port, handlers.CombinedLoggingHandler(os.Stdout, http.DefaultServeMux))
    if err != nil && *quiet == false {
        log.Fatal(err)
    }

}
