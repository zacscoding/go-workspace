package main

import (
	"fmt"
	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"io/ioutil"
	"log"
)

func listFiles(path string) func(string) []string {
	return func(line string) []string {
		names := make([]string, 0)
		files, _ := ioutil.ReadDir(path)
		for _, f := range files {
			names = append(names, f.Name())
		}
		return names
	}
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("mode",
		readline.PcItem("vi"),
		readline.PcItem("emacs"),
	),
	readline.PcItem("login"),
	readline.PcItem("say",
		readline.PcItemDynamic(listFiles("./"),
			readline.PcItem("with",
				readline.PcItem("following"),
				readline.PcItem("items"),
			),
		),
		readline.PcItem("hello"),
		readline.PcItem("bye"),
	),
	readline.PcItem("setprompt"),
	readline.PcItem("setpassword"),
	readline.PcItem("bye"),
	readline.PcItem("help"),
	readline.PcItem("go",
		readline.PcItem("build", readline.PcItem("-o"), readline.PcItem("-v")),
		readline.PcItem("install",
			readline.PcItem("-v"),
			readline.PcItem("-vv"),
			readline.PcItem("-vvv"),
		),
		readline.PcItem("test"),
	),
	readline.PcItem("sleep"),
	readline.PcItem("rpc",
		readline.PcItem("method1"),
		readline.PcItem("method2"),
	),
)

func main() {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:                 "\033[31m»\033[0m ",
		InterruptPrompt:        "^C",
		DisableAutoSaveHistory: true,
		AutoComplete:           completer,
		//HistoryFile:            ".history",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer rl.Close()

	printf := color.New(color.FgGreen).SprintfFunc()
	defer func() {
		println("Good Bye 🖐️🖐️")
	}()

	for i := 0; i < 2; i++ {
		line, err := rl.Readline()
		if err != nil {
			if err != readline.ErrInterrupt {
				fmt.Fprintf(rl.Stderr(), "failed to read line. err: %v", err)
			}
			return
		}
		if line == "exit" {
			return
		}
		fmt.Fprintf(rl.Stdout(), printf("%d)Read: %s. len(%d)\n", i+1, line, len(line)))
	}
}
