package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
)

func main() {
	filename := "text.txt"
	content := uuid.New().String()

	for {
		var f *os.File
		if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
			fmt.Println("Create a File")
			f, err = os.Create(filename)
			if err != nil {
				log.Fatal(err)
			}
		}
		defer f.Close()
		os.WriteFile(filename, []byte(content), 0644)
		content += "\n" + uuid.New().String()
		time.Sleep(5 * time.Second)
	}
}
