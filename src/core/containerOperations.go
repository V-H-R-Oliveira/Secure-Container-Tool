package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// InitContainer -> initate container instance
func InitContainer(path, name string, size float64) *Container {
	container := &Container{name, "", size, float64(0), make(map[string][]byte)}
	encodeToFile(container, path)
	return container
}

// MountVirtually -> Mount the container virtually
func MountVirtually(container *Container, containerPath string) *Container {
	return decodeFromFile(container, containerPath)
}

// AddFile -> Map a file to the container
func (c *Container) AddFile(filePath, containerPath string) {
	c = MountVirtually(c, containerPath)
	writeFileInContainer(c, filePath, containerPath)
	encodeToFile(c, containerPath)
}

// AddDir -> Add an entire dir to the container
func (c *Container) AddDir(dirPath, containerPath string) {
	c = decodeFromFile(c, containerPath)
	files, err := ioutil.ReadDir(dirPath)
	ch, quit := make(chan os.FileInfo), make(chan bool)

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for _, file := range files {
			if file.IsDir() { // reads inner dirs recursively
				fullPath := filepath.Join(dirPath, file.Name())
				c.AddDir(fullPath, containerPath)
			} else {
				ch <- file
			}
		}
		quit <- true
	}()

	writeDirInContainer(c, containerPath, dirPath, ch, quit)
	encodeToFile(c, containerPath)
}

// RemoveFileFromContainer -> Unlink a file from the container
func (c *Container) RemoveFileFromContainer(path, containerPath string) {
	size := getFileSize(path)
	c = decodeFromFile(c, containerPath)

	if unlinkFile(c, path) {
		c.Usage -= size * 100
		fmt.Printf("%s was unlinked.\n", path)
		encodeToFile(c, containerPath)
		return
	}
}

// MountContainer -> Mount the container physically
func MountContainer(container *Container, containerPath string) {
	dir, err := ioutil.TempDir(".", "container")
	container = decodeFromFile(container, containerPath)
	quit, file := make(chan bool), make(chan map[string][]byte)

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for path, content := range container.Structure {
			tmp := make(map[string][]byte)
			tmp[path] = content
			file <- tmp
		}
		quit <- true
	}()

	writeFileToDir(dir, file, quit)
	container.DirPath = dir
	encodeToFile(container, containerPath)
}

// ResetContainer -> Resets a container, by cleaning all the stuff inside the container
func (c *Container) ResetContainer(containerPath string) {
	for file := range c.Structure {
		unlinkFile(c, file)
	}

	c.Usage = float64(0)
	encodeToFile(c, containerPath)
}

// DismountContainer -> Dismount the tmp dir and save/update the container
func (c *Container) DismountContainer(containerPath string) {
	c = decodeFromFile(c, containerPath)
	files, err := ioutil.ReadDir(c.DirPath)
	ch := make(chan os.FileInfo)

	defer CleanTmpDir(c.DirPath)

	if err != nil {
		log.Fatal(err)
	}

	if len(files) == 0 {
		c.ResetContainer(containerPath)
		return
	}

	go func() {
		for _, file := range files {
			if file.IsDir() {
				fmt.Println("Only files are allowed")
			} else {
				ch <- file
			}
		}
		close(ch)
	}()

	saveContainer(c, containerPath, ch, &files)
	encodeToFile(c, containerPath)
}

//EncryptContainer -> Encrypts the container (modified from https://golang.org/pkg/crypto/cipher/#example_NewCFBEncrypter)
func EncryptContainer(containerPath string, key *[]byte) bool {
	if len(*key) != 16 && len(*key) != 24 && len(*key) != 32 {
		fmt.Println("Key length is too short, it must be 16, 24 or 32 length. (recommended 32 bytes)")
		return false
	}

	content, err := ioutil.ReadFile(containerPath)

	if err != nil {
		log.Fatal(err)
	}

	fileinfo, err2 := os.Stat(containerPath)

	if err2 != nil {
		log.Fatal(err2)
	}

	block, err3 := aes.NewCipher(*key)

	if err3 != nil {
		log.Fatal(err3)
	}

	ciphertext := make([]byte, aes.BlockSize+fileinfo.Size())
	iv := ciphertext[:aes.BlockSize]

	if _, err4 := io.ReadFull(rand.Reader, iv); err4 != nil {
		log.Fatal(err4)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], content)

	hmacAuthEncrypt(&ciphertext, key)

	if err5 := ioutil.WriteFile(containerPath, ciphertext, 0666); err5 == nil {
		fmt.Println("The container was encrypted")
	} else {
		log.Fatal(err5)
	}

	return true
}

// DecryptContainer -> Decrypts a container (modified from https://golang.org/pkg/crypto/cipher/#example_NewCFBDecrypter)
func DecryptContainer(containerPath string, key *[]byte) {
	ciphertext, err := ioutil.ReadFile(containerPath)

	if err != nil {
		log.Fatal(err)
	}

	hmacAuthDecrypt(&ciphertext, key)

	block, err2 := aes.NewCipher(*key)

	if err2 != nil {
		log.Fatal(err2)
	}

	if len(ciphertext) < aes.BlockSize {
		log.Fatal("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	// writes the decryted content
	if err3 := ioutil.WriteFile(containerPath, ciphertext, 0666); err3 == nil {
		fmt.Println("The container was decrypted")
	} else {
		log.Fatal(err3)
	}
}

// DisplayContainer -> Display the container information in a prettier way
func (c *Container) DisplayContainer() {
	file := 1
	fmt.Println("============================")
	fmt.Println("Container:", c.Name)
	fmt.Printf("It is already using %.2f kB from %.2f kB\n", c.Usage, c.Size)
	fmt.Println("Content composition:")
	for filename := range c.Structure {
		fmt.Println("File", file, "->", filename)
		file++
	}
	fmt.Println("============================")
}
