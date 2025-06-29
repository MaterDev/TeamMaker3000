package main

import (
	"database/sql"
	"fmt"
	"log"
	"teammaker3000/data" // For initial data constants if used by GUI/Storage for seeding
	"teammaker3000/gui"
	"teammaker3000/storage"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	_ "github.com/lib/pq" // PostgreSQL driver
)

const (
	dbHost = "localhost" // Consider making these configurable (e.g., via env vars or config file)
	dbPort = 5432
	dbUser = "postgres" // Default user, ensure this is correct for your setup
	// dbPass = "your_password" // It's better to use pgpass file or environment variables for passwords
	dbName = "teammaker3000"
)

func main() {
	// --- Database Setup ---
	// Construct connection string
	// For production, avoid hardcoding passwords. Use environment variables or a .pgpass file.
	// Example connStr: "postgres://username:password@localhost:5432/dbname?sslmode=disable"
	// connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
	// 	dbHost, dbPort, dbUser, dbPass, dbName)
	// Simpler connection string if password is via .pgpass or trust auth:
	connStr := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbName)


	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v. Ensure PostgreSQL is running and accessible with conn string: %s", err, connStr)
	}
	log.Println("Successfully connected to PostgreSQL database.")

	// Initialize storage
	postgresStore := storage.NewPostgresStorage(db)

	// Initialize database schema (creates tables if they don't exist)
	if err := postgresStore.InitializeSchema(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}
	log.Println("Database schema checked/initialized.")

	// --- GUI Setup ---
	fyneApp := app.New()
	// Consider setting a unique AppID for preferences, especially on macOS/Linux
	// fyneApp.SetUniqueID("io.github.youruser.teammaker3000")


	mainWindow := fyneApp.NewWindow("Team Maker 3000")
	mainWindow.SetMaster() // Sets this as the main window, app exits when closed.

	// Initialize the GuiApp structure, which sets up its own UI
	// The GuiApp's setupUi method will call mainWindow.SetContent()
	gui.NewGuiApp(fyneApp, mainWindow, postgresStore) // GuiApp constructor handles UI setup and data loading

	// For initial data seeding (if data.Names is used and DB is empty)
	// This logic is now inside gui.NewGuiApp -> loadInitialData
	// We could also have a specific admin/setup screen for this in a more complex app.
	if len(data.Names) > 0 { // Check if there's any seed data defined
		people, err := postgresStore.GetPeople()
		if err == nil && len(people) == 0 {
			log.Println("Database is empty. Initial data seeding might be offered by the GUI if configured.")
		}
	}

	mainWindow.Resize(fyne.NewSize(800, 600)) // Default size
	mainWindow.ShowAndRun()

	log.Println("Application exited.")
}
