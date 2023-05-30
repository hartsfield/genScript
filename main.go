package main

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	entries, err := os.ReadDir("./")
	if err != nil {
		log.Fatal(err)
	}

	var evs []string
	for _, e := range entries {
		if e.Name()[len(e.Name())-3:] == ".go" {
			lines := readFile("./" + e.Name())
			evs = findEnvVars(lines)
		}
	}
	spl := strings.Split(findGitOriginURL(), "/")
	appName := strings.Split(spl[len(spl)-1], ".")[0]
	script := genBashScript(evs, appName)
	log.Println("\n", script)

}

func findGitOriginURL() (originURL string) {
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
