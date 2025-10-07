package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"strings"

	"broomie/internal/model"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/pflag"
)

var (
	flagShowVersion = pflag.BoolP("version", "v", false, "Show app version")
	flagShowHelp    = pflag.BoolP("help", "h", false, "Show help message")
)

//go:embed .version
var version string

func main() {
	pflag.Parse()

	if *flagShowVersion {
		fmt.Print(version)
		os.Exit(0)
	}

	if *flagShowHelp {
		pflag.Usage()
		os.Exit(0)
	}

	logfile := "/tmp/broomie.log"
	f, err := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("failed to create log file: %v", err)
	}
	defer f.Close()
	// Send log output to the file
	log.SetOutput(f)

	// The WithAltScreen() option provides a full-screen TUI experience.
	p := tea.NewProgram(model.InitialModel(), tea.WithAltScreen())
	if m, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	} else if appModel, ok := m.(model.Model); ok {
		fmt.Printf("%s\n", strings.Join(appModel.Output, "\n"))
	}
}
