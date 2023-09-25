package main

import (
	// "bufio"
	"bufio"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"log"
	"strings"

	// "log"
	"os"
	"path/filepath"

	// "sort"
	// "strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	// "github.com/iafan/cwalk"
	// "github.com/lithammer/fuzzysearch/fuzzy"
)

type Folder struct {
	Name    string
	Files   []string
	Folders map[string]*Folder
	mu      sync.Mutex
}

var updateTreeIndex = 0
var pathsWatchable []string

func newFolder(name string) *Folder {
	return &Folder{name, []string{}, make(map[string]*Folder), sync.Mutex{}}
}

func (f *Folder) addFolder(path []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	temp := f
	for _, segment := range path {
		var temp1 *Folder
		if nextF, ok := temp.Folders[segment]; ok { // last segment == new folder
			temp1 = nextF
		} else {
			temp.Folders[segment] = newFolder(segment)
			temp1, _ = temp.Folders[segment]
		}
		temp = temp1
	}
}

func (f *Folder) addFile(path []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i, segment := range path {
		if i == len(path)-1 {
			f.Files = append(f.Files, segment)
		} else if nextF, ok := f.Folders[segment]; ok {
			f = nextF
		} else {
			f.Folders[segment] = newFolder(segment)
			f, _ = f.Folders[segment]
		}
	}

}

// p := 0
var path []string

func (f *Folder) String(init string) error {
	for _, file := range f.Files {
		path = append(path, init+string(filepath.Separator)+f.Name+string(filepath.Separator)+file)
	}
	for _, folder := range f.Folders {
		if len(f.Name) > 0 {
			folder.String(init + string(filepath.Separator) + f.Name)
		} else {
			folder.String(init)
		}
	}
	return nil
}

func Encode(root *Folder) {
	file, err := os.Create("treeNew1.bin.gz")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Create a Gzip writer
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	// Create an encoder using the gob package
	encoder := gob.NewEncoder(gzipWriter)

	// Encode the root structure into binary format
	err = encoder.Encode(root)
	if err != nil {
		fmt.Println("Error encoding to binary:", err)
		return
	}

}

func walkFunc(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	// fmt.Println(path)
	if info.IsDir() {
		pathsWatchable = append(pathsWatchable, path)
	}
	return nil
}

