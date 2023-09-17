package main

import (
	"encoding/json"
	"fmt"

	"github.com/iafan/cwalk"

	// "github.com/kelindar/binary"
	// "io"
	"os"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

var pathArray []string

func encodeJSON(array []string, outputFile *os.File) error {
	// Create a JSON encoder.
	encoder := json.NewEncoder(outputFile)

	// Encode the array of strings to JSON.
	err := encoder.Encode(array)
	if err != nil {
		return err
	}

	// Flush the encoder.
	if err != nil {
		return err
	}

	return nil
}

func decodeJSON(inputFile *os.File) ([]string, error) {
	// Create a JSON decoder.
	decoder := json.NewDecoder(inputFile)

	// Decode the JSON data into an array of strings.
	var array []string
	err := decoder.Decode(&array)
	if err != nil {
		return nil, err
	}

	return array, nil
}

func walkFunc(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	// fmt.Println(path)
	pathArray = append(pathArray, path)
	return nil
}
func main() {

	err := cwalk.Walk("/home/pks/", walkFunc)

	fmt.Println(err)
	// words := []string{"cartwheel", "foobar", "wheel", "baz"}
	// wordx := fuzzy.Find("whl", words) // [cartwheel wheel]
	// fmt.Println(wordx)
	outputFile, err := os.Create("output.json")
	if err != nil {
		// Handle error.
	}
	defer outputFile.Close()
	err = encodeJSON(pathArray, outputFile)
	if err != nil {
		// Handle error.
	}
	inputFile, err := os.Open("output.json")
	if err != nil {
		// Handle error.
	}
	defer inputFile.Close()

	array, err := decodeJSON(inputFile)
	if err != nil {
		// Handle error.
	}
	// fmt.Println(array)
	wordx := fuzzy.Find("WhiteSur", array) // [cartwheel wheel]
	// iterate and print wordx
	for _, word := range wordx {
		fmt.Println(word)
	}
}
