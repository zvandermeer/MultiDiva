# MultiDiva

WIP Experimental multiplayer support for Hatsune Miku: Project Diva MegaMix+

Server-side code can be found [here](https://github.com/zvandermeer/MultiDiva-Server)

## Build instructions

 ### MultiDiva-Loader.dll (C++)

 - Install [Microsoft Visual Studio](https://visualstudio.microsoft.com/)
 - Open "MultiDiva-Loader/MultiDiva.sln" in Visual Studio
 - Build the solution
 - The compiled .DLL can be found in the 'x64/' directory

### MultiDiva-Client.dll (Go)

- Install the [Go programming language](https://go.dev/doc/install)
- Open the "MultiDiva-Client" folder in your terminal and run the following:

	`go mod tidy`

	`go build -o ./bin/MultiDiva-Client.dll -buildmode=c-shared`

- The compiled .DLL can be found in the 'bin/' directory
