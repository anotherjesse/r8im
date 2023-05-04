package layers

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"strings"
)

func ExtractTarWithoutPrefixAndIgnoreWhiteout(r io.Reader, dest string) (bool, error) {
	var w io.Writer
	var file *os.File

	// Default to stdout
	if dest == "" {
		w = os.Stdout
	} else {
		var err error
		file, err = os.Create(dest)
		if err != nil {
			return false, err
		}
		defer file.Close()
		w = file
	}

	tr := tar.NewReader(r)
	tw := tar.NewWriter(w)
	defer tw.Close()

	prefix := "src/weights/"

	weightsFound := false

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return weightsFound, err
		}

		// Ignore whiteout files
		if strings.Contains(header.Name, ".wh..wh..") {
			continue
		}

		// Check if the path has the desired prefix
		if header.Typeflag == tar.TypeReg && strings.HasPrefix(header.Name, prefix) {
			weightsFound = true

			// Remove the prefix from the path
			newPath := strings.TrimPrefix(header.Name, prefix)
			header.Name = newPath

			fmt.Fprintln(os.Stderr, header.Name)

			err = tw.WriteHeader(header)
			if err != nil {
				return weightsFound, err
			}

			_, err = io.Copy(tw, tr)
			if err != nil {
				return weightsFound, err
			}
		}
	}
	return weightsFound, nil
}
