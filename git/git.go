package git

import (
	"errors"
	"strings"

	"gopkg.in/src-d/go-git.v4"
)

// Repository is a local git configuration
type Repository struct {
	Owner string
	Name  string
}

// Open returns a new Repository as found at the given path.
func Open(path string) (*Repository, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return nil, errors.New("Couldn't open git repository: " + err.Error())
	}

	c, err := r.Config()
	if err != nil {
		return nil, errors.New("Couldn't read git config: " + err.Error())
	}

	for k, v := range c.Remotes {
		if k == "origin" {
			url := v.URL
			parts := strings.SplitN(url, ":", -1)
			parts = strings.SplitN(parts[len(parts)-1], "/", -1)
			owner := parts[len(parts)-2]
			name := parts[len(parts)-1]
			name = name[:len(name)-4]
			return &Repository{
				Owner: owner,
				Name:  name,
			}, nil
		}
	}

	return nil, errors.New("No git origin configuration found")
}
