package core

// Container -> An abstract container representation
type Container struct {
	Name, DirPath string
	Size, Usage   float64
	Structure     map[string][]byte
}
