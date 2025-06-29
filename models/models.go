package models

import "github.com/google/uuid"

// Person represents an individual in the system.
type Person struct {
	ID   int    `json:"id"`   // Database ID
	Name string `json:"name"` // Name of the person
}

// Team represents a group of people.
type Team struct {
	ID        string    `json:"id"`        // Unique identifier for the team (UUID)
	Name      string    `json:"name"`      // Name of the team (e.g., "Pod 1", "Project Alpha")
	Members   []*Person `json:"members"`   // List of people in the team
	ProjectID string    `json:"project_id,omitempty"` // Optional: Identifier for a project this team is associated with
}

// NewTeam creates a new team with a unique ID.
func NewTeam(name string, members []*Person) *Team {
	return &Team{
		ID:      uuid.NewString(),
		Name:    name,
		Members: members,
	}
}

// Collaboration represents the history of two people working together.
// This might be simplified or absorbed into team save logic later.
type Collaboration struct {
	Person1ID         int
	Person2ID         int
	CollaborationCount int
}
