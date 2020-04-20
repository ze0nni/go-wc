package main

import (
	"io"
	"log"
	"os"
	"sync"

	"github.com/karrick/godirwalk"
)

const asciiMapSize = 127

type asciiMap = [asciiMapSize]int64

func main() {
	run(4, scanDir("."))
}

func run(numGorutines int, files <-chan string) {
	var wg = sync.WaitGroup{}

	for i := 0; i < numGorutines; i++ {
		wg.Add(1)

		go func() {
			task(files)
			defer wg.Done()
		}()
	}
	wg.Wait()
}

func task(files <-chan string) {
	for f := range files {
		_, err := scanFile(f)
		if nil != err {
			log.Panic(err)
		}
	}
}

// Тут asciiMap передается через стек, что, думаю, совсем не плохо
func scanFile(fileName string) (asciiMap, error) {
	var asciiMap asciiMap

	file, err := os.Open(fileName)
	if nil != err {
		return asciiMap, err
	}

	defer file.Close()

	buffer := make([]byte, 1024*60)

ReadLoop:
	for {
		readed, err := file.Read(buffer)
		if err != nil {
			if err != io.EOF {
				return asciiMap, err
			}
			break ReadLoop
		}

		for i := 0; i < readed; i++ {
			b := buffer[i]
			if b < 127 {
				asciiMap[b]++
			}
		}
	}

	return asciiMap, nil
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

func run(numGorutines int, files <-chan string) {
	var wg = sync.WaitGroup{}

	for i := 0; i < numGorutines; i++ {
		wg.Add(1)

		go func() {
			task(files)
			defer wg.Done()
		}()
	}
	wg.Wait()
}

func task(files <-chan string) {
	for range files {

	}
}
