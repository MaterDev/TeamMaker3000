package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"teammaker3000/data" // Replace with the actual path to the data package

	_ "github.com/lib/pq" // Import the pq package
)

const (
	dbHost = "localhost"
	dbPort = 5432
	dbName = "teammaker3000"
)

type Student interface {
	GetName() string
	GetPod() int
	SetPod(pod int)
}

type student struct {
	name string
	pod  int
}

func (s *student) GetName() string {
	return s.name
}

func (s *student) GetPod() int {
	return s.pod
}

func (s *student) SetPod(pod int) {
	s.pod = pod
}

func main() {
	connStr := fmt.Sprintf("host=%s port=%d dbname=%s sslmode=disable", dbHost, dbPort, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	nameData := data.Names
	teamSize := getTeamSize(len(nameData))

	var pods [][]string

	for {
		students := make([]Student, len(nameData))
		for i, name := range nameData {
			students[i] = &student{name: name}
		}

		createPodsMinCollabs(db, students, teamSize)

		numberOfPods := (len(students) + teamSize - 1) / teamSize // Calculate the number of pods based on the team size
		pods = make([][]string, numberOfPods)
		for _, student := range students {
			pod := student.GetPod()
			pods[pod] = append(pods[pod], student.GetName())
		}

		fmt.Printf("Number of students: %d\n", len(students))
		for i, pod := range pods {
			fmt.Printf("Pod %d:\n", i+1)
			for _, name := range pod {
				fmt.Printf("\t%s\n", name)
			}
		}

		action := getConfirmation()

		if action == "A" {
			saveGroupsToDB(db, pods)
			fmt.Println("Teams saved to database.")
			break
		} else if action == "X" {
			fmt.Println("Exiting without saving.")
			break
		}
	}
}

func getTeamSize(studentCount int) int {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter desired team size: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		teamSize, err := strconv.Atoi(input)
		if err != nil || teamSize <= 1 || teamSize > studentCount || studentCount%teamSize == 1 {
			fmt.Printf("Invalid input. Please enter a number between 2 and %d that won't result in a team of 1 person.\n", studentCount)
		} else {
			return teamSize
		}
	}
}

func getConfirmation() string {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter 'A' to accept or 'X' to exit without saving: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "A" || input == "X" {
			return input
		} else {
			fmt.Println("Invalid input. Please enter 'A', 'R', or 'X'.")
		}
	}
}

func createPodsMinCollabs(db *sql.DB, students []Student, teamSize int) {
	// Fetch collaborations from the database
	collabMap := make(map[string]map[string]int)
	rows, err := db.Query(`SELECT s1.name, s2.name, c.collaborations_count 
						   FROM collaborations c 
						   JOIN students s1 ON c.student1_id = s1.id 
						   JOIN students s2 ON c.student2_id = s2.id`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var student1, student2 string
		var count int
		if err := rows.Scan(&student1, &student2, &count); err != nil {
			log.Fatal(err)
		}
		if collabMap[student1] == nil {
			collabMap[student1] = make(map[string]int)
		}
		collabMap[student1][student2] = count
		if collabMap[student2] == nil {
			collabMap[student2] = make(map[string]int)
		}
		collabMap[student2][student1] = count
	}

	// Calculate collaboration weights for each student pair
	type studentPair struct {
		student1 string
		student2 string
		weight   int
	}

	var pairs []studentPair
	for i := 0; i < len(students); i++ {
		for j := i + 1; j < len(students); j++ {
			s1, s2 := students[i].GetName(), students[j].GetName()
			weight := collabMap[s1][s2]
			pairs = append(pairs, studentPair{student1: s1, student2: s2, weight: weight})
		}
	}

	// Sort pairs by collaboration weight in ascending order
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].weight < pairs[j].weight
	})

	// Assign students to pods based on minimal collaborations
	assigned := make(map[string]bool)
	pods := make([][]Student, 0)
	pod := make([]Student, 0, teamSize)
	for _, pair := range pairs {
		if len(pod) < teamSize {
			if !assigned[pair.student1] {
				for i := 0; i < len(students); i++ {
					if students[i].GetName() == pair.student1 && !assigned[pair.student1] {
						pod = append(pod, students[i])
						assigned[pair.student1] = true
						break
					}
				}
			}
			if len(pod) < teamSize && !assigned[pair.student2] {
				for i := 0; i < len(students); i++ {
					if students[i].GetName() == pair.student2 && !assigned[pair.student2] {
						pod = append(pod, students[i])
						assigned[pair.student2] = true
						break
					}
				}
			}
		}
		if len(pod) == teamSize {
			pods = append(pods, pod)
			pod = make([]Student, 0, teamSize)
		}
	}

	// Assign any remaining students to pods
	for _, student := range students {
		if !assigned[student.GetName()] {
			if len(pod) == teamSize {
				pods = append(pods, pod)
				pod = make([]Student, 0, teamSize)
			}
			pod = append(pod, student)
			assigned[student.GetName()] = true
		}
	}
	if len(pod) > 0 {
		pods = append(pods, pod)
	}

	// Assign pods to students
	for podIndex, pod := range pods {
		for _, student := range pod {
			student.SetPod(podIndex)
		}
	}
}

func saveGroupsToDB(db *sql.DB, pods [][]string) {
	for _, pod := range pods {
		for i := 0; i < len(pod); i++ {
			for j := i + 1; j < len(pod); j++ {
				student1 := pod[i]
				student2 := pod[j]
				student1ID := getStudentID(db, student1)
				student2ID := getStudentID(db, student2)
				if student1ID != 0 && student2ID != 0 {
					_, err := db.Exec(`
						INSERT INTO collaborations (student1_id, student2_id, collaborations_count) 
						VALUES ($1, $2, 1)
						ON CONFLICT (student1_id, student2_id) 
						DO UPDATE SET collaborations_count = collaborations.collaborations_count + 1
					`, student1ID, student2ID)
					if err != nil {
						log.Println("Error inserting/updating collaboration:", err)
					}
				}
			}
		}
	}
}

func getStudentID(db *sql.DB, name string) int {
	var id int
	err := db.QueryRow(`
		INSERT INTO students (name) 
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, name).Scan(&id)
	if err != nil {
		log.Println("Error inserting/retrieving student:", err)
		return 0
	}
	return id
}
