package menu

import (
	"flag"
	"fmt"
	"os"
)

// ParseArgs -> Parses the args
func ParseArgs() {
	var (
		name, containerPath, filePath, dirPath string
		length                                 float64
	)

	cContainer := flag.Bool("create", false, "create a new container")
	addSingleFile := flag.Bool("addFile", false, "add a single file")
	addD := flag.Bool("addFiles", false, "add multiples files, by passing their dir path")
	mountPhysically := flag.Bool("mount", false, "mount the container")
	unmount := flag.Bool("unmount", false, "unmount the container")
	reset := flag.Bool("reset", false, "Resets the container, by cleaning all the content")

	flag.StringVar(&name, "name", "", "the container name")
	flag.Float64Var(&length, "length", float64(256), "the container length in kB, default is 256 kB")
	flag.StringVar(&containerPath, "containerPath", ".", "The place where the container will be created. Default is the current working directory")
	flag.StringVar(&filePath, "filePath", "", "Single file path")
	flag.StringVar(&dirPath, "dirPath", "", "Dir path")

	flag.Parse() // falta pritnar uma mensagem acerca

	if flag.Parsed() {
		if *cContainer {
			createContainer(containerPath, name, length)
		} else if *addSingleFile {
			addFile(containerPath, filePath)
		} else if *addD {
			addDir(containerPath, dirPath)
		} else if *mountPhysically {
			mountContainer(containerPath)
		} else if *unmount {
			unmountContainer(containerPath)
		} else if *reset {
			resetContainer(containerPath)
		} else {
			fmt.Printf("Usage %v [options]\n", os.Args[0])
			flag.PrintDefaults()
		}
	}
}
