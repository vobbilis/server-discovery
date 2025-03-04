package main

import (
	"log"
	"sync"

	"github.com/vobbilis/codegen/server-discovery/scripts/linux"
	"github.com/vobbilis/codegen/server-discovery/scripts/windows"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	// Discover Linux servers
	go func() {
		defer wg.Done()
		log.Println("Starting Linux server discovery...")
		linux.DiscoverServers()
		log.Println("Linux server discovery completed")
	}()

	// Discover Windows servers
	go func() {
		defer wg.Done()
		log.Println("Starting Windows server discovery...")
		windows.DiscoverServers()
		log.Println("Windows server discovery completed")
	}()

	wg.Wait()
	log.Println("All server discovery completed")
}
