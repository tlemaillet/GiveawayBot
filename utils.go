package main

import "github.com/cosiner/argv"

func isInArray(value string, sArray []string) bool {
	for _, s := range sArray {
		if s == value {
			return true
		}
	}
	return false
}

func logError(value string, sArray []string) bool {
	for _, s := range sArray {
		if s == value {
			return true
		}
	}
	return false
}

func parseArguments(argsLine string) ([]string, error) {
	argss, err := argv.Argv([]rune(argsLine), nil, nil)
	if err != nil {
		return nil, err
	}
	args := argss[0]

	return args, nil
}
