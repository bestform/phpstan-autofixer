package main

import (
	"regexp"
	"io/ioutil"
	"strings"
	"os"
)

// StanFixer can decide if it can do something about the error and if so is responsible to do so.
// This includes reading and writing the file.
// It returns true if it actually did something, false otherwise.
type StanFixer func(stanError StanError) (bool, error)

// phpDocParamMissingType fixes missing types in php doc param declarations
func phpDocParamMissingType(stanError StanError) (bool, error) {
	matched, err := regexp.Match("^PHPDoc tag @param has invalid value.*", []byte(stanError.Message))
	if err != nil {
		return false, err
	}
	if !matched {
		return false, nil
	}

	fileMode, err := readFileMode(stanError.File)
	if err != nil {
		return false, err
	}

	content, err := ioutil.ReadFile(stanError.File)
	if err != nil {
		return false, err
	}

	lines := strings.Split(string(content), "\n")

	for i := range lines {
		if i+1 != stanError.Line {
			continue
		}

		lines = fixDocBlockBefore(lines, i)
	}

	ioutil.WriteFile(stanError.File, []byte(strings.Join(lines, "\n")), fileMode)

	return true, nil
}

func readFileMode(fileName string) (os.FileMode, error) {
	var mode os.FileMode
	file, err := os.Open(fileName)
	if err != nil {
		return mode, err
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return mode, err
	}

	return fileStat.Mode(), nil
}

func fixDocBlockBefore(lines []string, i int) []string {
	re := regexp.MustCompile("@param\\s*(\\$.*)")

	var line string
	for line = lines[i]; !strings.Contains(line, "/**") && i != 0; i-- {
		fixedLine := re.ReplaceAllString(line, "@param mixed $1 ")
		lines[i] = fixedLine
		line = lines[i-1]
	}

	return lines
}

