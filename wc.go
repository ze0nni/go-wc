package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/karrick/godirwalk"
)

const asciiMapSize = 128

type asciiMap = [asciiMapSize]int64

func main() {
	var root = "."
	if len(os.Args) >= 2 {
		root = os.Args[1]
	}

	result := run(
		runtime.NumCPU(),
		scanDir(root),
	)

	for i := 0; i < asciiMapSize; i++ {
		amount := result[i]
		if amount > 0 {
			fmt.Printf("%d %d\n", i, amount)
		}
	}
}

func run(numGorutines int, files <-chan string) asciiMap {
	var wg = sync.WaitGroup{}

	asciiMapChan := make(chan asciiMap)

	// Сканирование файлов в несколько потоков
	for i := 0; i < numGorutines; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			task(files, asciiMapChan)
		}()
	}

	scanDirDone := make(chan struct{})
	reduceDone := make(chan struct{})

	var reducedMap asciiMap

	// Объединение результатов сканирования других файлов
	go func() {
		for {
			select {
			case m := <-asciiMapChan:
				for i := 0; i < asciiMapSize; i++ {
					reducedMap[i] += m[i]
				}
			case <-scanDirDone:
				reduceDone <- struct{}{}
				return
			}
		}
	}()

	// Ждем пока просканируются все файлы
	wg.Wait()
	scanDirDone <- struct{}{}

	// Ждем пока результаты сканирования файлов объединяться
	<-reduceDone

	return reducedMap
}

func task(files <-chan string, results chan<- asciiMap) {
	for f := range files {
		result, err := scanFile(f)
		if nil != err {
			log.Panic(err)
		}
		results <- result
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
			if b < asciiMapSize {
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
