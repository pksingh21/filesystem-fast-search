package main

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/iafan/cwalk"
)

type Folder struct {
	Name    string
	Files   []string
	Folders map[string]*Folder
	mu      sync.Mutex
}

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
func (f *Folder) removeFile(path []string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	temp := f
	for i, segment := range path {
		if i == len(path)-1 {
			for j, file := range temp.Files {
				if file == segment {
					temp.Files = append(temp.Files[:j], temp.Files[j+1:]...)
					return true
				}
			}
			return false
		} else {
			if nextF, ok := temp.Folders[segment]; ok {
				temp = nextF
			} else {
				return false
			}
		}
	}
	return false
}

func (f *Folder) addFile(path []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i, segment := range path {
		if i == len(path)-1 {
			// fmt.Println("in add file function with path", segment, " appended successfully ")
			f.Files = append(f.Files, segment)
		} else if nextF, ok := f.Folders[segment]; ok {
			// fmt.Println("next F : ", segment, " searching  ")
			f = nextF
		} else {
			// fmt.Println("creating new folder ", segment, "   ")
			f.Folders[segment] = newFolder(segment)
			f, _ = f.Folders[segment]
		}
	}
}

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
func (f *Folder) PrintTree(indent string) {
	fmt.Println(indent + f.Name + "/")
	for _, file := range f.Files {
		fmt.Println(indent + "  - " + file)
	}
	for _, folder := range f.Folders {
		folder.PrintTree(indent + "  ")
	}
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

	encoder := gob.NewEncoder(gzipWriter)
	err = encoder.Encode(root)
	if err != nil {
		fmt.Println("Error encoding binary data:", err)
		return
	}
}

func buildTree() *Folder {
	rootFolder := newFolder("")
	visit := func(path string, info os.FileInfo, err error) error {
		segments := strings.Split(path, string(filepath.Separator))
		if len(segments) == 1 && segments[0] == "" {
			return nil
		}
		if path[0] == '.' {
			return nil
		}
		if strings.Contains(path, "White") {
			return nil
		}
		if info != nil && info.IsDir() {
			if len(segments) > 0 {
				rootFolder.addFolder(segments)
			}
		} else {
			rootFolder.addFile(segments)
		}
		return nil
	}

	err := cwalk.Walk("/", visit)
	if err != nil {
		fmt.Println(err)
	}

	Encode(rootFolder)
	return rootFolder
}
