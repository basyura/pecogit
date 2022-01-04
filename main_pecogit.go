package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"pecogit/config"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var regex_half = regexp.MustCompile(`^[0-9a-zA-Z_\.\-/>\[\] ]+$`)

func main() {
	s, err := doMain()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(s)
}

func doMain() (string, error) {

	conf, err := config.Initialize(os.Args[1:])
	if err != nil {
		return "", err
	}

	var candidates []string

	cmd := conf.Command
	if cmd == "branch" {
		candidates, err = execGitBranch(conf)
		if err != nil {
			return "", err
		}
	} else {
		candidates, err = execGitCommand(conf)
		if err != nil {
			return "", err
		}
	}

	sort.Strings(candidates)
	s := toString(candidates)

	return s, nil
}

func execGitCommand(conf *config.Config) ([]string, error) {

	bytes, err := exec.Command("git", conf.Args...).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to exec %s %w", strings.Join(conf.Args, " "), err)
	}

	candidates := []string{}
	for _, s := range strings.Split(string(bytes), "\n") {
		candidates = append(candidates, s)
	}

	return candidates, nil
}

func execGitBranch(conf *config.Config) (candidates []string, err error) {

	args := []string{}
	max := -1
	i := -1
	for {
		i++
		if i >= len(conf.Args) {
			break
		}

		arg := conf.Args[i]
		if arg == "-n" {
			if len(conf.Args) <= i+1 {
				return nil, errors.New("-n requires number")
			}
			if max, err = strconv.Atoi(conf.Args[i+1]); err != nil {
				return nil, fmt.Errorf("-n requires number, not %s.\n%w", conf.Args[i+1], err)
			}
			i++
		} else {
			args = append(args, arg)
		}
	}

	cmd := exec.Command("git", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer stdout.Close()

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		text := scanner.Text()
		if candidate, ok := newCandidate(conf, text); ok {
			candidates = append(candidates, candidate)
		}
		if max > 0 && len(candidates) >= max {
			return candidates, nil
		}
	}

	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	return candidates, nil
}

func newCandidate(conf *config.Config, s string) (string, bool) {
	// check setting's ignores
	if !conf.IsValid(s) {
		return "", false
	}
	// check double bytes
	v := trim(s)
	if !regex_half.Match([]byte(v)) {
		return "", false
	}

	return v, true
}

func trim(s string) string {
	s = strings.Replace(s, "remotes/", "", 1)
	s = strings.TrimLeft(s, " ")
	s = strings.TrimLeft(s, "* ")
	return s
}

func toString(candidates []string) string {
	s := strings.Join(candidates, "\n")
	s = strings.TrimLeft(s, "\n")
	return s
}
