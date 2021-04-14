package findhidraw

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/manifoldco/promptui"
)

func hidraws(path string) ([]string, error) {
	hs := []string{}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return []string{}, fmt.Errorf("could not dir files: %w", err)
	}

	for _, f := range files {
		if strings.Contains(f.Name(), "hidraw") {
			hs = append(hs, f.Name())
		}
	}

	return hs, nil
}

func FindHidraw(path string) (string, error) {
	prompt := promptui.Prompt{
		Label:     "please unplug reader",
		IsConfirm: true,
	}

	ok, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("could not read input: %w", err)
	}

	if ok != "y" {
		return "", fmt.Errorf("did not unplug the reader")
	}

	return "", nil
}
