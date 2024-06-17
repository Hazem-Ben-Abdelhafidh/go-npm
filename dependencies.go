package main

import (
	"encoding/json"
	"os"
	"strconv"
	"sync"
)

type PackageJson struct {
	Dependencies     map[string]string `json:"dependencies"`
	DevDependencies  map[string]string `json:"devDependencies"`
	PeerDependencies map[string]string `json:"peerDependencies"`
}

func ReadPackageJson() {

	data, err := os.ReadFile("package.json")
	if err != nil {
		panic("couldn't find package.json file")
	}

	var packageJson PackageJson

	err = json.Unmarshal(data, &packageJson)
	if err != nil {
		panic("couldn't unmarshal package.json, file might be badly formatted")
	}

	var wg sync.WaitGroup

	for packageName, version := range packageJson.Dependencies {
		wg.Add(1)
		go func(packageName, version string) {
			defer wg.Done()
			_, err := strconv.Atoi(string(version[0]))
			if err != nil {
				InstallPackage(packageName, version[1:])
			} else {
				InstallPackage(packageName, version)
			}
		}(packageName, version)
	}
	wg.Wait()

	var wg2 sync.WaitGroup

	for packageName, version := range packageJson.DevDependencies {
		wg2.Add(1)
		go func(packageName, version string) {
			defer wg2.Done()
			_, err := strconv.Atoi(string(version[0]))
			if err != nil {
				InstallPackage(packageName, version[1:])
			} else {
				InstallPackage(packageName, version)
			}
		}(packageName, version)
	}
	wg2.Wait()

	var wg3 sync.WaitGroup
	for packageName, version := range packageJson.PeerDependencies {
		wg3.Add(1)
		go func(packageName, version string) {
			defer wg3.Done()
			_, err := strconv.Atoi(string(version[0]))
			if err != nil {
				InstallPackage(packageName, version[1:])
			} else {
				InstallPackage(packageName, version)
			}
		}(packageName, version)
	}

	wg3.Wait()

}
