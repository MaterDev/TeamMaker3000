# Team Maker 3000 - GUI Edition

Team Maker 3000 is a desktop application built with Go and the Fyne GUI toolkit. It helps you organize people (e.g., students, employees) into teams, aiming to minimize repeat collaborations. The application allows you to manage lists of people, generate teams based on specified sizes, save these teams, and view collaboration history. Teams are assigned unique identifiers for potential analytics purposes.

<!-- Placeholder for new GUI screenshot -->
<!-- ![Team Maker 3000 GUI](./new_cover_image.png) -->
*Image: Placeholder for a screenshot of the main application window.*

## Features

*   **Manage People**: Add new individuals to a persistent list stored in a PostgreSQL database.
*   **Team Generation**: Create teams from the list of available people, specifying team size and a base name. The algorithm attempts to minimize repeat collaborations.
*   **Save Teams**: Persist generated team compositions to the database, including unique team IDs. Collaboration counts between members are automatically updated.
*   **View Saved Teams**: Browse previously saved teams and view their members.
*   **Data Persistence**: All data (people, teams, collaborations) is stored in a PostgreSQL database.
*   **Cross-Platform**: Built with Fyne, aiming for compatibility across Windows, macOS, and Linux.

## Getting Started

### Prerequisites

*   **Go**: Version 1.18 or higher.
*   **Git**: For cloning the repository.
*   **PostgreSQL**: A running PostgreSQL server (version 9.6+ recommended).
*   **C Compiler & System Libraries for Fyne**: Fyne requires a C compiler (like GCC or Clang) and certain development libraries.
    *   **Linux**: `gcc` and X11 development libraries (e.g., `libx11-dev`, `libgl1-mesa-dev`, `xorg-dev` - package names may vary by distribution).
    *   **macOS**: Xcode Command Line Tools.
    *   **Windows**: A GCC compiler like the one provided by MSYS2 or TDM-GCC.
    *   For detailed, OS-specific Fyne prerequisites, please refer to the [Fyne documentation](https://developer.fyne.io/started/).

### Installation

1.  **Clone the repository**:

   ```sh
   git clone https://github.com/materdev/teammaker3000.git
   cd teammaker3000
   ```

2.  **Initialize Go module and fetch dependencies**:
    ```sh
    go mod tidy
    ```
    This will download Fyne and other necessary packages.

### Setting Up PostgreSQL Database

1.  **Install PostgreSQL**: If you haven't already, install PostgreSQL. Follow the official instructions for your operating system.

2.  **Create Database and User (Recommended)**:
    It's good practice to create a dedicated user and database. Connect to PostgreSQL as a superuser (e.g., `postgres`):
    ```sh
    psql -U postgres
    ```
    Then, execute the following SQL commands:
    ```sql
    -- Optional: Create a dedicated user (replace 'your_password' with a strong password)
    -- CREATE USER teammaker_user WITH PASSWORD 'your_password';

    CREATE DATABASE teammaker3000;

    -- Optional: Grant privileges to the dedicated user
    -- GRANT ALL PRIVILEGES ON DATABASE teammaker3000 TO teammaker_user;

    \c teammaker3000
    ```
    *Note: The application's `main.go` currently uses the `postgres` user by default. You may need to adjust the `dbUser` constant or the connection string in `main.go` if you use a different user or require a password not handled by local authentication (e.g., `.pgpass` file).*

3.  **Create Tables**:
    Connect to your `teammaker3000` database (e.g., `psql -U postgres -d teammaker3000` or `psql -U teammaker_user -d teammaker3000` if you created a user) and run the following SQL commands. The application will also attempt to create these tables if they don't exist on startup via `InitializeSchema()`.

    ```sql
    CREATE TABLE IF NOT EXISTS students (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) UNIQUE NOT NULL
    );

    CREATE TABLE IF NOT EXISTS collaborations (
        id SERIAL PRIMARY KEY,
        student1_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
        student2_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
        collaborations_count INTEGER DEFAULT 0, -- Changed default to 0, incremented on first collab
        UNIQUE(student1_id, student2_id),
        CONSTRAINT check_student_order CHECK (student1_id < student2_id) -- Ensures unique pairs
    );

    CREATE TABLE IF NOT EXISTS teams (
        team_id VARCHAR(36) PRIMARY KEY, -- UUID
        team_name VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS team_members (
        team_id VARCHAR(36) NOT NULL REFERENCES teams(team_id) ON DELETE CASCADE,
        student_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
        PRIMARY KEY (team_id, student_id)
    );
    ```

### Running the Application

To run the application, execute the following command in the project root directory:

```sh
go run main.go
```

This will compile and run the Fyne GUI application.

## Using the Application

The application window has three main tabs:

### 1. Manage People

*   **View People**: Displays a list of all people currently in the database.
*   **Add Person**:
    *   Enter the person's name in the input field at the bottom.
    *   Click "Add Person". The person will be added to the database, and the list will refresh.
*   **Initial Data Seeding**: If the database is empty when the application starts, a dialog will prompt you to seed the database with an initial list of names (from `data/data.go`).

<!-- Placeholder for Manage People tab screenshot -->
<!-- ![Manage People Tab](./manage_people_tab.png) -->
*Image: Placeholder for a screenshot of the 'Manage People' tab.*

### 2. Manage Teams (Team Generation)

*   **Team Generation Settings**:
    *   **Team Size**: Enter the desired number of people per team.
    *   **Base Name**: Enter a base name for the generated teams (e.g., "Pod", "Project Group"). Teams will be named like "Pod 1", "Pod 2".
*   **Generate Teams Button**:
    *   Click this to generate teams based on the current list of people and settings.
    *   The generated teams will appear in the list below.
    *   The algorithm attempts to minimize repeat collaborations based on saved history.
*   **Generated Teams Display**: Shows the generated teams and their members.
*   **Save Generated Teams Button**:
    *   This button becomes active after teams are generated.
    *   Click to save the currently displayed generated teams to the database. This also updates collaboration counts between members of these new teams.

<!-- Placeholder for Manage Teams tab screenshot -->
<!-- ![Manage Teams Tab](./manage_teams_tab.png) -->
*Image: Placeholder for a screenshot of the 'Manage Teams' tab showing generated teams.*

### 3. Saved Teams

*   **View Saved Teams**: Displays a list of all teams that have been previously saved to the database (shows team name and part of its unique ID).
*   **View Team Members**: Click on a team in the list to see its members displayed on the right side.
*   **Refresh List Button**: Click to reload the list of saved teams from the database.

<!-- Placeholder for Saved Teams tab screenshot -->
<!-- ![Saved Teams Tab](./saved_teams_tab.png) -->
*Image: Placeholder for a screenshot of the 'Saved Teams' tab.*

## Application Structure (Brief Overview)

*   `main.go`: Entry point, initializes database, Fyne app, and main GUI structure.
*   `gui/`: Contains Fyne GUI layout and logic.
    *   `gui.go`: Defines the main `GuiApp` struct and its methods for UI setup and event handling.
*   `storage/`: Handles data persistence.
    *   `storage.go`: Defines the `Storage` interface.
    *   `postgres_storage.go`: Implements the `Storage` interface using PostgreSQL.
*   `core/`: Contains core business logic.
    *   `logic.go`: Implements the team generation algorithm (`GenerateTeams`).
*   `models/`: Defines data structures (`Person`, `Team`).
*   `data/`: Contains initial seed data (e.g., `data.Names`).

## License

This project is licensed under the MDGUL License - see the [LICENSE](LICENSE) file for details.

This project is licensed under the MDGUL License - see the [LICENSE](LICENSE) file for details.
```

This README includes instructions to set up the PostgreSQL database, run the application, and details about the application's functionality.