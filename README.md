# MultiDiva

Experimental multiplayer and custom leaderboard support for Hatsune Miku: Project Diva MegaMix+

## Build instructions

 ### MultiDiva-Loader.dll (C++)

 - Install [Microsoft Visual Studio](https://visualstudio.microsoft.com/) or [Jetbrains Rider](https://www.jetbrains.com/rider/)
 - Open "MultiDiva-Loader/MultiDiva.sln" in your .NET IDE of choice
 - Build the solution
 - The compiled .DLL can be found in the 'x64/' directory

### MultiDiva-Client.dll (Go)

- Install the [Go programming language](https://go.dev/doc/install)
- Open the "MultiDiva-Client" folder in your terminal and run the following:

	`go mod tidy`

	`go build -o ./bin/MultiDiva-Client.dll -buildmode=c-shared  ./cmd/MultiDiva-Client`

- The compiled .DLL can be found in the 'bin/' directory
