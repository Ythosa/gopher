package du

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"
)

var verbose = flag.Bool("v", false, "output of intermediate results")

var done = make(chan struct{})

// Du reports the amount of disk space used by one or more directories
// as a `du` Unix command
func Du() {
	// Defines the initial directories
	flag.Parse()
	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string{"."}
	}

	// Canceling a crawl when an input is detected
	go func() {
		os.Stdin.Read(make([]byte, 1)) // Reading 1 byte
		close(done)
	}()

	// Periodic output of results
	var tick <-chan time.Time
	if *verbose {
		tick = time.Tick(500 * time.Millisecond)
	}

	// Crawling the file tree
	fileSizes := make(chan int64)
	var n sync.WaitGroup
	for _, root := range roots {
		n.Add(1)
		go walkDir(root, &n, fileSizes)
	}

	// Waiting all goroutines
	go func() {
		n.Wait()
		close(fileSizes)
	}()

	// Counting results
	var nfiles, nbytes int64
	counting := true
	for counting {
		select {
		case <-done:
			// emptying the channel to allow all goroutines to end
			for range fileSizes {
				// Do nothing
			}
			return

		case size, ok := <-fileSizes:
			if !ok {
				counting = false // fileSizes is closed
			}
			nfiles++
			nbytes += size

		case <-tick:
			printDiskUsage(nfiles, nbytes)
		}
	}

	// Printing results
	fmt.Print("Total: ")
	printDiskUsage(nfiles, nbytes)
}

func printDiskUsage(nfiles, nbytes int64) {
	fmt.Printf("%d files %.5f GB\n", nfiles, float64(nbytes)/1e9)
}
