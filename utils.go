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
	"strconv"
	"strings"
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

		// Strip the first directory from the header name
		parts := strings.SplitN(header.Name, string(filepath.Separator), 2)
		var target string
		if len(parts) > 1 {
			target = filepath.Join(dest, parts[1])
		} else {
			target = filepath.Join(dest, parts[0])
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0777); err != nil {
				return fmt.Errorf("os.MkdirAll %s: %w", target, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0777); err != nil {
				return fmt.Errorf("os.MkdirAll %s: %w", filepath.Dir(target), err)
			}

			outFile, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("os.Create %s: %w", target, err)
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("io.Copy to %s: %w", target, err)
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
		panic("error while unmarshaling response")
	}

	if version == "" {
		version = packageData.DistTags.Latest

	}

	latestVersion := packageData.Versions[version]
	response, err = http.Get(latestVersion.Dist.Tarball)

	if err != nil {
		panic("error while fetching tarball")
	}

	body = response.Body

	err = ExtractTarGz(body, "./node_modules/"+packageData.Id)
	if err != nil {
		panic("error while extracting tarball")
	}
	InstallDependencies(packageData.Versions[version].Dependencies)
}

func InstallDependencies(dependencies map[string]string) {
	var wg sync.WaitGroup
	for dependencyName, version := range dependencies {
		wg.Add(1)
		go func(dependencyName, version string) {
			defer wg.Done()
			_, err := strconv.Atoi(string(version[0]))
			if err != nil {
				InstallPackage(dependencyName, version[1:])
			} else {
				InstallPackage(dependencyName, version)
			}
		}(dependencyName, version)
	}
	wg.Wait()
}
