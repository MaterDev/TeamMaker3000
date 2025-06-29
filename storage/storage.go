package storage

import (
	"database/sql"
	"teammaker3000/models"
)

// Storage defines the interface for data persistence.
type Storage interface {
	// People methods
	GetPeople() ([]*models.Person, error)
	AddPerson(name string) (*models.Person, error) // Returns the created person with their ID
	GetPersonByID(id int) (*models.Person, error)
	GetOrCreatePersonByName(name string) (*models.Person, error) // Retrieves a person by name or creates if not exists

	// Team methods
	SaveTeam(team *models.Team) error // Saves a complete team object, including its members
	GetTeamByID(teamID string) (*models.Team, error)
	GetAllTeams() ([]*models.Team, error) // Gets all saved teams with their members

	// Collaboration methods
	GetCollaborationCount(person1ID, person2ID int) (int, error)
	IncrementCollaborationCount(person1ID, person2ID int) error
	// SaveGeneratedTeamsAndCollaborations is a higher-level function that will:
	// 1. Save each team to the 'teams' table (if new)
	// 2. Save team memberships to 'team_members' table
	// 3. Increment collaboration counts between all members of each newly formed team
	SaveGeneratedTeamsAndCollaborations(teams []*models.Team) error

	// Database Initialization
	InitializeSchema() error // Ensures DB schema (students, collaborations, teams, team_members) is set up
}

// PostgresStorage implements the Storage interface for PostgreSQL.
// (This struct definition might move to postgres_storage.go if preferred,
// but often the interface and its primary implementation struct are kept together for clarity)
type PostgresStorage struct {
	DB *sql.DB
}

// NewPostgresStorage creates a new PostgresStorage instance.
func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{DB: db}
}

// Compile-time check to ensure PostgresStorage implements Storage
var _ Storage = (*PostgresStorage)(nil)
