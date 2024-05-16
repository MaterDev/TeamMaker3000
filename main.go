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

	// Will make a new slice with the names of the students in mixed order.
	randomOrderNames := make([]string, len(names))

	// Copy the names into the randomOrderNames slice.
	copy(randomOrderNames, names)

	// Use the rand.Shuffle function to randomize the order of the names.
	rand.Shuffle(len(randomOrderNames), func(i, j int) {
		randomOrderNames[i], randomOrderNames[j] = randomOrderNames[j], randomOrderNames[i]
	})

	// will print number of students
	fmt.Printf("Number of students: %d\n", len(names))

	// will iterate over the names slice and assign each name to a pod randomly. until all names are assigned.
	for i, name := range randomOrderNames {
		pod := i % numberOfPods
		pods[pod] = append(pods[pod], name)
	}

	// will iterate over the pods slice and print out the pod number and the students in that pod.
	for i, pod := range pods {
		fmt.Printf("Pod %d:\n", i+1)
		for _, name := range pod {
			fmt.Printf("\t%s\n", name)
		}
	}
}
