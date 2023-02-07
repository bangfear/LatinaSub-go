package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/LalatinaHub/LatinaSub-go/blacklist"
	D "github.com/LalatinaHub/LatinaSub-go/db"
	"github.com/LalatinaHub/LatinaSub-go/helper"
	"github.com/LalatinaHub/LatinaSub-go/sandbox"
	"github.com/LalatinaHub/LatinaSub-go/scraper"

	"github.com/LalatinaHub/LatinaSub-go/subscription"
)

var (
	ch        chan int = make(chan int, 500)
	wg        sync.WaitGroup
	GoodBoxes []*sandbox.SandBox
)

func initAll() {
	D.Init()

	subscription.Init()
	blacklist.Init()
}

func Start() int {
	// Initialize all required modules
	initAll()
	db := D.New()
	db.FlushAndCreate()

	// Merge sub list
	subscription.Merge()

	// Scrape nodes from sub list
	nodes := scraper.Run()
	numNodes := len(nodes)
	for i, node := range nodes {
		fmt.Println("Testing node no", i, "/", len(nodes))
		wg.Add(1)

		ch <- 1
		go func(node string, numNodes, id int) {
			// Catch on error and done wg
			defer helper.CatchError(false)

			// Make sure wg is done and channel released
			defer func() {
				wg.Done()
				<-ch
			}()

			// Start the test
			if box := sandbox.Test(node); box != nil {
				if len(box.ConnectMode) > 0 {
					GoodBoxes = append(GoodBoxes, box)
					fmt.Printf("[%d/%d] Connected in %s mode\n", id, numNodes, strings.Join(box.ConnectMode, " and "))
				}
			}
		}(node, numNodes, i)
	}

	// Wait for all process
	wg.Wait()

	// Write all result to database
	fmt.Println("Saving result to database, please wait !")
	for _, box := range GoodBoxes {
		db.Save(box)
	}

	// Write blacklist
	blacklist.Write()

	return db.TotalAccount
}

func main() {
	start := time.Now()

	// Start the main func
	totalAccount := Start()

	fmt.Printf("\n==============================\n")
	fmt.Println("Result:", totalAccount)
	fmt.Println("Time Collapsed:", time.Since(start))

	fmt.Println("Software will exit in 10 second !")
	time.Sleep(10 * time.Second)
}
