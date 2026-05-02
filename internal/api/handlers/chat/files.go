package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

type FilesConnector struct {
	logger *logrus.Logger
}

func NewFilesConnector(logger *logrus.Logger) *FilesConnector {
	if logger == nil {
		logger = logrus.New()
	}
	return &FilesConnector{logger: logger}
}

func (c *FilesConnector) Name() string { return "Files" }
func (c *FilesConnector) Icon() string { return "file" }
func (c *FilesConnector) IsConfigured() bool { return true }

func (c *FilesConnector) Authenticate(ctx context.Context, creds map[string]string) error {
	return nil
}

func (c *FilesConnector) FetchData(ctx context.Context, config map[string]interface{}) (interface{}, error) {
	filePath := config["file_path"]
	if filePath == nil {
		return nil, fmt.Errorf("file_path is required")
	}

	data, err := os.ReadFile(filePath.(string))
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(filePath.(string))
	switch ext {
	case ".json":
		var result interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		return json.Marshal(result)
	case ".csv":
		return string(data), nil
	case ".txt":
		return string(data), nil
	default:
		return fmt.Sprintf("File size: %d bytes", len(data)), nil
	}
}

func (c *FilesConnector) Search(ctx context.Context, query string, config map[string]interface{}) ([]SearchResult, error) {
	dirPath := config["dir_path"]
	if dirPath == nil {
		return nil, fmt.Errorf("dir_path is required")
	}

	entries, err := os.ReadDir(dirPath.(string))
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, entry := range entries {
		if !entry.IsDir() && strings.Contains(entry.Name(), query) {
			results = append(results, SearchResult{
				Title:   entry.Name(),
				Content: fmt.Sprintf("File: %s", entry.Name()),
			})
		}
	}
	return results, nil
}

func parseCSV(data string) ([]map[string]interface{}, error) {
	return nil, fmt.Errorf("CSV parsing not implemented - use a proper CSV library")
}

func handleFileUpload(r *http.Request) (map[string]interface{}, error) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"size": len(data),
		"data": string(data),
	}, nil
}
