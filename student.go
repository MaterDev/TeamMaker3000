package main

// Student interface with required attributes
type Student interface {
	GetName() string
	IsRemote() bool
	GetPod() int
	SetPod(int)
}

// student struct implementing the Student interface
type student struct {
	name     string
	isRemote bool
	pod      int
}

func (s *student) GetName() string {
	return s.name
}

func (s *student) IsRemote() bool {
	return s.isRemote
}

func (s *student) GetPod() int {
	return s.pod
}

func (s *student) SetPod(pod int) {
	s.pod = pod
}
