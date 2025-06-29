package core

import (
	"fmt"
	"sort"
	"teammaker3000/models"
	"teammaker3000/storage"
	"math/rand"
	"time"
)

// GenerateTeams assigns people to teams aiming to minimize repeat collaborations.
// It takes a list of all available people, the desired team size, and a base name for the generated teams.
func GenerateTeams(st storage.Storage, allPeople []*models.Person, teamSize int, teamBaseName string) ([]*models.Team, error) {
	if len(allPeople) == 0 {
		return nil, fmt.Errorf("no people available to form teams")
	}
	if teamSize <= 0 {
		return nil, fmt.Errorf("team size must be greater than 0")
	}
	if teamSize == 1 && len(allPeople) > 1 { // Allow team size of 1 if only one person
		// This check is tricky. The original CLI prevented teamSize that results in a remainder of 1.
		// For GUI, we might allow it but warn user or adjust.
		// For now, let's assume teamSize has been validated by UI to avoid lone members if total > teamSize.
	}


	// Create a modifiable copy of allPeople to track assignments
	availablePeople := make([]*models.Person, len(allPeople))
	copy(availablePeople, allPeople)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(availablePeople), func(i, j int) { availablePeople[i], availablePeople[j] = availablePeople[j], availablePeople[i] })


	// 1. Calculate collaboration scores for all possible pairs
	type personPair struct {
		person1    *models.Person
		person2    *models.Person
		collabCount int
	}
	var pairs []personPair

	for i := 0; i < len(availablePeople); i++ {
		for j := i + 1; j < len(availablePeople); j++ {
			p1 := availablePeople[i]
			p2 := availablePeople[j]
			count, err := st.GetCollaborationCount(p1.ID, p2.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get collaboration count for %s and %s: %v", p1.Name, p2.Name, err)
			}
			pairs = append(pairs, personPair{p1, p2, count})
		}
	}

	// Sort pairs by collaboration count (ascending), then randomly to break ties
	sort.SliceStable(pairs, func(i, j int) bool {
		if pairs[i].collabCount != pairs[j].collabCount {
			return pairs[i].collabCount < pairs[j].collabCount
		}
		return rand.Intn(2) == 0 // Random tie-breaking
	})

	// 2. Form teams
	// This logic is complex and needs careful implementation.
	// The original logic was: iterate through sorted pairs and add them to current pod if not assigned and pod not full.
	// Then assign remaining unassigned people.
	// This can be suboptimal. A better approach might involve a weighted graph or more sophisticated clustering.
	// For now, let's adapt the simpler approach and acknowledge its limitations.

	var generatedTeams []*models.Team
	assigned := make(map[int]bool) // Person.ID -> assigned
	numPeople := len(availablePeople)
	numGeneratedTeams := (numPeople + teamSize - 1) / teamSize

	// Ensure no team of 1 is formed if avoidable (original logic)
	if numPeople > teamSize && numPeople % teamSize == 1 {
		// This scenario should ideally be handled by GUI validation of teamSize or by adjusting team sizes dynamically.
		// For example, make one or more teams larger by 1.
		// For now, we'll proceed, but the last team might be small.
		// A more robust solution would adjust team sizes, e.g., if 7 people and teamSize=3, make teams of 4 and 3.
	}


	// Initialize empty teams
	for i := 0; i < numGeneratedTeams; i++ {
		teamName := fmt.Sprintf("%s %d", teamBaseName, i+1)
		generatedTeams = append(generatedTeams, models.NewTeam(teamName, []*models.Person{}))
	}

	// Distribute people, prioritizing those with fewest collaborations with potential teammates
	// This is a greedy approach: try to fill teams one by one.

	// Attempt 1: Simple distribution based on sorted pairs (similar to original)
	// This is difficult to adapt directly while ensuring good distribution.

	// Attempt 2: Iterative assignment to teams, trying to balance low collaboration scores.
	// This is still greedy and might not be optimal.

	// Let's use a simpler greedy assignment for now:
	// Iterate through people (shuffled) and assign them to the team that "needs" them most
	// or has the lowest total collaboration score with them.

	// Simpler approach: Iterate through people and assign to teams sequentially, then try to optimize.
	// Or, even simpler: just distribute them. The "minimal collaboration" part is tricky with a pure greedy approach.

	// Let's try to fill teams using the sorted pairs, but be mindful of team balance.
	// This is a hard problem (variant of bin packing or graph partitioning).

	// Backtrack to a simpler version of the original logic for now, then refine.
	// The original logic iterates through sorted pairs and assigns them.
	// This doesn't distribute well into multiple teams simultaneously.

	// New strategy:
	// 1. Create empty teams.
	// 2. Iteratively add people to teams. For each person, choose the team where they add the "least conflict"
	//    (sum of collaborations with existing members of that team).
	// This is still greedy.

	currentPersonIndex := 0
	for _, person := range availablePeople { // Iterate through shuffled people
		if assigned[person.ID] {
			continue
		}

		bestTeamIndex := -1
		minAddedConflict := -1 // Using -1 to indicate not yet calculated, as 0 is a valid conflict score

		// Find the best team for this person
		for i, team := range generatedTeams {
			if len(team.Members) < teamSize {
				currentConflict := 0
				for _, member := range team.Members {
					count, err := st.GetCollaborationCount(person.ID, member.ID)
					if err != nil {
						return nil, fmt.Errorf("error checking conflict for %s in team %s: %v", person.Name, team.Name, err)
					}
					currentConflict += count
				}

				if bestTeamIndex == -1 || currentConflict < minAddedConflict {
					minAddedConflict = currentConflict
					bestTeamIndex = i
				} else if currentConflict == minAddedConflict {
					// Tie-breaking: prefer smaller teams
					if len(generatedTeams[i].Members) < len(generatedTeams[bestTeamIndex].Members) {
						bestTeamIndex = i
					}
				}
			}
		}

		if bestTeamIndex != -1 {
			generatedTeams[bestTeamIndex].Members = append(generatedTeams[bestTeamIndex].Members, person)
			assigned[person.ID] = true
			currentPersonIndex++
		} else {
			// This should not happen if numGeneratedTeams is calculated correctly and teamSize is valid
			// unless all teams are full but there are still unassigned people (logic error in numGeneratedTeams or team full check)
			// Or, if teamSize is very large relative to numPeople.
			// For now, if this happens, try to add to any team that's not yet at max theoretical size for rebalancing.
			// This part is a fallback and indicates a potential flaw in simple greedy.
			foundSpot := false
			for i, team := range generatedTeams {
				// Allow slight overfilling if it means not leaving someone out, rebalance later.
				// Max size could be teamSize + 1 if we need to absorb remainders.
				// This needs a more robust rebalancing step.
				if len(team.Members) < teamSize+1 { // A bit of leeway
					team.Members = append(team.Members, person)
					assigned[person.ID] = true
					currentPersonIndex++
					foundSpot = true
					break
				}
			}
			if !foundSpot {
				// If still no spot, this is an issue. Maybe create a new "overflow" team.
				// For now, this indicates an unhandled edge case or flaw in the team size validation / distribution.
				// This could happen if teamSize validation (no person left alone) isn't perfectly handled before this function.
				return nil, fmt.Errorf("could not assign %s to any team; team distribution logic error or capacity issue", person.Name)
			}
		}
	}

	// Post-processing: Handle teams that might be too small, especially the last one.
	// If the last team is too small (e.g., 1 person) and there are other teams, try to re-distribute.
	// The original CLI had validation for `studentCount % teamSize == 1`.
	// This logic is complex and depends on how strictly "teamSize" must be adhered to.
	// Example: 7 people, teamSize = 3. Forms teams of 3, 3, 1.
	// Ideal: 3, 2, 2 or adjust teamSize to e.g. 4,3
	// For now, we accept potentially small last team. GUI can warn or guide user on teamSize.

	finalTeams := []*models.Team{}
	for _, t := range generatedTeams {
		if len(t.Members) > 0 { // Only include teams that have members
			finalTeams = append(finalTeams, t)
		}
	}

	return finalTeams, nil
}
