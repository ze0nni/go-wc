package main

import (
	"fmt"
	"log"
	"os"

	"github.com/karrick/godirwalk"
)

func main() {
	for f := range scanDir(".") {
		fmt.Printf("%s\n", f)
	}
}

func scanDir(root string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		godirwalk.Walk(root, &godirwalk.Options{
			Unsorted: true,
			Callback: func(osPathName string, do *godirwalk.Dirent) error {
				fileinfo, err := os.Stat(osPathName)
				if nil != err {
					return err
				}
				if false == fileinfo.IsDir() {
					out <- osPathName
				}

				return nil
			},
			ErrorCallback: func(message string, err error) godirwalk.ErrorAction {
				log.Panic(err)
				return godirwalk.Halt
			},
		})

	}()

	return out
}
