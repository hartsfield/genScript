package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	linesOfGoCode := findGoCode()
	evs := findEnvVars(linesOfGoCode)
	spl := strings.Split(findGitOriginURL(), "/")
	appName := strings.Split(spl[len(spl)-1], ".")[0]
	fmt.Println("App name established: " + appName)
	script := genBashScript(evs, appName)
	fmt.Println(script)
	writeToFile(script)
	fmt.Println("Written to 'restart-service.sh'")
}

func writeToFile(script string) {
	f, err := os.OpenFile("access.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	if _, err = f.WriteString(script); err != nil {
		log.Panic(err)
	}
}

func findGoCode() (goFiles []string) {
	fmt.Println("Looking for files ending in '.go' (non-recursive)")
	entries, err := os.ReadDir("./")
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range entries {
		if e.Name()[len(e.Name())-3:] == ".go" {
			fmt.Println("Found go code: ", e.Name())
			goFiles = append(goFiles, readFile("./"+e.Name())...)
		}
	}
	return
}

func findGitOriginURL() (originURL string) {
	fmt.Println("Attempting to establish app name based on git origin URL...")
	lines := readFile("./.git/config")
	for _, line := range lines {
		if strings.Contains(line, "url = ") {
			if strings.Contains(line, "@") {
				originURL = "https://" + strings.Replace(strings.Split(line, "@")[1], ":", "/", 1)
			} else {
				originURL = strings.Split(line, "rl = ")[1]
			}
		}
	}
	return
}

func genBashScript(envVars []string, appName string) (script string) {
	fmt.Println("Generating bash script for " + appName + "...\n\n")
	script = "#!/bin/bash"

	for _, ev := range envVars {
		script = script + "\n" + ev + "="
	}

	script = script +
		"\n" +
		"trap -- '' SIGTERM\n" +
		"git pull\n" +
		"go build -o " + appName + "\n" +
		"pkill -f " + appName + "\n" +
		"nohup ./" + appName + " > /dev/null & disown\n" +
		"sleep 2"

	return
}

func filterNonAlpha(inTokens []string) (alpha []string) {
	alphas := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for _, token := range inTokens {
	testAlpha:
		for _, char := range alphas {
			if strings.ContainsRune(token, char) {
				alpha = append(alpha, token)
				break testAlpha
			}
		}
	}
	return
}

func findEnvVars(lines []string) (envVars []string) {
	fmt.Println("Looking for environment variable names set with os.Getenv()...")
	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if strings.Contains(line, "os.Getenv(") {
			envs := strings.SplitAfter(line, "os.Getenv(")
			envs = filterNonAlpha(envs)
			for _, en := range envs {
				if !(strings.Contains(en, "os.Getenv(")) {
					envVars = append(envVars, strings.Split(en[1:], `"`)[0])
				}
			}
		}
	}
	if len(envVars) <= 0 {
		fmt.Println("Couldn't find any environment variables")
		log.Fatal()
		return
	}
	fmt.Println("Found " + strconv.Itoa(len(envVars)) + " environment variables")
	return
}

func readFile(filename string) (lines []string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return
}
