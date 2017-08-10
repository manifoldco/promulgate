package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/manifoldco/promulgate/artifact"
	"github.com/manifoldco/promulgate/changelog"
	"github.com/manifoldco/promulgate/git"
	"github.com/manifoldco/promulgate/github"
	"github.com/manifoldco/promulgate/s3"
)

var release = cli.Command{
	Name:      "release",
	Usage:     "Create and publish a new release from the given tag",
	ArgsUsage: "<tag>",
	Action:    releaseCmd,
}

func releaseCmd(cmd *cli.Context) error {
	if len(cmd.Args()) != 1 {
		return cli.NewExitError("tag is required", -1)
	}
	tag := cmd.Args()[0]

	ctx := context.Background()

	dir, err := os.Getwd()
	if err != nil {
		return cli.NewExitError("Could not get current dir: "+err.Error(), -1)
	}

	r, err := git.Open(dir)
	if err != nil {
		return cli.NewExitError("Could read git configuration: "+err.Error(), -1)
	}

	cl, err := changelog.Read("CHANGELOG.md")
	if err != nil {
		return cli.NewExitError("Could not parse changelog: "+err.Error(), -1)
	}

	c, err := github.New(r)
	if err != nil {
		return cli.NewExitError("Could not create github client: "+err.Error(), -1)
	}

	tags, err := c.Tags(ctx)
	if err != nil {
		return cli.NewExitError("Could not fetch github tags: "+err.Error(), -1)
	}

	var found bool
	for _, t := range tags {
		if *t.Name == tag {
			found = true
		}
	}
	if !found {
		return cli.NewExitError("No such tag: "+tag, -1)
	}

	releases, err := c.Releases(ctx)
	if err != nil {
		return cli.NewExitError("Could not fetch github releases: "+err.Error(), -1)
	}

	tagsWithNoRelease := tags.Difference(releases)

	found = false
	for _, t := range tagsWithNoRelease {
		if *t.Name == tag {
			found = true
		}
	}
	if !found {
		fmt.Println("tag already has github release")
	} else {

		body := "*No changelog entry*"
		if sec, ok := cl[tag]; ok {
			body = sec.Body
		}

		rel := &artifact.Release{
			Tag:  tag,
			Body: body,
		}

		fmt.Println("Creating release", rel.Tag)
		err = c.CreateRelease(ctx, rel)
		if err != nil {
			return cli.NewExitError("Could not create release: "+err.Error(), -1)
		}
	}

	zips, err := artifact.FindZips("build", r.Name, tag[1:])
	if err != nil {
		return cli.NewExitError("Could not find all build zips: "+err.Error(), -1)
	}

	s3c, err := s3.New("s3://releases.manifold.co")
	if err != nil {
		return cli.NewExitError("Could not create s3 client: "+err.Error(), -1)
	}

	for _, zip := range zips {
		err := s3c.Put(ctx, &zip)
		if err != nil {
			return cli.NewExitError("Could not upload file to s3: "+err.Error(), -1)

		}
	}

	fmt.Println("building cdn index pages")
	err = s3c.CreateIndexes()
	if err != nil {
		return cli.NewExitError("Could not build cdn index files: "+err.Error(), -1)
	}

	return nil
}

func init() {
	Cmds = append(Cmds, release)
}