package main

import (
	"os"
)

func main() {
	args := os.Args[1:]
	if args[0] == "install" || args[0] == "i" {
		packageName := args[1]
		if packageName == "" {
			// TODO: install all packages inside the package.json file
			panic("no pacakge specified !")
		}

		InstallPackage(packageName, "")

	}
}
