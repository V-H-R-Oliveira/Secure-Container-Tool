package core

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// encodeToFile -> Encode a golang struct to a file
func encodeToFile(container *Container, path string) {
	buffer := &bytes.Buffer{}
	e := gob.NewEncoder(buffer)
	err := e.Encode(container)

	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(path, buffer.Bytes(), 0600)

	if err != nil {
		log.Fatal(err)
	}
}

// decodeFromFile -> Decode binary data from file to a Container struct
func decodeFromFile(container *Container, path string) *Container {
	content, err := ioutil.ReadFile(path)
	decoder := gob.NewDecoder(bytes.NewReader(content))

	if err != nil {
		log.Fatal(err)
	}

	err = decoder.Decode(&container)

	if err != nil {
		log.Fatal(err)
	}

	return container
}

// writeFileInContainer -> write a file from the path to the container
func writeFileInContainer(container *Container, filePath, containerPath string) {
	content, err := ioutil.ReadFile(filePath)
	filename := filepath.Base(filePath)

	if err != nil {
		log.Fatal(err)
	}

	size := getFileSize(filePath)

	if size > container.Size || (container.Usage+size) > container.Size {
		fmt.Println("Size is greater than container size")
		return
	}

	if _, ok := container.Structure[filename]; ok {
		fmt.Printf("The %s is already in the container.\n", filename)
		return
	}

	container.Structure[filename] = content
	container.Usage += size
}

// unlinkFile -> Remove a file from the container
func unlinkFile(c *Container, filename string) bool {
	if _, ok := c.Structure[filename]; !ok {
		return false
	}

	delete(c.Structure, filename)
	return true
}

// getFileSize -> returns the filesize in kB
func getFileSize(path string) float64 {
	fileinfo, err := os.Stat(path)
	size := float64(float64(fileinfo.Size()) / 1024)

	if err != nil {
		log.Fatal(err)
	}

	return size
}

// willDir -> Fill the tmp dir with the files
func fillDir(dirPath string, file map[string][]byte) {
	for path, content := range file {
		fileDir := filepath.Join(dirPath, path)

		if err := ioutil.WriteFile(fileDir, content, 0666); err != nil {
			log.Fatal(err)
		}
	}
}

// writeFileToDir -> write a file from the container to a dir
func writeFileToDir(dirPath string, file chan map[string][]byte, quit chan bool) {
	for {
		select {
		case val := <-file:
			fillDir(dirPath, val)
		case <-quit:
			return
		}
	}
}

// writeDirInContainer -> write a full dir inside the container
func writeDirInContainer(container *Container, containerPath, dirName string, ch chan os.FileInfo, quit chan bool) {
	for {
		select {
		case file := <-ch:
			filePath := filepath.Join(dirName, file.Name())
			writeFileInContainer(container, filePath, containerPath)
		case <-quit:
			return
		}
	}
}

// updateContainer -> Updates a container file
func updateContainer(container *Container, filePath string, oldContent *[]byte) {
	newContent, err := ioutil.ReadFile(filePath)
	oldLength := float64((float64(len(*oldContent)) / 1024.0))
	newSize := getFileSize(filePath)
	filename := filepath.Base(filePath)

	if err != nil {
		log.Fatal(err)
	}

	if oldLength == newSize {
		fmt.Println("Equal length, only updates the content...")
		container.Structure[filename] = newContent
		return
	}

	container.Usage -= (oldLength - newSize)
	container.Structure[filename] = newContent
}

// contains -> auxiliary function, that performs a linear search in the slice
func contains(files *[]os.FileInfo, fileName string) bool {
	for _, file := range *files {
		if fileName == file.Name() {
			return true
		}
	}
	return false
}

// removeUnseen -> Remove the dead files in the container
func removeUnseen(container *Container, files *[]os.FileInfo) {
	for file, content := range container.Structure {
		if !contains(files, file) {
			fmt.Printf("Removing %s from the container\n", file)
			updatedSize := float64((float64(len(content)) / 1024.0))

			if unlinkFile(container, file) {
				container.Usage -= updatedSize
				fmt.Println("The file was removed")
			} else {
				log.Fatal("There was a mistake")
			}
		}
	}
}

// saveContainer -> Save and Update the container.
func saveContainer(container *Container, containerPath string, ch chan os.FileInfo, files *[]os.FileInfo) {
	for file := range ch {
		fileName := file.Name()
		fullPath := filepath.Join(container.DirPath, fileName)

		if _, ok := container.Structure[fileName]; !ok {
			writeFileInContainer(container, fullPath, containerPath)
		} else if content, exist := container.Structure[fileName]; exist {
			updateContainer(container, fullPath, &content)
		}

		removeUnseen(container, files)
	}
}

// CleanTmpDir -> Remove the tmp dir
func CleanTmpDir(dirPath string) {
	if err := os.RemoveAll(dirPath); err != nil {
		log.Fatal(err)
	}
}

// hmacAuthEncrypt -> Sign the container after encrypt the data (based on https://github.com/andrewromanenco/gcrypt/blob/master/gcrypt.go)
func hmacAuthEncrypt(ciphertext, key *[]byte) {
	h := hmac.New(sha512.New, *key)
	h.Write(*ciphertext)
	hash := h.Sum(nil)
	*ciphertext = append(*ciphertext, hash...)
}

// hmacAuthDecrypt -> Authenticate the checksum (based on https://github.com/andrewromanenco/gcrypt/blob/master/gcrypt.go)
func hmacAuthDecrypt(ciphertext, key *[]byte) {
	offset := len(*ciphertext) - 64
	sum := (*ciphertext)[offset:]
	*ciphertext = (*ciphertext)[:offset]
	h := hmac.New(sha512.New, *key)
	h.Write(*ciphertext)
	hash := h.Sum(nil)

	if hash == nil || len(hash) != len(sum) {
		log.Fatal("The container is currupted\n")
	}

	for index := range hash {
		if hash[index] != sum[index] {
			log.Fatal("Bytes mismatch...")
		}
	}
}
