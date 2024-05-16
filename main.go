package main

import (
	"fmt"
	"math/rand"
)

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

func main() {
	// A slice of strings containing the names of the students.
	names := []string{
		"Alex Ryan",
		"Fernando Valdez",
		"Javier Rice",
		"Quique Jefferson",
		"Nija Desirae",
		"Chava Rosenwort",
		"Salaam Muhammad",
		"Braeden Kincade",
		"Peng Zhang",
		"Axe Rivera",
		"Zander Ofosu",
		"Nicolas Sosa",
		"Wendy Daye",
		"TacoCat Dogod",
		"Locke Lamora",
	}

	numberOfPods := 4

	// Create students from names
	students := make([]Student, len(names))
	for i, name := range names {
		students[i] = &student{name: name}
	}

	// Assign pods to students
	createPods(students, numberOfPods)

	// Initialize a slice of slices of strings to represent the pods.
	pods := make([][]string, numberOfPods)

	// Fill pods with students
	for _, student := range students {
		pod := student.GetPod()
		pods[pod] = append(pods[pod], student.GetName())
	}

	// Will print number of students
	fmt.Printf("Number of students: %d\n", len(students))

	// Will iterate over the pods slice and print out the pod number and the students in that pod.
	for i, pod := range pods {
		fmt.Printf("Pod %d:\n", i+1)
		for _, name := range pod {
			fmt.Printf("\t%s\n", name)
		}
	}
}

// Take a slice of students and assign them to pods in a random order.
func createPods(students []Student, numberOfPods int) {
	rand.Shuffle(len(students), func(i, j int) {
		students[i], students[j] = students[j], students[i]
	})

	for i, student := range students {
		pod := i % numberOfPods
		student.SetPod(pod)
	}
}
