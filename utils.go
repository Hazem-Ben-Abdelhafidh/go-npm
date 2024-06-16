package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

const npmRegistry = "https://registry.npmjs.org/"

type DistTags struct {
	Latest string `json:"latest"`
}

type Version struct {
	Dist         Dist              `json:"dist"`
	Dependencies map[string]string `json:"dependencies"`
}

type Dist struct {
	Tarball   string `json:"tarball"`
	Integrity string `json:"integrity"`
}

type Package struct {
	Id       string   `json:"_id"`
	DistTags DistTags `json:"dist-tags"`
	Versions map[string]Version
}

// ExtractTarGz extracts a tar.gz file to the specified destination directory
func ExtractTarGz(gzipStream io.Reader, dest string) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return fmt.Errorf("gzip.NewReader: %w", err)
	}
	defer uncompressedStream.Close()

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()

		switch {
		case err == io.EOF:
			return nil // End of tar archive
		case err != nil:
			return fmt.Errorf("tarReader.Next: %w", err)
		case header == nil:
			continue
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(target, 0777); err != nil {
				return fmt.Errorf("os.MkdirAll: %w", err)
			}
		case tar.TypeReg:
			// Create file
			if err := os.MkdirAll(filepath.Dir(target), 0777); err != nil {
				return fmt.Errorf("os.MkdirAll: %w", err)
			}

			outFile, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("os.Create: %w", err)
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("io.Copy: %w", err)
			}
		default:
			return fmt.Errorf("unsupported type: %v in %s", header.Typeflag, header.Name)
		}
	}
}

func InstallPackage(packageName, version string) {
	response, err := http.Get(npmRegistry + packageName)
	if err != nil {
		panic("error while installing " + packageName)
	}

	body := response.Body
	defer body.Close()

	var packageData Package
	decoder := json.NewDecoder(body)
	err = decoder.Decode(&packageData)
	if err != nil {
		fmt.Println("error while unmarshaling response", err)
		panic("error while unmarshaling response")
	}

	fmt.Printf("%+v \n", packageData.Versions[packageData.DistTags.Latest].Dist.Tarball)
	fmt.Printf("%+v \n", packageData.Versions)
	if version == "" {
		version = packageData.DistTags.Latest

	}

	latestVersion := packageData.Versions[version]
	response, err = http.Get(latestVersion.Dist.Tarball)
	if err != nil {
		panic("error while fetching tarball")
	}

	body = response.Body

	err = ExtractTarGz(body, packageData.Id)
	if err != nil {
		fmt.Println("error while extracting tarball", err)
		panic("error while extracting tarball")
	}
	fmt.Println(packageData.Versions[version].Dist.Integrity)
	InstallDependencies(packageData.Versions[version].Dependencies)
}

func InstallDependencies(dependencies map[string]string) {
	var wg sync.WaitGroup
	for dependencyName, version := range dependencies {
		wg.Add(1)
		go func(dependencyName, version string) {
			defer wg.Done()
			InstallPackage(dependencyName, version[1:])
		}(dependencyName, version)
	}
	wg.Wait()
}
