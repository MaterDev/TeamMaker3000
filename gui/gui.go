package gui

import (
	"log"
	"teammaker3000/models"
	"teammaker3000/storage" // To interact with the database
	"teammaker3000/data"    // For initial seed data

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"strconv" // For team size conversion
	"teammaker3000/core" // For team generation logic
)

type GuiApp struct {
	App fyne.App
	Win fyne.Window

	Storage storage.Storage

	// People list related
	People        binding.UntypedList // Holds *models.Person
	PeopleList    *widget.List
	NewPersonName *widget.Entry
	AddPersonBt   *widget.Button

	// Team generation related
	TeamSizeInput      *widget.Entry
	TeamBaseNameInput  *widget.Entry
	GenerateTeamsBt    *widget.Button
	SaveTeamsBt        *widget.Button
	GeneratedTeamsList *widget.List
	GeneratedTeamsData binding.UntypedList // Holds *models.Team

	// Store the last generated teams temporarily before saving
	lastGeneratedTeams []*models.Team

	// Saved Teams View related
	SavedTeamsData         binding.UntypedList // Holds *models.Team from DB
	SavedTeamsList         *widget.List
	SelectedTeamMembers    *widget.List // Shows members of the team selected in SavedTeamsList
	selectedTeamMembersData binding.StringList // Holds names of members for the SelectedTeamMembers list
	RefreshSavedTeamsBt    *widget.Button
}

func NewGuiApp(app fyne.App, win fyne.Window, store storage.Storage) *GuiApp {
	gui := &GuiApp{
		App:                     app,
		Win:                     win,
		Storage:                 store,
		People:                  binding.NewUntypedList(),
		GeneratedTeamsData:      binding.NewUntypedList(),
		SavedTeamsData:          binding.NewUntypedList(),
		selectedTeamMembersData: binding.NewStringList(), // Initialize the string list for members
	}
	gui.setupUi()       // Initialize UI components
	gui.loadInitialData() // Load people for "Manage People" tab
	gui.loadSavedTeams()  // Load teams for "Saved Teams" tab initially
	return gui
}

func (g *GuiApp) loadInitialData() {
	// Load people from DB
	people, err := g.Storage.GetPeople()
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to load people from database: %w", err), g.Win)
		return
	}

	if len(people) == 0 && len(data.Names) > 0 {
		// Database is empty, offer to seed from data.Names
		dialog.ShowConfirm("Seed Initial Data?", "The database is empty. Would you like to add the initial list of people?", func(confirm bool) {
			if confirm {
				for _, name := range data.Names {
					_, err := g.Storage.GetOrCreatePersonByName(name)
					if err != nil {
						log.Printf("Error seeding person %s: %v\n", name, err)
						// Optionally show an error for each failed seed
					}
				}
				// Reload after seeding
				g.refreshPeopleList()
			}
		}, g.Win)
	} else {
		for _, p := range people {
			g.People.Append(p)
		}
	}
}

func (g *GuiApp) refreshPeopleList() {
	// Clear current list
	currentItems, _ := g.People.Get()
	for i := len(currentItems) - 1; i >= 0; i-- {
		g.People.Remove(i)
	}

	// Load fresh from DB
	people, err := g.Storage.GetPeople()
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to refresh people list: %w", err), g.Win)
		return
	}
	for _, p := range people {
		g.People.Append(p)
	}
	g.PeopleList.Refresh()
}

