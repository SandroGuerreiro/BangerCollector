package main

import (
	"fmt"
	"os"
	"encoding/csv"
	"regexp"
  "path/filepath"
)

func processRecords () [][]string {

	files, err :=	os.ReadDir("./Import")
	if err != nil {
		fmt.Println("Error opening directory", err)
		return [][]string{}
	}
	
	var musicInfo [][]string

	for _, file := range files {

    fileName := file.Name()
    extension := filepath.Ext(fileName)
		if (extension == ".csv") {
			musicInfo = processFile(file.Name())
		}
	}
	
	return musicInfo

}

func processFile (file string) [][]string {

	fmt.Println("Processing file - ", file)

	fd, err := os.Open("./Import/" + file)
	if err != nil {
		fmt.Println("Error opening file")
		return [][]string{}
	}
	fmt.Println("File opened successfully")
	defer fd.Close()

  fileReader := csv.NewReader(fd)

	var musicInfo [][]string

 	for {
		record, err := fileReader.Read()
		if err != nil {
			break
		}
		
		processedData := processRecord(record)
		if len(processedData) > 0 {
			musicInfo = append(musicInfo, processedData)
		}
	}

	fmt.Println("File " + file + " processed")
	return musicInfo
}

func processRecord (record []string) []string {
			
	author := record[1]
	message := record[3]
	
	if (author != "Tempo") {
		return []string{} 
	}

	pattern := `<@(\d+)>.*(\bAdded\b).*(\bby\b)`
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(message)
	if len(match) == 0 {
		return []string{}
	}

	pattern = "`(.*?)`"
	re = regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(message, -1)

	var musicInfo []string
	for _, match := range matches {
		if len(match) >= 2 {
			musicInfo = append(musicInfo, match[1])
		}
	}
		
	return musicInfo

}

