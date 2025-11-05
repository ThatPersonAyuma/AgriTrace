package log

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func CreateLog(text string, status string) func() error{
	return func() error{
		wd, err := os.Getwd() 
		if err != nil {
			return fmt.Errorf("failed to get working directory: %v", err)
		} 
		filePath := filepath.Join(wd, "application.log")
		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
		log.Printf("Status [%s]: %s", status, text)
		return nil
	}
}