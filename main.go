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
	"sync"
	"teammaker3000/data" // Replace with the actual path to the data package

	_ "github.com/lib/pq" // Import the pq package
)

const (
	dbHost     = "localhost"
	dbPort     = 5432
	dbName     = "teammaker3000"
	maxRetries = 3
	batchSize  = 1000 // Number of database operations to perform in a single transaction
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
	db, err := connectToDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	nameData := data.Names
	minTeamSize, targetTeamSize := getTeamSizes(len(nameData))

	var pods [][]string

	for {
		students := make([]Student, len(nameData))
		for i, name := range nameData {
			students[i] = &student{name: name}
		}

		if err := createPodsMinCollabs(db, students, minTeamSize, targetTeamSize); err != nil {
			log.Printf("Error creating pods: %v", err)
			continue
		}

		numberOfPods := (len(students) + targetTeamSize - 1) / targetTeamSize
		pods = make([][]string, numberOfPods)
		for _, student := range students {
			pod := student.GetPod()
			pods[pod] = append(pods[pod], student.GetName())
		}

		pods, message := redistributeGroups(pods, minTeamSize, targetTeamSize)

		fmt.Printf("Number of students: %d\n", len(students))
		for i, pod := range pods {
			fmt.Printf("Pod %d (%d members):\n", i+1, len(pod))
			for _, name := range pod {
				fmt.Printf("\t%s\n", name)
			}
		}

		if message != "" {
			fmt.Println(message)
		}

		action := getConfirmation()

		if action == "A" {
			if err := saveGroupsToDB(db, pods); err != nil {
				log.Printf("Error saving groups to database: %v", err)
				fmt.Println("Failed to save teams. Please try again.")
			} else {
				fmt.Println("Teams saved to database.")
				break
			}
		} else if action == "X" {
			fmt.Println("Exiting without saving.")
			break
		}
	}
}

func connectToDB() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d dbname=%s sslmode=disable", dbHost, dbPort, dbName)
	var db *sql.DB
	var err error
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			if err = db.Ping(); err == nil {
				return db, nil
			}
		}
		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
	}
	return nil, fmt.Errorf("failed to connect to database after %d attempts", maxRetries)
}

func redistributeGroups(pods [][]string, minSize, targetSize int) ([][]string, string) {
	sort.Slice(pods, func(i, j int) bool {
		return len(pods[i]) < len(pods[j])
	})

	message := ""
	smallestPod := pods[0]
	if len(smallestPod) < minSize {
		studentsToRedistribute := len(smallestPod)
		message = fmt.Sprintf("The target team size of %d was not possible for all teams. %d student(s) from the smallest team have been redistributed to meet the minimum team size of %d.", targetSize, studentsToRedistribute, minSize)

		// Redistribute students from the smallest pod
		for _, student := range smallestPod {
			// Find the smallest pod that's not the current one
			smallestValidPod := -1
			for i, pod := range pods {
				if i != 0 && (smallestValidPod == -1 || len(pod) < len(pods[smallestValidPod])) {
					smallestValidPod = i
				}
			}
			pods[smallestValidPod] = append(pods[smallestValidPod], student)
		}

		// Remove the now-empty smallest pod
		pods = pods[1:]
	}

	return pods, message
}

func getTeamSizes(studentCount int) (int, int) {
	reader := bufio.NewReader(os.Stdin)
	var minSize, targetSize int
	
	for {
		fmt.Print("Enter minimum team size: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		var err error
		minSize, err = strconv.Atoi(input)
		if err != nil || minSize <= 1 || minSize > studentCount {
			fmt.Printf("Invalid input. Please enter a number between 2 and %d.\n", studentCount)
		} else {
			break
		}
	}

	for {
		fmt.Print("Enter target team size: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		var err error
		targetSize, err = strconv.Atoi(input)
		if err != nil || targetSize < minSize || targetSize > studentCount {
			fmt.Printf("Invalid input. Please enter a number between %d and %d.\n", minSize, studentCount)
		} else {
			break
		}
	}

	return minSize, targetSize
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
			fmt.Println("Invalid input. Please enter 'A' or 'X'.")
		}
	}
}

