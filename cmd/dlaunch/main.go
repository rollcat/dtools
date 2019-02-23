// https://specifications.freedesktop.org/desktop-entry-spec/latest/
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/rollcat/dtools/dmenu"

	"gopkg.in/ini.v1"
)

type desktopFile struct {
	fpath       string
	Comment     string
	Exec        string // Total mess. Do not pass to shell.
	GenericName string
	Hidden      string // "This should have been called 'deleted'"...
	Icon        string
	Name        string
	NoDisplay   string // "It exists, but do not show it"
	Path        string
	Terminal    string
	TryExec     string // Show if this is in $PATH
	Type        string // One of: Application(1), Link(2), Directory(3)
	Keywords    string // Used for indexing in search
}

var termPath = ""
var desktopPaths = []string{}
var desktopFiles = map[string]*desktopFile{}

func assert(ex ...error) {
	for _, e := range ex {
		if e != nil {
			panic(e)
		}
	}
}

func findDesktopFiles(paths []string) <-chan *desktopFile {
	out := make(chan *desktopFile, 1)
	wg := sync.WaitGroup{}
	for _, path := range paths {
		wg.Add(1)
		go func(path string, out chan<- *desktopFile) {
			f := func(fpath string, info os.FileInfo, err error) error {
				if filepath.Ext(fpath) == ".desktop" {
					df, err := loadDesktopFile(fpath)
					assert(err)
					if df != nil {
						out <- df
					}
				}
				return nil
			}
			err := filepath.Walk(path, f)
			assert(err)
			wg.Done()
		}(path, out)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func loadDesktopFile(fpath string) (*desktopFile, error) {
	cfg, err := ini.Load(fpath)
	if err != nil {
		return nil, err
	}
	entry, err := cfg.GetSection("Desktop Entry")
	if err != nil {
		return nil, err
	}
	df := desktopFile{
		fpath: fpath,
	}
	err = entry.MapTo(&df)
	if err != nil {
		return nil, err
	}
	if df.TryExec != "" {
		df.TryExec, err = exec.LookPath(df.TryExec)
		if err != nil {
			log.Print(err)
			return nil, nil
		}
	}
	if df.NoDisplay == "true" || df.Hidden == "true" {
		return nil, nil
	}
	return &df, nil
}

func (df *desktopFile) launch() error {
	// Please go read up the spec on Exec and tell me it makes sense?
	// Especially the input mangling. (If your program calls a shell
	// on untrusted input, you have way worse problems.)
	origExecArgs := strings.Split(df.Exec, " ")
	var execArgs []string
	for _, arg := range origExecArgs {
		if len(arg) == 0 {
			// WTF? remove
			continue
		}
		switch arg[0:1] {
		case "%":
			if len(arg) == 1 {
				// Broken, remove
				continue
			}
			switch arg[0:2] {
			case "%%":
				execArgs = append(execArgs, fmt.Sprintf("%%%s", arg[2:]))
			case "%c":
				execArgs = append(execArgs, df.Name)
			case "%k":
				execArgs = append(execArgs, df.fpath)
			case "%i":
				if df.Icon != "" {
					execArgs = append(execArgs, "--icon", df.Icon)
				}
			case "%f":
			case "%F":
			case "%u":
			case "%U":
				// Cases of %f, %F, %u and %U can be ignored - we do
				// not open files, only launch applications.
			default:
				// do not use this argument.
			}
		default:
			execArgs = append(execArgs, arg)
		}
	}
	if df.Terminal == "true" {
		execArgs = append([]string{"x-terminal-emulator", "-e"}, execArgs...)
	}
	cmd := exec.Command(execArgs[0], execArgs[1:]...)
	cmd.Dir = df.Path
	log.Print("command:", cmd)
	err := cmd.Start()
	go func() { cmd.Wait() }()
	return err
}

func main() {
	var err error
	termPath, err = exec.LookPath("urxvt") // TODO: getenv("XTERM") ?
	assert(err)

	// TODO: $XDG_WHATEVER
	desktopPaths = append(desktopPaths,
		"/usr/share/applications",
		"/usr/local/share/applications",
		path.Join(os.Getenv("HOME"), ".local/share/applications"),
	)

	options := []string{}
	for df := range findDesktopFiles(desktopPaths) {
		key := df.Name
		if df.GenericName != "" && df.Comment != "" {
			key = fmt.Sprintf(
				"%s (%s; %s)",
				df.Name,
				df.GenericName,
				df.Comment,
			)
		} else if df.Comment != "" {
			key = fmt.Sprintf(
				"%s (%s)",
				df.Name,
				df.Comment,
			)
		} else if df.GenericName != "" {
			key = fmt.Sprintf(
				"%s (%s)",
				df.Name,
				df.GenericName,
			)
		}
		desktopFiles[key] = df
		options = append(options, key)
	}
	sort.Strings(options)
	choice := dmenu.Dmenu(options)
	choice = strings.TrimRight(choice, "\n")
	df, ok := desktopFiles[choice]
	if ok {
		err = df.launch()
	} else {
		program, err := exec.LookPath(choice)
		if err == nil {
			cmd := exec.Command(program)
			err = cmd.Start()
			go func() { cmd.Wait() }()
		}
	}
	if err != nil {
		log.Print(err)
	}
}
