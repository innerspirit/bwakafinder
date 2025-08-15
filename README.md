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

Run the PowerShell build script:
```powershell
.\build.ps1
```

This creates a standalone executable in the `dist` directory with all dependencies included.

For manual builds:
```bash
go build -tags=windows -o bwakafinder.exe .
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
## GitHub Release Process

This project includes a GitHub Actions workflow that automatically builds and creates releases when you push a tag.

### Creating a Release

1. Create a new tag:
   ```bash
   git tag -a v1.0.0 -m "Release version 1.0.0"
   git push origin v1.0.0
   ```

2. The workflow will:
   - Build the Windows executable
   - Create a draft release with the executable
   - Generate release notes

3. Edit and publish the draft release on GitHub.

## Dependencies

All dependencies are embedded in the executable:

- fyne.io/fyne/v2 - UI toolkit
- github.com/icza/screp - Replay parsing
- github.com/innerspirit/getscprocess - StarCraft process detection

## License

This project is licensed under the MIT License.
