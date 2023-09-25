package filesystemsearch

import (
	"fmt"
	"path/filepath"
	"strings"
)

// function to move a file from sourcePath to destinationPath
func MoveFile(sourcePath string, destinationPath string, rootNode *Folder) error {
	segments := strings.Split(sourcePath, string(filepath.Separator))
	rootNode.removeFile(segments)
	segments = strings.Split(destinationPath, string(filepath.Separator))
	rootNode.addFile(segments)
	fmt.Printf("moved %s to %s\n", sourcePath, destinationPath)
	return nil
}

// function to delete a file at filePath
func DeleteFile(filePath string, rootNode *Folder) error {
	segments := strings.Split(filePath, string(filepath.Separator))
	rootNode.removeFile(segments)
	fmt.Printf("deleted %s\n", filePath)
	return nil
}

// function to create a file at filePath
func CreateFile(filePath string, rootNode *Folder) error {
	segments := strings.Split(filePath, string(filepath.Separator))
	rootNode.addFile(segments)
	fmt.Printf("created %s\n", filePath)
	return nil
}
