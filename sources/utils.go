package sources

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	lxd "github.com/lxc/lxd/shared"
)

// downloadChecksum downloads or opens URL, and matches fname against the
// checksums inside of the downloaded or opened file.
func downloadChecksum(targetDir string, URL string, fname string, hashFunc hash.Hash, hashLen int) ([]string, error) {
	var (
		client   http.Client
		tempFile *os.File
		err      error
	)

	// do not re-download checksum file if it's already present
	fi, err := os.Stat(filepath.Join(targetDir, URL))
	if err == nil && !fi.IsDir() {
		tempFile, err = os.Open(filepath.Join(targetDir, URL))
		if err != nil {
			return nil, err
		}
		defer os.Remove(tempFile.Name())
	} else {
		tempFile, err = ioutil.TempFile(targetDir, "hash.")
		if err != nil {
			return nil, err
		}
		defer os.Remove(tempFile.Name())

		_, err = lxd.DownloadFileHash(&client, "", nil, nil, "", URL, "", hashFunc, tempFile)
		// ignore hash mismatch
		if err != nil && !strings.HasPrefix(err.Error(), "Hash mismatch") {
			return nil, err
		}
	}

	tempFile.Seek(0, 0)

	checksum := getChecksum(filepath.Base(fname), hashLen, tempFile)
	if checksum != nil {
		return checksum, nil
	}

	return nil, errors.New("Could not find checksum")
}

func getChecksum(fname string, hashLen int, r io.Reader) []string {
	scanner := bufio.NewScanner(r)

	var matches []string
	var result []string

	for scanner.Scan() {
		if !strings.Contains(scanner.Text(), fname) {
			continue
		}

		for _, s := range strings.Split(scanner.Text(), " ") {
			m, _ := regexp.MatchString("[[:xdigit:]]+", s)
			if !m {
				continue
			}

			if hashLen == 0 || hashLen == len(strings.TrimSpace(s)) {
				matches = append(matches, scanner.Text())
			}
		}
	}

	// Check common checksum file (pattern: "<hash> <filename>") with the exact filename
	for _, m := range matches {
		fields := strings.Split(m, " ")

		if strings.TrimSpace(fields[len(fields)-1]) == fname {
			result = append(result, strings.TrimSpace(fields[0]))
		}
	}

	if len(result) > 0 {
		return result
	}

	// Check common checksum file (pattern: "<hash> <filename>") which contains the filename
	for _, m := range matches {
		fields := strings.Split(m, " ")

		if strings.Contains(strings.TrimSpace(fields[len(fields)-1]), fname) {
			result = append(result, strings.TrimSpace(fields[0]))
		}
	}

	if len(result) > 0 {
		return result
	}

	// Special case: CentOS
	for _, m := range matches {
		for _, s := range strings.Split(m, " ") {
			m, _ := regexp.MatchString("[[:xdigit:]]+", s)
			if !m {
				continue
			}

			if hashLen == 0 || hashLen == len(strings.TrimSpace(s)) {
				result = append(result, s)
			}
		}
	}

	if len(result) > 0 {
		return result
	}

	return nil
}

func recvGPGKeys(gpgDir string, keyserver string, keys []string) (bool, error) {
	args := []string{"--homedir", gpgDir}

	var fingerprints []string
	var publicKeys []string

	for _, k := range keys {
		if strings.HasPrefix(strings.TrimSpace(k), "-----BEGIN PGP PUBLIC KEY BLOCK-----") {
			publicKeys = append(publicKeys, strings.TrimSpace(k))
		} else {
			fingerprints = append(fingerprints, strings.TrimSpace(k))
		}
	}

	for _, f := range publicKeys {
		args := append(args, "--import")

		cmd := exec.Command("gpg", args...)
		cmd.Stdin = strings.NewReader(f)
		cmd.Env = append(os.Environ(), "LANG=C.UTF-8")

		var buffer bytes.Buffer
		cmd.Stderr = &buffer

		err := cmd.Run()
		if err != nil {
			return false, fmt.Errorf("Failed to run: %s: %s", strings.Join(cmd.Args, " "), strings.TrimSpace(buffer.String()))
		}
	}

	if keyserver != "" {
		args = append(args, "--keyserver", keyserver)
	}

	args = append(args, append([]string{"--recv-keys"}, fingerprints...)...)

	_, out, err := lxd.RunCommandSplit(append(os.Environ(), "LANG=C.UTF-8"), nil, "gpg", args...)
	if err != nil {
		return false, err
	}

	// Verify output
	var importedKeys []string
	var missingKeys []string
	lines := strings.Split(out, "\n")

	for _, l := range lines {
		if strings.HasPrefix(l, "gpg: key ") && (strings.HasSuffix(l, " imported") || strings.HasSuffix(l, " not changed")) {
			key := strings.Split(l, " ")
			importedKeys = append(importedKeys, strings.Split(key[2], ":")[0])
		}
	}

	// Figure out which key(s) couldn't be imported
	if len(importedKeys) < len(fingerprints) {
		for _, j := range fingerprints {
			found := false

			for _, k := range importedKeys {
				if strings.HasSuffix(j, k) {
					found = true
				}
			}

			if !found {
				missingKeys = append(missingKeys, j)
			}
		}

		return false, fmt.Errorf("Failed to import keys: %s", strings.Join(missingKeys, " "))
	}

	return true, nil
}
