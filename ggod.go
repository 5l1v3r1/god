package main

import (
	"io"
    "fmt"
	"os"
	"os/exec"
	"bytes"
	"strings"
	"github.com/carmark/pseudo-terminal-go/terminal"
	"github.com/abiosoft/ishell"
)

var outb, errb bytes.Buffer
var cmdSlice, cmdList []string

func execCmd(input string, stdout bool) string {
    // Remove the newline character.
    input = strings.TrimSuffix(input, "\n")
	
	// Prepare the command to execute.	
	cmd := exec.Command("sh", "-c", input)

	// Set the correct output device.
	if stdout{
    	cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}else{
		cmd.Stdout = &outb
		cmd.Stderr = &errb
	}
    // Execute the command and return the error.
	err := cmd.Run()
	if err != nil{
		return err.Error()
	}
	if !stdout{
		return outb.String()
	}
	return ""
}

func searchInSlice(slice []string, query string) bool{
	set := make(map[string]bool)
	for _, v := range slice {
		set[v] = true
	}
	return set[query]
}

func prepareCmds(){
	removeSpaces := "sed -e 's/^\\s*//' -e 's/ *[A-Z].*//' && "
	parseGitCmd := 
		"git help | grep '^  *[a-z]' | " + removeSpaces +
		"git branch | tr -d '*' | " + removeSpaces +
		"git remote"
	cmdStr := execCmd(parseGitCmd, false)
	cmdList = strings.Split(cmdStr, "\n")
	for _, cmd := range cmdList {
		if (len(cmd) > 0){
			firstChar := string([]rune(cmd)[0])
			if (!searchInSlice(cmdSlice, firstChar)){
				cmdSlice = append(cmdSlice, firstChar)
			}else{
				cmdSlice = append(cmdSlice, firstChar + 
					string([]rune(cmd)[len(cmd)/2])) 
			}
		}
	}
}

func buildCmd(line string) string {
	line = strings.Replace(line, " git ", " ", -1)
	for index, cmd := range cmdSlice {
		cmd = " " + cmd + " "
		if (strings.Contains(line, cmd)) {
			line = strings.Replace(line, cmd, 
				" " + cmdList[index] + " ", -1)
		}
	}
	return "git" + line
}

func startTerm() {
	term, err := terminal.NewWithStdInOut()
	if err != nil {
		panic(err)
	}
	defer term.ReleaseFromStdInOut()
	term.SetPrompt("> ")
	cmdLoop:
	for {
		// Read the keyboad input.
		line, err := term.ReadLine()
		// Exit on Ctrl-D and Ctrl-C
		if err == io.EOF ||  line == "^C" {
			fmt.Println()
			return
		}
		// Built-in commands
		switch line{
		case "", " ": 
			break
		case "clear":
			execCmd("clear", true)
		case "exit":
			break cmdLoop
		case "?", "help":
			printUsage()
		default:
			// Handle the execution of the input.
			gitCmd := buildCmd(" " + line + " ")
			if retval := execCmd(gitCmd, true); len(retval) > 0 {
				fmt.Fprintln(os.Stderr, retval)
			}
		}
	}
}
func printUsage(){
	for index, cmd := range cmdSlice {
		fmt.Println(cmdList[index] + " -> " + cmd)
	}
}
func main() {
	//prepareCmds()
	//printUsage()
	//startTerm()

	shell := ishell.New()
	shell.Println("GGod")
	shell.SetPrompt("> ")

    shell.AddCmd(&ishell.Cmd{
        Name: "clear",
        Help: "clear the terminal",
        Func: func(c *ishell.Context) {
            execCmd("clear", true)
        },
	})

	shell.AddCmd(&ishell.Cmd{
        Name: "?",
        Help: "display help",
        Func: func(c *ishell.Context) {
            c.Println(c.HelpText())
        },
	})
	
    shell.Run()
}