func (g *GuiApp) setupUi() {
	// --- People Management Section ---
	g.PeopleList = widget.NewListWithData(
		g.People,
		func() fyne.CanvasObject {
			return widget.NewLabel("template person name") // Placeholder content
		},
		func(item binding.DataItem, obj fyne.CanvasObject) {
			untypedItem, err := item.(binding.Untyped).Get()
			if err != nil {
				log.Printf("Error getting untyped item: %v", err)
				obj.(*widget.Label).SetText("Error: Data error")
				return
			}
			person, ok := untypedItem.(*models.Person)
			if !ok {
				log.Printf("Error: item in PeopleList is not a *models.Person: %T, value: %+v", untypedItem, untypedItem)
				obj.(*widget.Label).SetText("Error: Invalid data type")
				return
			}
			obj.(*widget.Label).SetText(person.Name)
		},
	)

	g.NewPersonName = widget.NewEntry()
	g.NewPersonName.SetPlaceHolder("Enter new person's name")

	g.AddPersonBt = widget.NewButton("Add Person", func() {
		name := g.NewPersonName.Text
		if name == "" {
			dialog.ShowInformation("Input Error", "Person's name cannot be empty.", g.Win)
			return
		}
		newOrExistingPerson, err := g.Storage.GetOrCreatePersonByName(name)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to add person %s: %w", name, err), g.Win)
			return
		}
		g.NewPersonName.SetText("") // Clear input
		g.refreshPeopleList()     // Refresh the list from DB
		dialog.ShowInformation("Success", fmt.Sprintf("Person '%s' (ID: %d) added/ensured.", newOrExistingPerson.Name, newOrExistingPerson.ID), g.Win)
	})

	peopleManagementContent := container.NewBorder(
		nil, // top
		container.NewVBox(widget.NewLabel("Add New Person:"), g.NewPersonName, g.AddPersonBt), // bottom
		nil, // left
		nil, // right
		g.PeopleList, // center
	)
	peopleTab := container.NewTabItem("Manage People", peopleManagementContent)

	// --- Team Management Section ---
	g.TeamSizeInput = widget.NewEntry()
	g.TeamSizeInput.SetPlaceHolder("Enter desired team size (e.g., 3)")
	g.TeamBaseNameInput = widget.NewEntry()
	g.TeamBaseNameInput.SetPlaceHolder("Base name for teams (e.g., 'Pod')")
	g.TeamBaseNameInput.SetText("Team") // Default base name

	g.GeneratedTeamsList = widget.NewListWithData(
		g.GeneratedTeamsData,
		func() fyne.CanvasObject { // Item create
			return container.NewVBox(
				widget.NewLabel("Team Name Template"), // For team name
				widget.NewList( // For team members
					func() int { return 0 },
					func() fyne.CanvasObject { return widget.NewLabel("Member Name Template") },
					func(id widget.ListItemID, c fyne.CanvasObject) {},
				),
			)
		},
		func(item binding.DataItem, obj fyne.CanvasObject) { // Item update
			untypedTeamItem, _ := item.(binding.Untyped).Get()
			team, ok := untypedTeamItem.(*models.Team)
			if !ok {
				log.Printf("Error: item in GeneratedTeamsList is not a *models.Team: %T", untypedTeamItem)
				// Potentially set an error message on the UI element
				return
			}

			vbox := obj.(*fyne.Container)
			teamNameLabel := vbox.Objects[0].(*widget.Label)
			membersListWidget := vbox.Objects[1].(*widget.List)

			teamNameLabel.SetText(fmt.Sprintf("%s (ID: %s...)", team.Name, team.ID[:8])) // Show partial ID for reference

			membersListWidget.Length = func() int { return len(team.Members) }
			membersListWidget.CreateItem = func() fyne.CanvasObject { return widget.NewLabel("") }
			membersListWidget.UpdateItem = func(id widget.ListItemID, c fyne.CanvasObject) {
				if id < len(team.Members) {
					c.(*widget.Label).SetText(team.Members[id].Name)
				}
			}
			membersListWidget.Refresh()
		},
	)

	g.GenerateTeamsBt = widget.NewButton("Generate Teams", g.handleGenerateTeams)
	g.SaveTeamsBt = widget.NewButton("Save Generated Teams", g.handleSaveTeams)
	g.SaveTeamsBt.Disable() // Disabled until teams are generated

	teamGenControls := container.NewVBox(
		widget.NewLabel("Team Generation Settings:"),
		g.TeamSizeInput,
		g.TeamBaseNameInput,
		g.GenerateTeamsBt,
	)

	generatedTeamsDisplay := container.NewScroll(g.GeneratedTeamsList) // Make list scrollable

	teamManagementContent := container.NewBorder(
		teamGenControls, // Top: controls for generation
		g.SaveTeamsBt,   // Bottom: save button
		nil,             // Left
		nil,             // Right
		generatedTeamsDisplay, // Center: list of generated teams
	)
	teamsTab := container.NewTabItem("Manage Teams", teamManagementContent)

	// --- Main Tabs ---
	tabs := container.NewAppTabs(
		peopleTab,
		teamsTab,
		g.createSavedTeamsTab(), // Add the new tab
	)

	g.Win.SetContent(tabs)
	// g.Win.Resize(fyne.NewSize(600, 400)) // Size is set in main.go
}

