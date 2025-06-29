package storage

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"teammaker3000/models"

	"github.com/google/uuid" // For generating team IDs
	_ "github.com/lib/pq"    // PostgreSQL driver
)

// InitializeSchema creates the necessary tables if they don't exist.
func (ps *PostgresStorage) InitializeSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS students (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS collaborations (
			id SERIAL PRIMARY KEY,
			student1_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
			student2_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
			collaborations_count INTEGER DEFAULT 0,
			UNIQUE(student1_id, student2_id),
			CONSTRAINT check_student_order CHECK (student1_id < student2_id) -- Ensures unique pairs regardless of order
		);`,
		`CREATE TABLE IF NOT EXISTS teams (
			team_id VARCHAR(36) PRIMARY KEY, -- UUID
			team_name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS team_members (
			team_id VARCHAR(36) NOT NULL REFERENCES teams(team_id) ON DELETE CASCADE,
			student_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
			PRIMARY KEY (team_id, student_id)
		);`,
	}

	for _, query := range queries {
		_, err := ps.DB.Exec(query)
		if err != nil {
			return fmt.Errorf("error executing schema query: %v\nQuery: %s", err, query)
		}
	}
	log.Println("Database schema initialized successfully.")
	return nil
}

// AddPerson adds a new person to the students table.
func (ps *PostgresStorage) AddPerson(name string) (*models.Person, error) {
	var id int
	err := ps.DB.QueryRow(`
		INSERT INTO students (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name -- Effectively does nothing if name exists, but returns ID
		RETURNING id
	`, name).Scan(&id)
	if err != nil {
		// If the conflict is on the unique constraint, try to get the existing user.
		// This handles cases where ON CONFLICT DO UPDATE might not return the ID as expected by some drivers/versions or if we want to be explicit.
		var existingID int
		errQuery := ps.DB.QueryRow("SELECT id FROM students WHERE name = $1", name).Scan(&existingID)
		if errQuery == nil {
			return &models.Person{ID: existingID, Name: name}, nil
		}
		return nil, fmt.Errorf("error inserting or retrieving student %s: %v", name, err)
	}
	return &models.Person{ID: id, Name: name}, nil
}

// GetOrCreatePersonByName retrieves a person by name or creates them if they don't exist.
func (ps *PostgresStorage) GetOrCreatePersonByName(name string) (*models.Person, error) {
	var person models.Person
	err := ps.DB.QueryRow("SELECT id, name FROM students WHERE name = $1", name).Scan(&person.ID, &person.Name)
	if err == sql.ErrNoRows {
		// Person doesn't exist, create them
		return ps.AddPerson(name)
	} else if err != nil {
		return nil, fmt.Errorf("error retrieving person by name %s: %v", name, err)
	}
	return &person, nil
}

// GetPersonByID retrieves a person by their ID.
func (ps *PostgresStorage) GetPersonByID(id int) (*models.Person, error) {
	var person models.Person
	err := ps.DB.QueryRow("SELECT id, name FROM students WHERE id = $1", id).Scan(&person.ID, &person.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("person with ID %d not found", id)
		}
		return nil, fmt.Errorf("error retrieving person by ID %d: %v", id, err)
	}
	return &person, nil
}

// GetPeople retrieves all people from the students table.
func (ps *PostgresStorage) GetPeople() ([]*models.Person, error) {
	rows, err := ps.DB.Query("SELECT id, name FROM students ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("error querying people: %v", err)
	}
	defer rows.Close()

	var people []*models.Person
	for rows.Next() {
		var person models.Person
		if err := rows.Scan(&person.ID, &person.Name); err != nil {
			return nil, fmt.Errorf("error scanning person row: %v", err)
		}
		people = append(people, &person)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error after iterating person rows: %v", err)
	}
	return people, nil
}

// SaveTeam saves a team and its members. If the team ID is empty, it generates one.
func (ps *PostgresStorage) SaveTeam(team *models.Team) error {
	tx, err := ps.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback() // Rollback if not committed

	if team.ID == "" {
		team.ID = uuid.NewString()
	}

	_, err = tx.Exec(`
		INSERT INTO teams (team_id, team_name)
		VALUES ($1, $2)
		ON CONFLICT (team_id) DO UPDATE SET team_name = EXCLUDED.team_name
	`, team.ID, team.Name)
	if err != nil {
		return fmt.Errorf("error saving team %s (ID: %s): %v", team.Name, team.ID, err)
	}

	// Clear existing members for this team before adding current ones to handle updates correctly
	_, err = tx.Exec("DELETE FROM team_members WHERE team_id = $1", team.ID)
	if err != nil {
		return fmt.Errorf("error clearing existing members for team ID %s: %v", team.ID, err)
	}

	for _, member := range team.Members {
		if member.ID == 0 { // Ensure member has a valid ID
			// Try to get or create the person if their ID is missing
			p, err := ps.GetOrCreatePersonByName(member.Name) // This uses a non-transactional ps.DB, consider passing tx or making GetOrCreatePersonByName accept a Querier
			if err != nil {
				return fmt.Errorf("error ensuring person %s exists for team %s: %v", member.Name, team.Name, err)
			}
			member.ID = p.ID
		}
		_, err := tx.Exec("INSERT INTO team_members (team_id, student_id) VALUES ($1, $2)", team.ID, member.ID)
		if err != nil {
			return fmt.Errorf("error adding member %s (ID: %d) to team %s (ID: %s): %v", member.Name, member.ID, team.Name, team.ID, err)
		}
	}

	return tx.Commit()
}

// GetTeamByID retrieves a team and its members by the team's UUID.
func (ps *PostgresStorage) GetTeamByID(teamID string) (*models.Team, error) {
	var team models.Team
	err := ps.DB.QueryRow("SELECT team_id, team_name FROM teams WHERE team_id = $1", teamID).Scan(&team.ID, &team.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team with ID %s not found", teamID)
		}
		return nil, fmt.Errorf("error retrieving team by ID %s: %v", teamID, err)
	}

	rows, err := ps.DB.Query(`
		SELECT s.id, s.name
		FROM students s
		JOIN team_members tm ON s.id = tm.student_id
		WHERE tm.team_id = $1
	`, teamID)
	if err != nil {
		return nil, fmt.Errorf("error querying members for team ID %s: %v", teamID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var member models.Person
		if err := rows.Scan(&member.ID, &member.Name); err != nil {
			return nil, fmt.Errorf("error scanning member for team ID %s: %v", teamID, err)
		}
		team.Members = append(team.Members, &member)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error after iterating members for team ID %s: %v", teamID, err)
	}

	return &team, nil
}

// GetAllTeams retrieves all saved teams with their members.
func (ps *PostgresStorage) GetAllTeams() ([]*models.Team, error) {
	rows, err := ps.DB.Query("SELECT team_id, team_name FROM teams ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("error querying all teams: %v", err)
	}
	defer rows.Close()

	var teams []*models.Team
	for rows.Next() {
		var partialTeam models.Team
		if err := rows.Scan(&partialTeam.ID, &partialTeam.Name); err != nil {
			return nil, fmt.Errorf("error scanning team row: %v", err)
		}
		// For each team, fetch its members. This can lead to N+1 queries.
		// For larger datasets, a JOIN and careful row processing would be more efficient.
		fullTeam, err := ps.GetTeamByID(partialTeam.ID)
		if err != nil {
			// Log or handle error for individual team fetch, maybe skip this team
			log.Printf("Warning: could not fetch full details for team ID %s: %v", partialTeam.ID, err)
			continue
		}
		teams = append(teams, fullTeam)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error after iterating team rows: %v", err)
	}
	return teams, nil
}


// GetCollaborationCount retrieves the number of times two people have collaborated.
func (ps *PostgresStorage) GetCollaborationCount(person1ID, person2ID int) (int, error) {
	// Ensure order for querying unique pair
	if person1ID > person2ID {
		person1ID, person2ID = person2ID, person1ID
	}

	var count int
	err := ps.DB.QueryRow(`
		SELECT collaborations_count
		FROM collaborations
		WHERE student1_id = $1 AND student2_id = $2
	`, person1ID, person2ID).Scan(&count)

	if err == sql.ErrNoRows {
		return 0, nil // No collaboration record means count is 0
	}
	if err != nil {
		return 0, fmt.Errorf("error getting collaboration count for students %d and %d: %v", person1ID, person2ID, err)
	}
	return count, nil
}

// IncrementCollaborationCount increments the collaboration count between two people.
func (ps *PostgresStorage) IncrementCollaborationCount(person1ID, person2ID int) error {
	// Ensure order for inserting/updating unique pair
	if person1ID > person2ID {
		person1ID, person2ID = person2ID, person1ID
	}
    if person1ID == person2ID { // Cannot collaborate with oneself
        return fmt.Errorf("cannot increment collaboration for a person with themselves (ID: %d)", person1ID)
    }


	_, err := ps.DB.Exec(`
		INSERT INTO collaborations (student1_id, student2_id, collaborations_count)
		VALUES ($1, $2, 1)
		ON CONFLICT (student1_id, student2_id)
		DO UPDATE SET collaborations_count = collaborations.collaborations_count + 1
	`, person1ID, person2ID)

	if err != nil {
		// Check if the error is due to foreign key constraint violation
		if strings.Contains(err.Error(), "violates foreign key constraint") {
			// Attempt to get more specific info about which student ID might be missing
			var exists1, exists2 bool
			ps.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM students WHERE id = $1)", person1ID).Scan(&exists1)
			ps.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM students WHERE id = $1)", person2ID).Scan(&exists2)
			if !exists1 {
				return fmt.Errorf("error incrementing collaboration: student with ID %d does not exist. Error: %v", person1ID, err)
			}
			if !exists2 {
				return fmt.Errorf("error incrementing collaboration: student with ID %d does not exist. Error: %v", person2ID, err)
			}
		}
		return fmt.Errorf("error incrementing collaboration count for students %d and %d: %v", person1ID, person2ID, err)
	}
	return nil
}

// SaveGeneratedTeamsAndCollaborations saves newly generated teams and updates collaboration counts.
func (ps *PostgresStorage) SaveGeneratedTeamsAndCollaborations(teams []*models.Team) error {
	tx, err := ps.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for saving generated teams: %v", err)
	}
	defer tx.Rollback()

	for _, team := range teams {
		if team.ID == "" {
			team.ID = uuid.NewString() // Generate UUID if not already set
		}
		if team.Name == "" { // Provide a default name if empty
			team.Name = "Generated Team " + team.ID[:8]
		}


		// Save the team itself
		_, err = tx.Exec(`
			INSERT INTO teams (team_id, team_name)
			VALUES ($1, $2)
			ON CONFLICT (team_id) DO UPDATE SET team_name = EXCLUDED.team_name
		`, team.ID, team.Name)
		if err != nil {
			return fmt.Errorf("error saving team %s (ID: %s) in transaction: %v", team.Name, team.ID, err)
		}

		// Save team members and update collaborations
		for i := 0; i < len(team.Members); i++ {
			member1 := team.Members[i]
			// Ensure member1 has an ID (should be guaranteed by core logic if fetched/created properly)
			if member1.ID == 0 {
				p, err := ps.getOrCreatePersonByNameTx(tx, member1.Name) // Use transactional version
				if err != nil {
					return fmt.Errorf("error ensuring person %s exists for team %s: %v", member1.Name, team.Name, err)
				}
				member1.ID = p.ID
				member1.Name = p.Name // Update name just in case there was a slight variation resolved by DB
			}


			// Save team membership
			_, err = tx.Exec("INSERT INTO team_members (team_id, student_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", team.ID, member1.ID)
			if err != nil {
				return fmt.Errorf("error adding member %s (ID: %d) to team %s (ID: %s) in transaction: %v", member1.Name, member1.ID, team.Name, team.ID, err)
			}

			// Increment collaborations with other members of the same team
			for j := i + 1; j < len(team.Members); j++ {
				member2 := team.Members[j]
				if member2.ID == 0 {
					p, err := ps.getOrCreatePersonByNameTx(tx, member2.Name) // Use transactional version
					if err != nil {
						return fmt.Errorf("error ensuring person %s exists for team %s: %v", member2.Name, team.Name, err)
					}
					member2.ID = p.ID
					member2.Name = p.Name
				}


				p1ID, p2ID := member1.ID, member2.ID
				if p1ID > p2ID {
					p1ID, p2ID = p2ID, p1ID
				}
                if p1ID == p2ID { continue } // Skip if it's the same person

				_, err = tx.Exec(`
					INSERT INTO collaborations (student1_id, student2_id, collaborations_count)
					VALUES ($1, $2, 1)
					ON CONFLICT (student1_id, student2_id)
					DO UPDATE SET collaborations_count = collaborations.collaborations_count + 1
				`, p1ID, p2ID)
				if err != nil {
                    // More detailed error logging for collaboration increment failure
                    var exists1, exists2 bool
                    tx.QueryRow("SELECT EXISTS(SELECT 1 FROM students WHERE id = $1)", p1ID).Scan(&exists1)
                    tx.QueryRow("SELECT EXISTS(SELECT 1 FROM students WHERE id = $1)", p2ID).Scan(&exists2)
                    log.Printf("Failed to increment collaboration for %d (exists: %t) and %d (exists: %t). Error: %v", p1ID, exists1, p2ID, exists2, err)
					return fmt.Errorf("error incrementing collaboration for students %d and %d in transaction: %v", p1ID, p2ID, err)
				}
			}
		}
	}

	return tx.Commit()
}


// getOrCreatePersonByNameTx is a helper function to get or create a person within an existing transaction.
func (ps *PostgresStorage) getOrCreatePersonByNameTx(tx *sql.Tx, name string) (*models.Person, error) {
	var person models.Person
	err := tx.QueryRow("SELECT id, name FROM students WHERE name = $1", name).Scan(&person.ID, &person.Name)
	if err == sql.ErrNoRows {
		// Person doesn't exist, create them
		errInsert := tx.QueryRow(`
			INSERT INTO students (name)
			VALUES ($1)
			RETURNING id, name
		`, name).Scan(&person.ID, &person.Name)
		if errInsert != nil {
			return nil, fmt.Errorf("error inserting student %s within transaction: %v", name, errInsert)
		}
		return &person, nil
	} else if err != nil {
		return nil, fmt.Errorf("error retrieving person by name %s within transaction: %v", name, err)
	}
	return &person, nil
}
