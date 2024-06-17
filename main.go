package main

import (
	"os"
	"sync"
)

func main() {
	args := os.Args[1:]
	if args[0] == "install" || args[0] == "i" {
		packageNames := args[1:]
		if len(packageNames) == 0 {
			ReadPackageJson()
		}
		var wg sync.WaitGroup
		for _, packageName := range packageNames {
			wg.Add(1)
			go func(packageName, version string) {
				defer wg.Done()
				InstallPackage(packageName, "")
			}(packageName, "")
		}

		wg.Wait()
	}
}
