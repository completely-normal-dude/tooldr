package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"golang.org/x/exp/slices"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

var (
	dir       = dataDir() + "/TLDR"
	en_pages  = dir + "/pages/common"
	en_pagesL = dir + "/pages/linux"
	separator = string(os.PathSeparator)
)

const (
	giturl = "https://github.com/tldr-pages/tldr.git"
)

func main() {
	nameProgram := os.Args[0]
	var update bool
	var length int = len(os.Args)
	if length < 2 {
		fmt.Fprintf(os.Stderr, "%s: Provide a package name to fetch tldr help\n", nameProgram)
		os.Exit(1)
	}
	if slices.Contains(os.Args, "-h") {
		helpMessage(nameProgram, 0)
	}
	if slices.Contains(os.Args, "-u") {
		update = true
		// Check if the repo is already cloned
		_, err := os.ReadDir(en_pages)
		if err != nil {
			fmt.Println("Cloning...")
			_, err = git.PlainClone(dir, false, &git.CloneOptions{
				URL:               giturl,
				ShallowSubmodules: true,
				SingleBranch:      true,
				Depth:             1,
				Progress:          os.Stdout,
			})
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("Updating...")
			os.RemoveAll(dir)
			_, err = git.PlainClone(dir, false, &git.CloneOptions{
				URL:               giturl,
				ShallowSubmodules: true,
				SingleBranch:      true,
				Depth:             1,
				Progress:          os.Stdout,
			})
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	if length == 2 && update {
		os.Exit(0)
	}
	if c := slices.Index(os.Args, "-f"); c != -1 {
		if length > c+1 {
			program := os.Args[c+1]
			var url string = "https://raw.githubusercontent.com/tldr-pages/tldr/main/pages/common/" + program + ".md"
			get, err := http.Get(url)
			body := get.Body
			content, _ := io.ReadAll(body)
			if err != nil {
				fmt.Println("Error while downloading tldr page:", err)
				os.Exit(1)
			} else if string(content) == "404: Not Found" {
				url = "https://raw.githubusercontent.com/tldr-pages/tldr/main/pages/linux/" + program + ".md"
				fmt.Println("URL:", url)
				get1, error := http.Get(url)
				if error != nil {
					fmt.Println(error)
				}
				io.Copy(os.Stdout, get1.Body)
				get1.Body.Close()
				os.Exit(0)
			}
			fmt.Println("URL:", url)
			fmt.Println(string(content))
			get.Body.Close()
			os.Exit(0)
		} else {
			fmt.Fprintln(os.Stderr, "Missing argument: '-f'")
			helpMessage(nameProgram, 1)
		}

	}
	var program int
	if update {
		for i := 1; i < 3; i++ {
			a := strings.Split(os.Args[i], "")
			if a[0] != "-" {
				program = i
				break
			}

		}
	} else {
		program = 1
	}
	var run bool
	pathToPage := en_pages + separator + os.Args[program] + ".md"
retry2:
	_, err := os.Stat(pathToPage)
	if err != nil {
		if !run {
			run = true
			pathToPage = en_pagesL + separator + os.Args[program] + ".md"
			goto retry2
		}
		fmt.Println("There's no tldr page for that program")
		os.Exit(1)
	}
	content, _ := os.ReadFile(pathToPage)
	fmt.Printf("%s", string(content))
}

func dataDir() (dir string) {
	switch runtime.GOOS {
	case "windows":
		dir = os.Getenv("APPDATA") + "/.local/share"
	case "linux", "solaris", "darwin":
		dir = os.Getenv("HOME") + "/.local/share"
	}
	return
}
func helpMessage(nameProgram string, exitCode int) {
	fmt.Printf(`USAGE: %s  <NameOfProgram> --options
		         
FLAGS:
    -f <NameOfProgram>  fetch tldr page directly from the repo (requires Internet connection)
    -u                  clone tldr repository in "%s"
    -h                  show this help message
`, nameProgram, dir)
	if exitCode == -2 {
		return
	}
	os.Exit(exitCode)
}
