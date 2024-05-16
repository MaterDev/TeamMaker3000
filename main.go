package main

import (
	"fmt"
	"math/rand"
)

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

	// Initialize a slice of slices of strings to represent the pods.
	pods := make([][]string, numberOfPods)

	// Will make a new slice to store the names of the students in a random order.
	// Copy the names into the randomOrderNames slice.
	// Use the rand.Shuffle function to randomize the order of the names.
	teams := createPods(names)

	// Will print number of students
	fmt.Printf("Number of students: %d\n", len(names))

	// Will print number of pods by iterating over the pods slice and printing out the pod number and the students in that pod.
	for i, name := range teams {
		pod := i % numberOfPods
		pods[pod] = append(pods[pod], name)
	}

	// Will iterate over the pods slice and print out the pod number and the students in that pod.
	for i, pod := range pods {
		fmt.Printf("Pod %d:\n", i+1)
		for _, name := range pod {
			fmt.Printf("\t%s\n", name)
		}
	}
}

// Take a slice of strings and return a new slice of strings with the same elements in a random order.
func createPods(names []string) []string {
	randomOrderNames := make([]string, len(names))

	copy(randomOrderNames, names)

	rand.Shuffle(len(randomOrderNames), func(i, j int) {
		randomOrderNames[i], randomOrderNames[j] = randomOrderNames[j], randomOrderNames[i]
	})
	return randomOrderNames
}
