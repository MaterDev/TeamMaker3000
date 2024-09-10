# Team Maker 3000

Team Maker 3000 is a Go application that assigns students to different pods based on user-defined minimum and target team sizes. This application is useful for organizing students into groups (or pods) for projects, discussions, or activities, while also ensuring minimal repeat collaborations and maintaining balanced group sizes.

![Team Maker 3000](./coverImage.png)

## Features

- **Minimum and Target Team Sizes**: Users can specify both a minimum and target team size, providing flexibility in group formation.
- **Smart Redistribution**: Ensures all teams meet the minimum size requirement, even if it means exceeding the target size for some groups.
- **Collaboration History**: Uses past collaboration data to minimize repeat pairings.
- **Transparent Process**: Provides clear feedback when redistributions occur.
- **Database Integration**: Stores and retrieves collaboration data for long-term tracking.
- **Concurrent Processing**: Optimized for performance with larger datasets.

## Getting Started

### Prerequisites

- Go 1.18 or higher installed on your machine.
- Git installed on your machine.
- PostgreSQL installed and running.

### Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/materdev/teammaker3000.git
   cd teammaker3000
   ```

2. Initialize the Go module:

   ```sh
   go mod tidy
   ```

### Setting Up PostgreSQL Database

1. **Install PostgreSQL**: Follow the instructions for your operating system to install PostgreSQL.

2. **Create Database**: Create a new database named `teammaker3000`:

   ```sh
   psql -U postgres
   CREATE DATABASE teammaker3000;
   \c teammaker3000;
   ```

3. **Create Tables**: Run the following SQL commands to create the required tables:

   ```sql
   CREATE TABLE students (
       id SERIAL PRIMARY KEY,
       name VARCHAR(255) UNIQUE NOT NULL
   );

   CREATE TABLE collaborations (
       id SERIAL PRIMARY KEY,
       student1_id INTEGER NOT NULL REFERENCES students(id),
       student2_id INTEGER NOT NULL REFERENCES students(id),
       collaborations_count INTEGER DEFAULT 1,
       UNIQUE(student1_id, student2_id)
   );
   ```

### Running the Application

To run the application, execute the following command in the project root directory:

```sh
go run .
```

### Using the Application

1. **Enter Team Sizes**: 
   - When prompted, enter the minimum team size.
   - Then, enter the target team size (must be equal to or greater than the minimum).

2. **Review Teams**: 
   - The application will display the generated teams.
   - If redistributions occurred to meet the minimum size requirement, a message will explain the changes.

3. **Confirm Teams**: 
   - Enter 'A' to accept the teams and save to the database.
   - Enter 'X' to exit without saving.

### Application Details

The application performs the following tasks:

1. **Read Student Names**:
   - The application reads a list of student names from the `data` package.

2. **Create Students**:
   - It creates student instances from the list of names.

3. **Fetch Collaboration History**:
   - Retrieves past collaboration data from the database.

4. **Assign Pods**:
   - The application assigns students to pods based on the minimum and target team sizes, ensuring minimal repeat collaborations.

5. **Redistribute if Necessary**:
   - If any team is smaller than the minimum size, students are redistributed to meet this requirement.

6. **Display Pods**:
   - It displays the number of students and lists the students in each pod.
   - If redistribution occurred, it explains why and how many students were moved.

7. **Save Collaborations**:
   - Upon user confirmation, the application saves the new collaboration data to the PostgreSQL database.

### Example Output

```sh
Enter minimum team size: 3
Enter target team size: 4
Number of students: 15
Pod 1 (4 members):
        Alex Ryan
        Fernando Valdez
        Javier Rice
        Quique Jefferson
Pod 2 (4 members):
        Nija Desirae
        Chava Rosenwort
        Salaam Muhammad
        Braeden Kincade
Pod 3 (4 members):
        Peng Zhang
        Axe Rivera
        Zander Ofosu
        Nicolas Sosa
Pod 4 (3 members):
        Wendy Daye
        TacoCat Dogod
        Locke Lamora
The target team size of 4 was not possible for all teams. 3 student(s) from the smallest team have been redistributed to meet the minimum team size of 3.
Enter 'A' to accept or 'X' to exit without saving: A
Teams saved to database.
```

## Configuration

You can modify the following constants in the `main.go` file to adjust the application's behavior:

- `dbHost`: Database host (default: "localhost")
- `dbPort`: Database port (default: 5432)
- `dbName`: Database name (default: "teammaker3000")
- `maxRetries`: Maximum number of database connection retries (default: 3)
- `batchSize`: Number of database operations to perform in a single transaction (default: 1000)

## License

This project is licensed under the MDGUL License - see the [LICENSE](LICENSE) file for details.