package changelog

import (
	"io/ioutil"
	"regexp"
)

var sectionHeader = regexp.MustCompile(`(?m:^##\s+(v[0-9]+\.[0-9]+\.[0-9]+)\s*$)`)

// Section holds CHANGELOG.md sections, mapping the release tag to the body
// contents.
type Section struct {
	Tag  string
	Body string
}

// Changelog is a representation of a CHANGELOG file, split by release sections.
type Changelog map[string]Section

// Read reads an existing CHANGELOG file.
func Read(path string) (Changelog, error) {
	cl, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	changelog := make(Changelog)
	parts := sectionHeader.FindAllSubmatchIndex(cl, -1)
	for i, part := range parts {
		var end int
		if i+1 == len(parts) {
			end = len(cl)
		} else {
			end = parts[i+1][0]
		}

		sec := Section{
			Tag:  string(cl[part[2]:part[3]]),
			Body: string(cl[part[1]:end]),
		}
		changelog[sec.Tag] = sec
	}

	return changelog, nil
}
