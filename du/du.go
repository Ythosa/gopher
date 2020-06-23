package du

import (
	"flag"
	"fmt"
	"time"
)

var verbose = flag.Bool("v", false, "output of intermediate results")

// Du reports the amount of disk space used by one or more directories
// as a `du` Unix command
func Du() {
	// Defines the initial directories
	flag.Parse()
	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string{"."}
	}

	// Crawling the file tree
	fileSizes := make(chan int64)
	go func() {
		for _, root := range roots {
			walkDir(root, fileSizes)
		}
		close(fileSizes)
	}()

	// Periodic output of results
	var tick <-chan	time.Time
	if *verbose {
		tick = time.Tick(500 * time.Millisecond)
	}

	// Counting results
	var nfiles, nbytes int64
	counting := true
	for counting {
		select {
			case size, ok := <-fileSizes:
				if !ok {
					counting = false
				}
				nfiles++
				nbytes += size

			case <-tick:
				printDiskUsage(nfiles, nbytes)
		}
	}

	// Printing results
	printDiskUsage(nfiles, nbytes)
}

func printDiskUsage(nfiles, nbytes int64) {
	fmt.Printf("%d files %.5f GB\n", nfiles, float64(nbytes) / 1e9)
}
