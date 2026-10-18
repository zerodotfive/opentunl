package ui

import (
	"embed"
	"io"
	"path"
)

//go:embed dist/*.html
var EmbedHTML embed.FS

func GetEmbedHTMLBytes(filename string) ([]byte, error) {
	file, err := EmbedHTML.Open(path.Join("dist", filename))
	if err != nil {
		return nil, err
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}
