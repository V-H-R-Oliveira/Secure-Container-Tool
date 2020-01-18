package menu

import (
	"core"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/howeyc/gopass"
)

var container *core.Container

func getUserKey() []byte {
	fmt.Print("Enter your key: ")
	key, err := gopass.GetPasswdMasked()
	if err != nil {
		log.Fatal(err)
	}
	return key
}

// createContainer -> Creates a new container
func createContainer(path, name string, length float64) {
	if path == "." {
		path, _ = os.Getwd()
	}

	if name != "" {
		path = filepath.Join(path, name)
		path += ".container"
		fmt.Println("The container was created, but i need a secure key to encrypt it (you must save the key, otherwise you will not be able to open the container again)")
		fmt.Println("The key size must be 16, 24 or 32")
		key := getUserKey()
		core.InitContainer(path, name, length)

		if !core.EncryptContainer(path, &key) {
			core.CleanTmpDir(path)
			return
		}

		fmt.Println("The container was created...")
	}else {
		log.Fatal("flags usage: -create -name=<container name> -containerPath=<where to save the .container, default is current working directory> -size=<container size in kB (only the number), default is 256 kB>")
	}
}

// addFile -> add a single file
func addFile(containerPath, filePath string) {
	if containerPath == "." || filePath == "" {
		log.Fatal("flags usage: -addFile -containerPath=<your .container path> -filePath=<file path>")
	}

	key := getUserKey()
	core.DecryptContainer(containerPath, &key)
	container.AddFile(filePath, containerPath)
	container = core.MountVirtually(container, containerPath)
	container.DisplayContainer()
	core.EncryptContainer(containerPath, &key)
}

// addDir -> Add multiples files, by passing the dir path
func addDir(containerPath, dirPath string) {
	if containerPath == "." || dirPath == "" {
		log.Fatal("flags usage: -addFiles -containerPath=<your .container path> -dirPath=<dir path>")
	}

	key := getUserKey()
	core.DecryptContainer(containerPath, &key)
	container.AddDir(dirPath, containerPath)
	container = core.MountVirtually(container, containerPath)
	container.DisplayContainer()
	core.EncryptContainer(containerPath, &key)
}

// mountContainer -> Mounts the container
func mountContainer(containerPath string) {
	if containerPath == "." {
		log.Fatal("flags usage: -mount -containerPath=<your .container path>")
	}

	key := getUserKey()
	core.DecryptContainer(containerPath, &key)
	core.MountContainer(container, containerPath)
	container = core.MountVirtually(container, containerPath)
	fmt.Println("The container was mounted at:", container.DirPath)
	core.EncryptContainer(containerPath, &key)
}

// unmountContainer -> Unmounts the container
func unmountContainer(containerPath string) {
	if containerPath == "." {
		log.Fatal("flags usage: -unmount -containerPath=<your .container path>")
	}

	key := getUserKey()
	core.DecryptContainer(containerPath, &key)
	container.DismountContainer(containerPath)
	fmt.Println("The container was dismounted at:")
	container = core.MountVirtually(container, containerPath)
	container.DisplayContainer()
	core.EncryptContainer(containerPath, &key)
}

func resetContainer(containerPath string) {
	if containerPath == "." {
		log.Fatal("flags usage: -reset -containerPath=<your .container path>")
	}

	key := getUserKey()
	core.DecryptContainer(containerPath, &key)
	container = core.MountVirtually(container, containerPath)
	container.ResetContainer(containerPath)
	fmt.Println("The container has been cleaned...")
	core.EncryptContainer(containerPath, &key)
}
