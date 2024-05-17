package main

import (
	"fmt"
	"math/rand"
	"teammaker3000/data" // Replace with the actual path to the data package
)

func main() {
	nameData := data.Names
	numberOfPods := 4

	// Create students from names
	students := make([]Student, len(nameData))
	for i, name := range nameData {
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

	// Print number of students
	fmt.Printf("Number of students: %d\n", len(students))

	// Iterate over the pods slice and print out the pod number and the students in that pod.
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
