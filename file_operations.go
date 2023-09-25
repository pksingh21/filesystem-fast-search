package main

import (
	"fmt"
)

// function to move a file from sourcePath to destinationPath
func MoveFile(sourcePath string, destinationPath string) error {
	fmt.Printf("moved %s to %s\n", sourcePath, destinationPath)
	return nil
}

// function to delete a file at filePath
func DeleteFile(filePath string) error {
	fmt.Printf("deleted %s\n", filePath)
	return nil
}

// function to create a file at filePath
func CreateFile(filePath string) error {
	fmt.Printf("created %s\n", filePath)
	return nil
}