func (g *GuiApp) createSavedTeamsTab() *container.TabItem {
	g.SavedTeamsList = widget.NewListWithData(
		g.SavedTeamsData,
		func() fyne.CanvasObject { // CreateItem
			return widget.NewLabel("Team Name (ID)")
		},
		func(item binding.DataItem, obj fyne.CanvasObject) { // UpdateItem
			untypedTeam, _ := item.(binding.Untyped).Get()
			team, ok := untypedTeam.(*models.Team)
			if !ok {
				obj.(*widget.Label).SetText("Error: Invalid team data")
				log.Printf("Error: item in SavedTeamsList is not *models.Team: %T", untypedTeam)
				return
			}
			obj.(*widget.Label).SetText(fmt.Sprintf("%s (ID: ...%s)", team.Name, team.ID[len(team.ID)-6:]))
		},
	)

	g.SelectedTeamMembers = widget.NewListWithData(
		g.selectedTeamMembersData,
		func() fyne.CanvasObject { return widget.NewLabel("Member Name") },
		func(item binding.DataItem, obj fyne.CanvasObject) {
			memberName, _ := item.(binding.String).Get()
			obj.(*widget.Label).SetText(memberName)
		},
	)

	g.SavedTeamsList.OnSelected = func(id widget.ListItemID) {
		g.selectedTeamMembersData.Set(nil) // Clear previous members
		item, err := g.SavedTeamsData.GetValue(int(id))
		if err != nil {
			log.Printf("Error getting selected saved team: %v", err)
			return
		}
		team, ok := item.(*models.Team)
		if !ok {
			log.Printf("Error: selected item is not *models.Team: %T", item)
			return
		}
		var memberNames []string
		for _, member := range team.Members {
			memberNames = append(memberNames, member.Name)
		}
		g.selectedTeamMembersData.Set(memberNames)
	}
	g.SavedTeamsList.OnUnselected = func(id widget.ListItemID) {
		g.selectedTeamMembersData.Set(nil) // Clear members when nothing is selected
	}


	g.RefreshSavedTeamsBt = widget.NewButton("Refresh List", g.loadSavedTeams)

	membersContainer := container.NewBorder(
		widget.NewLabel("Selected Team Members:"), nil, nil, nil,
		container.NewScroll(g.SelectedTeamMembers),
	)

	// Make the list of saved teams take up more vertical space
	savedTeamsListContainer := container.NewScroll(g.SavedTeamsList)


	mainContent := container.NewHSplit(
		savedTeamsListContainer, // Left side: list of saved teams
		membersContainer,        // Right side: members of selected team
	)
	mainContent.SetOffset(0.4) // Adjust split ratio: 40% for team list, 60% for members

	return container.NewTabItem("Saved Teams", container.NewBorder(
		g.RefreshSavedTeamsBt, // Top: Refresh button
		nil, nil, nil,
		mainContent,
	))
}


func (g *GuiApp) loadSavedTeams() {
	teams, err := g.Storage.GetAllTeams()
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to load saved teams: %w", err), g.Win)
		return
	}

	g.SavedTeamsData.Set(nil) // Clear current list
	for _, team := range teams {
		g.SavedTeamsData.Append(team)
	}
	g.SavedTeamsList.Refresh()
	g.selectedTeamMembersData.Set(nil) // Clear any selected member list
	if len(teams)>0 {
		log.Printf("Loaded %d saved teams successfully.", len(teams))
	}
}


// Placeholder for later use if needed
// func (g *GuiApp) Show() {
// 	g.Win.ShowAndRun()
// }

// Note: main.go will call app.New() and win.NewWindow()
// and then pass them to NewGuiApp.
// The main.go will also call win.ShowAndRun() after setting up the GuiApp.
// This GuiApp struct will then primarily manage the content of the window.

