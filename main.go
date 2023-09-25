package main

import (
	// "bufio"
	"bufio"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

var updateTreeIndex = 0
var pathsWatchable []string

var path []string
var UpdateTreeDone = make(chan bool)

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


func updateTree(rootNode *Folder) {
	file, err := os.OpenFile("FileSystemChanges.log", os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
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
			err = MoveFile(sourcePath, destinationPath, rootNode)
		case "deleted":
			err = DeleteFile(sourcePath, rootNode)
		case "created":
			err = CreateFile(sourcePath, rootNode)
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
func watch(rootNode *Folder) {
	// check if FileSystemChanges.log exists and if it's not zero then call the updateTree function
	fileInfo, err := os.Stat("FileSystemChanges.log")
	if err != nil {
		log.Fatal(err)
	}
	if fileInfo.Size() > 0 {
		updateTree(rootNode)
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
					updateTree(rootNode)
				} else {
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
	// rootFolder := buildTree()
	// Encode(rootFolder)
	// elapsed := time.Since(start)
	// fmt.Printf("The operation took %s\n", elapsed)
	file1, err := os.Open("treeNew1.bin.gz")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file1.Close()
	// Create a Gzip reader
	gzipReader, err := gzip.NewReader(file1)
	if err != nil {
		fmt.Println("Error creating Gzip reader:", err)
		return
	}
	defer gzipReader.Close()
	decoder := gob.NewDecoder(gzipReader)
	var root1 Folder
	err = decoder.Decode(&root1)
	if err != nil {
		fmt.Println("Error decoding binary data:", err)
		return
	}
	go watch(&root1)
	// root1.PrintTree("")
	// fmt.Println(UpdateTreeDone)
	for {
		root1.String("")
		fmt.Print("Enter the term to search for (type 'exit' to quit): ")
		var searchTerm string
		fmt.Scanln(&searchTerm)
		searchTerm = strings.TrimSpace(searchTerm)
		if searchTerm == "exit" {
			break
		}
		wordx := fuzzy.RankFindFold(searchTerm, path)
		sort.Sort(wordx)
		// fmt.Println(wordx)
		// limit upto 10 results in wordx
		fmt.Println(len(path), " total paths available")
		if len(wordx) > 50 {
			wordx = wordx[:50]
		}
		for _, word := range wordx {
			fmt.Println(word)
		}
		fmt.Println("***************************************************************")
		path = []string{}
	}
	Encode(&root1)
	fmt.Println("Exiting...")
}
