package main

import "fmt"

func main() {

	/**
	Ani Subbarao
	Brock Biever
	Crystal EvenRam (Remote)
	Joshua Bass  (Remote)
	Justin Kahn (Remote)
	Kalia Vang
	Luke Edwards
	Meleah Roe
	Michael Lafond
	Sam Shacter
	Sumeiya Abdi (Remote)
	Tierra Byrd
	Tristan Bridges
	 */
	// An array of names, which will be student names.
	names := []string{
		"Ani Subbarao",
		"Brock Biever",
		"Crystal EvenRam (Remote)",
		"Joshua Bass  (Remote)",
		"Justin Kahn (Remote)",
		"Kalia Vang",
		"Luke Edwards",
		"Michael Lafond",
		"Sam Shacter",
		"Sumeiya Abdi (Remote)",
		"Tierra Byrd",
		"Tristan Bridges",
		}

	numberOfPods := 4

	// will create a 2-dimensional slice of strings with the number of pods. Students will be divided into pods.
	pods := make([][]string, numberOfPods)

	// will print number of students
	fmt.Printf("Number of students: %d\n", len(names))

	// will iterate over the names slice and assign each name to a pod.
	for i, name := range names {
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