// --- Handler Methods ---

func (g *GuiApp) handleGenerateTeams() {
	teamSizeStr := g.TeamSizeInput.Text
	teamSize, err := strconv.Atoi(teamSizeStr)
	if err != nil || teamSize <= 0 {
		dialog.ShowError(fmt.Errorf("invalid team size: please enter a positive number"), g.Win)
		return
	}

	teamBaseName := g.TeamBaseNameInput.Text
	if teamBaseName == "" {
		dialog.ShowInformation("Input Required", "Please enter a base name for the teams (e.g., 'Project Group').", g.Win)
		return
	}

	// Fetch current people from the bound list (which should be up-to-date from DB)
	// Or, fetch fresh from DB to be absolutely sure
	currentPeopleModels := []*models.Person{}
	peopleInterface, _ := g.People.Get() // This gets a []interface{}
	for _, pInterface := range peopleInterface {
		person, ok := pInterface.(*models.Person)
		if !ok {
			log.Printf("Error: Found non-*models.Person data in g.People: %T", pInterface)
			dialog.ShowError(fmt.Errorf("internal data error: people list contains invalid data"), g.Win)
			return
		}
		currentPeopleModels = append(currentPeopleModels, person)
	}

	if len(currentPeopleModels) == 0 {
		dialog.ShowInformation("No People", "There are no people to form teams. Please add people first.", g.Win)
		return
	}

	// Basic validation for team size vs number of people
	if teamSize > len(currentPeopleModels) {
		dialog.ShowError(fmt.Errorf("team size (%d) cannot be greater than the number of available people (%d)", teamSize, len(currentPeopleModels)), g.Win)
		return
	}
    // Add the check for teams of 1: studentCount % teamSize == 1 (and studentCount > teamSize)
    if len(currentPeopleModels) > teamSize && len(currentPeopleModels)%teamSize == 1 {
        dialog.ShowWarning("Team Formation Warning",
            fmt.Sprintf("With %d people and team size %d, one person will be left alone. Consider adjusting team size or adding/removing a person.", len(currentPeopleModels), teamSize),
            g.Win)
        // We can choose to proceed or not. For now, we proceed and let core.GenerateTeams handle it.
    }


	generated, err := core.GenerateTeams(g.Storage, currentPeopleModels, teamSize, teamBaseName)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to generate teams: %w", err), g.Win)
		g.SaveTeamsBt.Disable()
		g.lastGeneratedTeams = nil
		return
	}

	if len(generated) == 0 {
		dialog.ShowInformation("No Teams Generated", "Team generation resulted in no teams. This might be due to the number of people and team size.", g.Win)
		g.SaveTeamsBt.Disable()
		g.lastGeneratedTeams = nil
		return
	}

	g.lastGeneratedTeams = generated // Store for saving

	// Update the GeneratedTeamsData binding list
	g.GeneratedTeamsData.Set(nil) // Clear previous
	for _, team := range generated {
		g.GeneratedTeamsData.Append(team)
	}
	g.GeneratedTeamsList.Refresh()
	g.SaveTeamsBt.Enable()
	dialog.ShowInformation("Success", fmt.Sprintf("%d teams generated.", len(generated)), g.Win)
}

func (g *GuiApp) handleSaveTeams() {
	if len(g.lastGeneratedTeams) == 0 {
		dialog.ShowInformation("No Teams", "No teams have been generated yet to save.", g.Win)
		return
	}

	err := g.Storage.SaveGeneratedTeamsAndCollaborations(g.lastGeneratedTeams)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to save teams: %w", err), g.Win)
		return
	}

	dialog.ShowInformation("Success", "Generated teams and collaboration data saved successfully!", g.Win)
	g.SaveTeamsBt.Disable() // Disable after saving to prevent re-saving same set without re-generation
	g.lastGeneratedTeams = nil // Clear buffer
	// Optionally clear the display of generated teams or leave them for review
	// g.GeneratedTeamsData.Set(nil)
	// g.GeneratedTeamsList.Refresh()
}


// Make sure to import "fmt"
import "fmt"
