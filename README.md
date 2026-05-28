# 🧹 broomie

```text
 ______   _______  _______  _______  _______  _______ _________
(  ___ \ (  ____ )(  ___  )(  ___  )(       )\__   __/(  ____ \
| (   ) )| (    )|| (   ) || (   ) || () () |   ) (   | (    \/
| (__/ / | (____)|| |   | || |   | || || || |   | |   | (__
|  __ (  |     __)| |   | || |   | || |(_)| |   | |   |  __)
| (  \ \ | (\ (   | |   | || |   | || |   | |   | |   | (
| )___) )| ) \ \__| (___) || (___) || )   ( |___) (___| (____/\
|/ \___/ |/   \__/(_______)(_______)|/     \|\_______/(_______/
```

**broomie** is a sleek, fast, and feature-rich macOS system junk cleaner with a beautiful Terminal User Interface (TUI). Built in Go, it leverages Charm's TUI suite (`bubbletea`, `lipgloss`, `bubbles`) to provide a visual interactive interface for scanning, sorting, selecting, and reclaiming disk space on your Mac.

---

## ✨ Features

- **🚀 Highly Efficient Scanning**: Performs concurrent checks of common macOS system and application junk paths, using optimized `du` commands for rapid subdirectory size calculations.
- **📦 Comprehensive Junk Detection**:
  - **System & Application Cache** (e.g., `~/Library/Caches`, `~/.cache`)
  - **System & Application Logs** (e.g., `~/Library/Logs`)
  - **Temporary Files** (e.g., `/tmp`, `/private/var/tmp`)
  - **Deleted App Leftovers** (uninstalled application artifacts remaining in `~/Library/Application Support`, `~/Library/Containers`, etc.)
  - **Leftover System Update Data** (e.g., `/Library/Updates`, `/macOS Install Data`)
  - **iPhone Device Backups** (e.g., `~/Library/Application Support/MobileSync/Backup`)
  - **Xcode Caches & Derived Data**
  - **Xcode Simulator Device Environments**
- **📊 Dynamic Visual Interface**:
  - Beautiful, retro-futuristic loading screen tracking execution metrics.
  - Interactive multi-column table displaying path name, file size, relative modification age, and scan category.
  - Real-time calculations of selected folders and how much disk space will be reclaimed.
- **🔌 Scripting-Friendly Pipeline**: Can terminate and print selected paths to standard output (`stdout`), allowing seamless integration with custom deletion scripts or commands.

![](https://raw.github.com/hzqtc/taproom/master/screenshot.png)

---

## ⌨️ Controls & Keybindings

Once the scan is complete, you can interact with the table using the following controls:

| Key                      | Action                                                                  |
| :----------------------- | :---------------------------------------------------------------------- |
| `↑` / `↓` (or `k` / `j`) | Navigate up and down through the list of scanned entries                |
| `<Space>`                | Toggle selection on the highlighted item                                |
| `a`                      | Select all scanned items                                                |
| `A`                      | Unselect all scanned items                                              |
| `s` / `S`                | Cycle through sortable columns (`Reason`, `Path`, `Size`, `Date`)       |
| `r`                      | Reverse/Toggle sorting direction (Ascending / Descending)               |
| `q`                      | Quit the application without outputting                                 |
| `X`                      | Quit and output the selected file paths to `stdout` (useful for piping) |

---

## 🛠️ Installation & Building

### Prerequisites

- macOS operating system (specifically tailored for macOS paths and behaviors).
- [Go](https://go.dev/) 1.25.1 or later.

### Building from Source

1. Clone this repository to your local machine.
2. Build the binary using the provided `Makefile`:
   ```bash
   make build
   ```
3. Run the executable immediately:
   ```bash
   make run
   ```

### Global Installation

To install `broomie` into your local executable bin directory (defaults to `~/.local/bin`):

```bash
make install
```

To clean build artifacts:

```bash
make clean
```

---

## 🖥️ Command Line Usage

```bash
broomie [flags]
```

### Available Flags

- `-h`, `--help`: Show the help message and exit.
- `-v`, `--version`: Print the application version and exit.

### Automated Deletion Workflow Example

Since `broomie` prints selected items to `stdout` when exiting via `X`, you can combine it with `xargs` to automatically review and delete directories:

```bash
# Preview, select items in the TUI, hit 'X', and delete them safely:
broomie | xargs rm -rf
```

> [!WARNING]
> Always verify the directories and files selected inside `broomie` before piping to destructive commands like `rm -rf`.

---

## 📝 Diagnostic Logs

`broomie` records execution logs and error diagnostics silently in the background to avoid disrupting your fullscreen TUI workspace.
Logs can be viewed at:

```bash
tail -f /tmp/broomie.log
```
