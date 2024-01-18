package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"log"
)

func main() {
	p := tea.NewProgram(
		InitialModel(),
		tea.WithAltScreen()) // use the full size of the terminal in its "alternate screen buffer"
	go func() {
		run()
	}()

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
