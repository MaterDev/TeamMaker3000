# Team Maker 3000

Team Maker 3000 is a Go application that randomly assigns students to different pods. This application is useful for organizing students into groups (or pods) for projects, discussions, or activities.

![Team Maker 3000](./coverImage.png)

## Project Structure

```
your_project_name/
├── dataReal/
│   └── dataReal.go
├── main.go
├── student.go
└── go.mod
```

- **dataReal/dataReal.go**: Contains the list of student names.
- **main.go**: The main entry point of the application.
- **student.go**: Defines the `Student` interface and the `student` struct.

## Getting Started

### Prerequisites

- Go 1.18 or higher installed on your machine.
- Git installed on your machine.

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

### Running the Application

To run the application, execute the following command in the project root directory:

```sh
go run .
```

## Application Details

The application performs the following tasks:

1. **Read Student Names**:
   - The application reads a list of student names from the `dataReal` package.

2. **Create Students**:
   - It creates student instances from the list of names.

3. **Assign Pods**:
   - The application randomly assigns students to a specified number of pods.

4. **Display Pods**:
   - It displays the number of students and lists the students in each pod.

### Example Output

![Example Output](./exampleOutput.png)

<!-- Hideable code block -->
<details>
<summary>Click to expand the example output text</summary>

```sh
Number of students: 15
Pod 1:
    Alex Ryan
    Quique Jefferson
    Salaam Muhammad
    Axe Rivera
Pod 2:
    Fernando Valdez
    Nija Desirae
    Braeden Kincade
    Zander Ofosu
Pod 3:
    Javier Rice
    Chava Rosenwort
    Peng Zhang
    Wendy Daye
Pod 4:
    Nicolas Sosa
    TacoCat Dogod
    Locke Lamora
```
</details>


## License

This project is licensed under the MDGUL License - see the [LICENSE](LICENSE) file for details.
