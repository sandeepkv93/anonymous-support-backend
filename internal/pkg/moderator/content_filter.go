package moderator

import (
	"strings"
)

var profanityList = []string{
	"fuck", "shit", "damn", "ass", "bitch", "bastard",
}

var harmfulKeywords = []string{
	"suicide", "kill yourself", "end it all", "self-harm",
}

type ContentFilter struct {
	level string
}

func NewContentFilter(level string) *ContentFilter {
	return &ContentFilter{level: level}
}

func (cf *ContentFilter) ContainsProfanity(text string) bool {
	lowerText := strings.ToLower(text)
	for _, word := range profanityList {
		if strings.Contains(lowerText, word) {
			return true
		}
	}
	return false
}

func (cf *ContentFilter) ContainsHarmfulContent(text string) bool {
	lowerText := strings.ToLower(text)
	for _, keyword := range harmfulKeywords {
		if strings.Contains(lowerText, keyword) {
			return true
		}
	}
	return false
}

func (cf *ContentFilter) CheckContent(text string) []string {
	flags := []string{}

	if cf.ContainsProfanity(text) {
		flags = append(flags, "profanity")
	}

	if cf.ContainsHarmfulContent(text) {
		flags = append(flags, "harmful_content")
	}

	return flags
}

func (cf *ContentFilter) ShouldAutoFlag(text string) bool {
	return len(cf.CheckContent(text)) > 0
}
