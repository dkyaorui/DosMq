package data

import "os"

// In order to remove the database, we use the local file to replce the database.
// The design for this file system is written in README.md

// const data
const (
	oneGB    = 1073741824
	filePerm = 0666
)

// global data in package
var (
	nowWhence              int
	nowMessageFilePosition uint32
)

// write data to file
func write(filename string, data []byte) error {
	// check ths size of file first
	fileInfo, err := os.Stat(filename)
	if err == nil {
		fileSize := fileInfo.Size()
		if fileSize > 0.5*oneGB {
			// use new file
			nowMessageFilePosition++
			filename += string(nowMessageFilePosition)
		}
	}

	// append data to file
	file, _ := os.OpenFile(
		filename,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		filePerm,
	)
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	// create index

	return nil
}

// read data from file
func read(filename string, whence int) ([]byte, error) {
	return nil, nil
}
