package cli

import (
	"errors"
	"strings"
)

func Parse(cmd string) (Command, error) {
	tokens := strings.Fields(cmd)
	if len(tokens) < 2 {
		return Command{}, errors.New("Invalid input")
	}
	return Command{ Name: strings.ToUpper(tokens[0]), Args: tokens[1:]}, nil
}