func updateTree() {
	file, err := os.OpenFile("FileSystemChanges.log", os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	fmt.Println("updating tree", updateTreeIndex)
	updateTreeIndex++
	// Process each line in the log file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Parse the line to extract the timestamp, action, source path, and destination path
		_, action, sourcePath, destinationPath := parseLine(line)
		// fmt.Println(action, sourcePath, destinationPath)
		// Process the line based on the action
		switch action {
		case "moved":
			err = MoveFile(sourcePath, destinationPath)
		case "deleted":
			err = DeleteFile(sourcePath)
		case "created":
			err = CreateFile(sourcePath)
		default:
			log.Printf("unknown action: %s", action)
			continue
		}

		// Delete the line from the file if it was processed successfully
		if err == nil {
			_, err = file.Seek(0, 0)
			if err != nil {
				log.Fatal(err)
			}
			err = scanner.Err()
			if err != nil {
				log.Fatal(err)
			}
			writer := bufio.NewWriter(file)
			for scanner.Scan() {
				if scanner.Text() != line {
					_, err = writer.WriteString(scanner.Text() + "\n")
					if err != nil {
						log.Fatal(err)
					}
				}
			}
			err = writer.Flush()
			if err != nil {
				log.Fatal(err)
			}
			err = file.Truncate(int64(writer.Buffered()))
			if err != nil {
				log.Fatal(err)
			}
			err = file.Sync()
			if err != nil {
				log.Fatal(err)
			}
		}
		if err != nil {
			log.Printf("error processing line: %s", err)
		}

		// Stop processing lines if the file is empty
		fileInfo, err := file.Stat()
		if err != nil {
			log.Fatal(err)
		}
		if fileInfo.Size() == 0 {
			return
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

// function to parse a log line and extract the timestamp, action, source path, and destination path
func parseLine(line string) (time.Time, string, string, string) {
	// Extract the timestamp from the line
	timestamp, err := ExtractTimestamp(line)
	if err != nil {
		log.Fatal(err)
	}

	// Extract the action from the line
	action := ""
	if strings.Contains(line, "moved") {
		action = "moved"
	} else if strings.Contains(line, "deleted") {
		action = "deleted"
	} else if strings.Contains(line, "created") {
		action = "created"
	}

	// Extract the source path from the line
	sourcePath := ""
	if strings.Contains(line, "src_path$=$[") {
		startIndex := strings.Index(line, "src_path$=$[") + len("src_path$=$[")
		endIndex := strings.Index(line[startIndex:], "]") + startIndex
		sourcePath = line[startIndex:endIndex]
	}

	// Extract the destination path from the line
	destinationPath := ""
	if strings.Contains(line, "dest_path$=$[") {
		startIndex := strings.Index(line, "dest_path$=$[") + len("dest_path$=$[")
		endIndex := strings.Index(line[startIndex:], "]") + startIndex
		destinationPath = line[startIndex:endIndex]
	}

	return timestamp, action, sourcePath, destinationPath
}

// helper function to check if a slice contains a string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
func watch() {
	// check if FileSystemChanges.log exists and if it's not zero then call the updateTree function
	fileInfo, err := os.Stat("FileSystemChanges.log")
	if err != nil {
		log.Fatal(err)
	}
	if fileInfo.Size() > 0 {
		updateTree()
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Add FileSystemChanges.log to the watcher
	err = watcher.Add("FileSystemChanges.log")
	if err != nil {
		log.Fatal(err)
	}

	// Update the tree when changes are detected
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Name == "FileSystemChanges.log" && event.Op&fsnotify.Write == fsnotify.Write {
				fileInfo, err := os.Stat("FileSystemChanges.log")
				if err != nil {
					log.Fatal(err)
				}
				if fileInfo.Size() > 0 {
					time.Sleep(1 * time.Second)
					updateTree()
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}
func main() {
	// start := time.Now()
	// rootFolder := newFolder("")
	watch()
	// visit := func(path string, info os.FileInfo, err error) error {
	// 	segments := strings.Split(path, string(filepath.Separator))
	// 	if len(segments) == 1 && segments[0] == "" {
	// 		return nil
	// 	}
	// 	if info != nil && info.IsDir() {
	// 		if len(segments) > 0 {
	// 			rootFolder.addFolder(segments)
	// 		}
	// 	} else {
	// 		rootFolder.addFile(segments)
	// 	}
	// 	return nil
	// }

	// err := cwalk.Walk("/", visit)
	// // err := filepath.Walk("/", visit)
	// if err != nil {
	// 	// log.Fatal(err)
	// 	fmt.Println(err)
	// }

	// Encode(rootFolder)
	// elapsed := time.Since(start)
	// fmt.Printf("The operation took %s\n", elapsed)

	// file1, err := os.Open("treeNew1.bin.gz")
	// if err != nil {
	// 	fmt.Println("Error opening file:", err)
	// 	return
	// }
	// defer file1.Close()

	// // Create a Gzip reader
	// gzipReader, err := gzip.NewReader(file1)
	// if err != nil {
	// 	fmt.Println("Error creating Gzip reader:", err)
	// 	return
	// }
	// defer gzipReader.Close()

	// start1 := time.Now()
	// // Create a decoder using the gob package
	// decoder := gob.NewDecoder(gzipReader)

	// // Create a root structure to decode into
	// var root1 Folder

	// // Decode the binary data into the root structure
	// err = decoder.Decode(&root1)
	// if err != nil {
	// 	fmt.Println("Error decoding binary data:", err)
	// 	return
	// }

	// root1.String("")
	// // println(root1.Name)
	// fmt.Println("***************************************************************")
	// wordx := fuzzy.RankFindFold("20CS01003", path) // [cartwheel wheel]
	// elapsed1 := time.Since(start1)
	// fmt.Printf("The operation took %s\n", elapsed1)
	// sort.Sort(wordx)
	// for _, word := range wordx {
	// 	fmt.Println(word)
	// }
}
