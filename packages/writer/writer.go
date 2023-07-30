package writer

import (
	"encoding/csv"
	"os"
	"sync"
)

var location string = "Writer"

var mu sync.Mutex

func OpenCsv(folderPath, fileName string, header []string) (*os.File, error) {

	filePath := folderPath + string(os.PathSeparator) + fileName

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// File do not exist, we need to create a new one
		file, err = os.Create(filePath)
		if err != nil {
			// Error while creating the new file
			return nil, err
		}

		// We set the title names
		if header != nil {
			err = writeHeader(file, header)
			if err != nil {
				return nil, err
			}
		}
	}

	return file, nil
}

func writeHeader(file *os.File, header []string) error {
	mu.Lock()
	defer mu.Unlock()

	writer := csv.NewWriter(file)

	err := writer.Write(header)
	if err != nil {
		return err
	}

	writer.Flush()
	return writer.Error()
}

func appendRow(file *os.File, row []string) error {
	mu.Lock()
	defer mu.Unlock()

	writer := csv.NewWriter(file)
	err := writer.Write(row)
	if err != nil {
		return err
	}

	// Flush dei dati nel file e salva le modifiche
	writer.Flush()

	return writer.Error()
}
