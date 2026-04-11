package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var placeholderMap = map[string]string{
	"devpod-provider-digitalocean-linux-amd64":       "##CHECKSUM_LINUX_AMD64##",
	"devpod-provider-digitalocean-linux-arm64":       "##CHECKSUM_LINUX_ARM64##",
	"devpod-provider-digitalocean-darwin-amd64":      "##CHECKSUM_DARWIN_AMD64##",
	"devpod-provider-digitalocean-darwin-arm64":      "##CHECKSUM_DARWIN_ARM64##",
	"devpod-provider-digitalocean-windows-amd64.exe": "##CHECKSUM_WINDOWS_AMD64##",
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: main.go <version>")
		os.Exit(1)
	}

	checksums, err := parseChecksums("./dist/checksums.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading checksums: %v\n", err)
		os.Exit(1)
	}

	content, err := os.ReadFile("./hack/provider/provider.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading provider.yaml: %v\n", err)
		os.Exit(1)
	}

	replaced := strings.ReplaceAll(string(content), "##VERSION##", os.Args[1])
	for filename, placeholder := range placeholderMap {
		checksum, ok := checksums[filename]
		if !ok {
			fmt.Fprintf(os.Stderr, "Warning: no checksum found for %s\n", filename)
			continue
		}
		replaced = strings.ReplaceAll(replaced, placeholder, checksum)
	}

	fmt.Print(replaced)
}

func parseChecksums(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	checksums := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if checksum, filename, ok := strings.Cut(scanner.Text(), "  "); ok {
			checksums[strings.TrimSpace(filename)] = checksum
		}
	}

	return checksums, scanner.Err()
}
