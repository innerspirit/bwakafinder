# BW AKA Finder

A Windows application for StarCraft: Remastered players to identify opponents' alternate accounts (AKAs) and their highest MMR (Matchmaking Rating).

## Features

- Monitors StarCraft: Remastered replay files in real-time
- Identifies opponent's alternate accounts
- Displays highest MMR and corresponding rank for each AKA
- Custom futuristic UI with dark theme
- Standalone executable with no external dependencies

## Building from Source

### Prerequisites

- Go 1.19 or later
- Windows OS (required for StarCraft: Remastered integration)

### Build Process

1. Open PowerShell in the project directory
2. Run the build script:
   ```powershell
   .\build.ps1
   ```

This will create a standalone executable in the `dist` directory that contains all dependencies.

### Manual Build

If you prefer to build manually:

```bash
# Set environment variables for Windows build
$env:GOOS="windows"
$env:GOARCH="amd64"

# Build with embedded resources
go build -ldflags "-s -w" -tags=windows -o bwakafinder.exe .
```

## Usage

1. Launch StarCraft: Remastered
2. Run `bwakafinder.exe`
3. Play a match - the application will automatically detect the replay
4. View opponent's AKAs, MMR, and ranks in the application window

## How It Works

1. The application monitors the StarCraft: Remastered replay file (`LastReplay.rep`) for changes
2. When a new replay is detected, it parses the replay to identify the winner and loser
3. It determines which player is the local user based on account history
4. It queries the StarCraft: Remastered web API to find the opponent's alternate accounts
5. It displays the highest MMR and corresponding rank for each AKA found

## Dependencies

All dependencies are embedded in the executable:

- fyne.io/fyne/v2 - UI toolkit
- github.com/icza/screp - Replay parsing
- github.com/innerspirit/getscprocess - StarCraft process detection

## License

This project is licensed under the MIT License.
