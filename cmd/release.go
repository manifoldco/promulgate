package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/manifoldco/promulgate/artifact"
	"github.com/manifoldco/promulgate/brew"
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
	Flags: []cli.Flag{
		cli.BoolTFlag{
			Name:  "homebrew",
			Usage: "Create a new Homebrew version",
		},
	},
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

	zips, err := artifact.FindCompressedFiles("build", r.Name, tag[1:])
	if err != nil {
		return cli.NewExitError("Could not find all build zips: "+err.Error(), -1)
	}

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
		if sec, ok := cl[tag[1:]]; ok {
			body = sec.Body
		} else {
			return cli.NewExitError("No changelog entry found", -1)
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

		fmt.Println("Adding built zips")
		for _, zip := range zips {
			err := c.AddArtifact(ctx, rel, &zip)
			if err != nil {
				return cli.NewExitError("Couldn't upload zip: "+err.Error(), -1)
			}
		}
	}

	var darwin *artifact.File
	gzip := fmt.Sprintf("%s_%s_%s.tar.gz", r.Name, tag[1:], "darwin_amd64")
	zip := fmt.Sprintf("%s_%s_%s.zip", r.Name, tag[1:], "darwin_amd64")
	for _, f := range zips {
		if f.Name == gzip || f.Name == zip {
			darwin = &f
			break
		}
	}

	if darwin == nil {
		return cli.NewExitError("Could not find zip to convert to bottle", -1)
	}

	if cmd.Bool("homebrew") {
		bottles, binname, err := brew.NewBottles(darwin, r, tag[1:])
		if err != nil {
			return cli.NewExitError("Could not convert zip into bottle: "+err.Error(), -1)
		}

		gr, err := c.Info(ctx)
		if err != nil {
			return cli.NewExitError("Could not get repository info from github: "+err.Error(), -1)
		}

		formula, err := brew.NewFormula(r, tag, binname, gr.Homepage, gr.Description, "https://releases.manifold.co/", bottles)
		if err != nil {
			return cli.NewExitError("Could not create brew formula", -1)
		}

		bc, err := github.New(&git.Repository{
			Owner: r.Owner,
			Name:  "homebrew-brew",
		})
		if err != nil {
			return cli.NewExitError("Could not create homebrew repo client: "+err.Error(), -1)
		}

		err = bc.Commit(ctx, r, tag, formula)
		if err != nil {
			return cli.NewExitError("Could not update homebrew formula: "+err.Error(), -1)
		}

		zips = append(zips, bottles...)
	}

	s3c, err := s3.New("s3://releases.manifold.co")
	if err != nil {
		return cli.NewExitError("Could not create s3 client: "+err.Error(), -1)
	}

	for _, file := range zips {
		err := s3c.Put(ctx, &file)
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
