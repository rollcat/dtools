package dmenu

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var dmenuPath = ""

func init() {
	var err error
	dmenuPath, err = exec.LookPath("dmenu")
	assert(err)
}

type DmenuOpts struct {
	ChoiceList      []string
	ChoiceChan      <-chan string
	ShowLines       int
	CaseInsensitive bool
	Bottom          bool
	Prompt          string
	NoEarlyGrab     bool
}

func (d *DmenuOpts) Choose() string {
	var err error
	var stdin io.Reader
	if d.ShowLines < 0 {
		panic("ShowLines must be 0 or greater")
	}
	args := []string{
		dmenuPath,
		"-l", fmt.Sprintf("%d", d.ShowLines),
	}
	if !d.NoEarlyGrab {
		args = append(args, "-f")
	}
	if d.CaseInsensitive {
		args = append(args, "-i")
	}
	if d.Bottom {
		args = append(args, "-b")
	}
	if d.Prompt != "" {
		args = append(args, "-p", d.Prompt)
	}
	if len(d.ChoiceList) > 0 || d.ChoiceChan != nil {
		var wp io.WriteCloser
		stdin, wp, err = os.Pipe()
		assert(err)
		go func() {
			defer wp.Close()
			for _, s := range d.ChoiceList {
				wp.Write([]byte(s))
				wp.Write([]byte("\n"))
			}
			if d.ChoiceChan == nil {
				return
			}
			for s := range d.ChoiceChan {
				wp.Write([]byte(s))
				wp.Write([]byte("\n"))
			}
		}()
	} else {
		stdin = strings.NewReader("")
	}
	stdout := &strings.Builder{}
	err = (&exec.Cmd{
		Path:   dmenuPath,
		Args:   args,
		Stdin:  stdin,
		Stdout: stdout,
	}).Run()
	if err != nil {
		return ""
	}
	return stdout.String()
}

func Dmenu(choices []string) string {
	return (&DmenuOpts{
		ChoiceList:      choices,
		ShowLines:       10,
		CaseInsensitive: true,
	}).Choose()
}

func DmenuChan(ch <-chan string) string {
	return (&DmenuOpts{
		ChoiceChan:      ch,
		ShowLines:       10,
		CaseInsensitive: true,
	}).Choose()
}

func Prompt(prompt string) string {
	return (&DmenuOpts{
		Prompt: prompt,
	}).Choose()
}