func createPodsMinCollabs(db *sql.DB, students []Student, minSize, targetSize int) error {
	collabMap, err := fetchCollaborations(db)
	if err != nil {
		return fmt.Errorf("error fetching collaborations: %v", err)
	}

	pairs := generateStudentPairs(students, collabMap)

	// Assign students to pods based on minimal collaborations
	assigned := make(map[string]bool)
	pods := make([][]Student, 0)
	pod := make([]Student, 0, targetSize)

	for _, pair := range pairs {
		if len(pod) < targetSize {
			for _, s := range []string{pair.student1, pair.student2} {
				if !assigned[s] {
					for i := range students {
						if students[i].GetName() == s {
							pod = append(pod, students[i])
							assigned[s] = true
							break
						}
					}
				}
				if len(pod) == targetSize {
					break
				}
			}
		}
		if len(pod) == targetSize || (len(pod) >= minSize && len(pairs) == 0) {
			pods = append(pods, pod)
			pod = make([]Student, 0, targetSize)
		}
	}

	// Assign any remaining students to pods
	for i := range students {
		if !assigned[students[i].GetName()] {
			if len(pod) == targetSize {
				pods = append(pods, pod)
				pod = make([]Student, 0, targetSize)
			}
			pod = append(pod, students[i])
			assigned[students[i].GetName()] = true
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

	return nil
}

func fetchCollaborations(db *sql.DB) (map[string]map[string]int, error) {
	collabMap := make(map[string]map[string]int)
	rows, err := db.Query(`SELECT s1.name, s2.name, c.collaborations_count 
						   FROM collaborations c 
						   JOIN students s1 ON c.student1_id = s1.id 
						   JOIN students s2 ON c.student2_id = s2.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var student1, student2 string
		var count int
		if err := rows.Scan(&student1, &student2, &count); err != nil {
			return nil, err
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
	return collabMap, rows.Err()
}

type studentPair struct {
	student1, student2 string
	weight             int
}

func generateStudentPairs(students []Student, collabMap map[string]map[string]int) []studentPair {
	pairs := make([]studentPair, 0, len(students)*(len(students)-1)/2)
	for i := 0; i < len(students); i++ {
		for j := i + 1; j < len(students); j++ {
			s1, s2 := students[i].GetName(), students[j].GetName()
			weight := collabMap[s1][s2]
			pairs = append(pairs, studentPair{student1: s1, student2: s2, weight: weight})
		}
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].weight < pairs[j].weight
	})

	return pairs
}

func saveGroupsToDB(db *sql.DB, pods [][]string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO collaborations (student1_id, student2_id, collaborations_count) 
		VALUES ($1, $2, 1)
		ON CONFLICT (student1_id, student2_id) 
		DO UPDATE SET collaborations_count = collaborations.collaborations_count + 1
	`)
	if err != nil {
		return fmt.Errorf("error preparing statement: %v", err)
	}
	defer stmt.Close()

	var wg sync.WaitGroup
	errorChan := make(chan error, len(pods))

	for _, pod := range pods {
		wg.Add(1)
		go func(pod []string) {
			defer wg.Done()
			for i := 0; i < len(pod); i++ {
				for j := i + 1; j < len(pod); j++ {
					student1ID, err := getStudentID(db, pod[i])
					if err != nil {
						errorChan <- fmt.Errorf("error getting student ID for %s: %v", pod[i], err)
						return
					}
					student2ID, err := getStudentID(db, pod[j])
					if err != nil {
						errorChan <- fmt.Errorf("error getting student ID for %s: %v", pod[j], err)
						return
					}
					if _, err := stmt.Exec(student1ID, student2ID); err != nil {
						errorChan <- fmt.Errorf("error inserting/updating collaboration: %v", err)
						return
					}
				}
			}
		}(pod)
	}

	wg.Wait()
	close(errorChan)

	for err := range errorChan {
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func getStudentID(db *sql.DB, name string) (int, error) {
	var id int
	err := db.QueryRow(`
		INSERT INTO students (name) 
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error inserting/retrieving student: %v", err)
	}
	return id, nil
}