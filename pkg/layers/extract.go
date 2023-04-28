package layers

import (
	"archive/tar"
	"io"
	"os"
	"strings"
)

func ExtractTarWithoutPrefix(r io.Reader, dest string) error {
	var w io.Writer
	var file *os.File

	// Default to stdout
	if dest == "" {
		w = os.Stdout
	} else {
		var err error
		file, err = os.Create(dest)
		if err != nil {
			return err
		}
		defer file.Close()
		w = file
	}

	tr := tar.NewReader(r)
	tw := tar.NewWriter(w)
	defer tw.Close()

	prefix := "src/weights/"

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Check if the path has the desired prefix
		if strings.HasPrefix(header.Name, prefix) {
			// Remove the prefix from the path
			newPath := strings.TrimPrefix(header.Name, prefix)
			header.Name = newPath
		}

		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(tw, tr)
		if err != nil {
			return err
		}
	}
	return nil
}
