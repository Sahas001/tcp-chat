package main

import "regexp"

var mentionRegex = regexp.MustCompile(`@(\w+)`)

func ExtractMentions(msg string) []string {
	matches := mentionRegex.FindAllStringSubmatch(msg, -1)
	var tags []string

	for _, match := range matches {
		tags = append(tags, match[1])
	}
	return tags
}
