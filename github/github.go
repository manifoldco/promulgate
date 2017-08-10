package github

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/manifoldco/promulgate/artifact"
	"github.com/manifoldco/promulgate/git"
)

// Client is a github client
type Client struct {
	c *github.Client
	r *git.Repository
}

// New creates a new github client
func New(repo *git.Repository) (*Client, error) {
	tok, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		return nil, errors.New("Please set GITHUB_TOKEN")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: tok})
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	return &Client{
		c: github.NewClient(tc),
		r: repo,
	}, nil
}

// Tags returns a list of semver sorted tags that exist on the repo
func (c *Client) Tags(ctx context.Context) (Tags, error) {
	var tags []*github.RepositoryTag
	opt := &github.ListOptions{PerPage: 100}
	for {
		tagPage, resp, err := c.c.Repositories.ListTags(ctx, c.r.Owner, c.r.Name, opt)
		if err != nil {
			return nil, errors.New("Could not read tags: " + err.Error())
		}

		tags = append(tags, tagPage...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	sort.Sort(tagSorter(tags))
	return tags, nil
}

// Releases returns a list of semver sorted releases that exist on the repo
func (c *Client) Releases(ctx context.Context) ([]*github.RepositoryRelease, error) {
	var releases []*github.RepositoryRelease
	opt := &github.ListOptions{PerPage: 100}
	for {
		releasePage, resp, err := c.c.Repositories.ListReleases(ctx, c.r.Owner, c.r.Name, opt)
		if err != nil {
			return nil, errors.New("Could not read releases: " + err.Error())
		}

		releases = append(releases, releasePage...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	sort.Sort(releaseSorter(releases))
	return releases, nil
}

// CreateRelease creates a new release on github
func (c *Client) CreateRelease(ctx context.Context, rel *artifact.Release) error {
	_, _, err := c.c.Repositories.CreateRelease(ctx, c.r.Owner, c.r.Name, &github.RepositoryRelease{
		Name:    github.String(rel.Tag),
		TagName: github.String(rel.Tag),
		Body:    github.String(rel.Body),
	})

	return err
}

// AddArtifact adds a file on to a release
func (c *Client) AddArtifact(ctx context.Context, rel *artifact.Release, file *artifact.File) error {
	ghr, _, err := c.c.Repositories.GetReleaseByTag(ctx, c.r.Owner, c.r.Name, rel.Tag)
	if err != nil {
		return err
	}

	u := fmt.Sprintf("repos/%s/%s/releases/%d/assets", c.r.Owner, c.r.Name, *ghr.ID)

	req, err := c.c.NewUploadRequest(u, file.Data, file.Size, file.Type)
	if err != nil {
		return err
	}

	_, err = c.c.Do(ctx, req, nil)
	return err
}

func semverCmp(v1s, v2s string) int {
	v1, err1 := semver.ParseTolerant(v1s)
	v2, err2 := semver.ParseTolerant(v2s)

	// unparseable semvers are 'less' than regulars. Two unparseable semvers
	// are compared as regular strings.
	switch {
	case err1 == nil && err2 == nil:
		return v1.Compare(v2)
	case err1 != nil && err2 != nil:
		return strings.Compare(v1s, v2s)
	case err1 != nil:
		return -1
	default:
		return 1
	}
}

// Sorters for tags and github releases. Elements are sorted in ascending order,
// based on semver. tags that can't be parsed as semvers are ordered before
// parseable tags.

type tagSorter []*github.RepositoryTag

func (t tagSorter) Len() int           { return len(t) }
func (t tagSorter) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t tagSorter) Less(i, j int) bool { return semverCmp(*t[i].Name, *t[j].Name) < 0 }

type releaseSorter []*github.RepositoryRelease

func (r releaseSorter) Len() int           { return len(r) }
func (r releaseSorter) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r releaseSorter) Less(i, j int) bool { return semverCmp(*r[i].Name, *r[j].Name) < 0 }

// Tags is the list of tags on the repo, used for computing the difference with
// releases
type Tags []*github.RepositoryTag

// Difference computes the difference between the repository tags, and releases.
// It returns the tags that do not have matching releases.
func (tags Tags) Difference(releases []*github.RepositoryRelease) Tags {
	var tagsWithNoRelease []*github.RepositoryTag

	i := 0
outerLoop:
	for _, release := range releases {
		for i < len(tags) {
			tag := tags[i]

			switch semverCmp(*release.Name, *tag.Name) {
			case 1: // no release for this tag
				tagsWithNoRelease = append(tagsWithNoRelease, tag)
				i++
			case 0: // Found it
				i++
				continue outerLoop
			default: // no tag for this release (we ignore this case)
				continue outerLoop
			}
		}

	}

	tagsWithNoRelease = append(tagsWithNoRelease, tags[i:]...)

	return tagsWithNoRelease
}

// ValidSemver filters out any tags that don't have valid semvers
func (tags Tags) ValidSemver() Tags {
	var valid []*github.RepositoryTag
	for _, tag := range tags {
		_, err := semver.ParseTolerant(*tag.Name)
		if err != nil { // If its not a valid semver, its not a tag we release
			continue
		}

		valid = append(valid, tag)
	}

	return valid
}

// NoPrerelease filters out any tags that are prereleases
func (tags Tags) NoPrerelease() Tags {
	var released []*github.RepositoryTag
	for _, tag := range tags {
		ver, err := semver.ParseTolerant(*tag.Name)
		if err != nil { // If its not a valid semver, its not a tag we release
			continue
		}

		prerelease := len(ver.Pre) > 0
		if prerelease { // We don't create real releases for release candidates.
			continue
		}

		released = append(released, tag)
	}

	return released
}
