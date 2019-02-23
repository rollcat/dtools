package main

import (
	"fmt"
	"os/exec"

	"github.com/rollcat/dtools/dmenu"
)

func Sh(shell string) {
	(&exec.Cmd{
		Path: "/bin/sh",
		Args: []string{"/bin/sh", "-c", shell},
	}).Run()
}

func main() {
	keymapChoices := []string{"pl", "el", "de", "ru", "hr"}
	choice := (&dmenu.DmenuOpts{
		ChoiceList: keymapChoices,
		Prompt:     "setxkbmap",
	}).Choose()
	Sh(fmt.Sprintf("setxkbmap %s", choice))
	Sh("xmodmap ~/.xmodmap")
	Sh("setxkbmap -option ctrl:nocaps")
}
