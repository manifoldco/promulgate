```
                                  __               __
.-----.----.-----.--------.--.--.|  |.-----.---.-.|  |_.-----.
|  _  |   _|  _  |        |  |  ||  ||  _  |  _  ||   _|  -__|
|   __|__| |_____|__|__|__|_____||__||___  |___._||____|_____|
|__|                                 |_____|
```

# promulgate - Manifold's tool to make things widely known

[Code of Conduct](./CODE_OF_CONDUCT.md) |
[Contribution Guidelines](./.github/CONTRIBUTING.md)

[![GitHub release](https://img.shields.io/github/tag/manifoldco/promulgate.svg?label=latest)](https://github.com/manifoldco/promulgate/releases)
[![Travis](https://img.shields.io/travis/manifoldco/promulgate/master.svg)](https://travis-ci.org/manifoldco/promulgate)
[![License](https://img.shields.io/badge/license-BSD-blue.svg)](./LICENSE.md)

## Overview

promulgate is used in manifold to release our cli tools. It:
- creates github releases from changelog contents
- adds built zip files to the github release
- uploads zip files to s3 (which backs https://releases.manifold.co)
- rebuilds the index files on s3

## Configuring a repository for promulgate

You'll need to set the following env vars:
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `GITHUB_TOKEN`

With those set, have your release step call `promulgate release <VERSION>`.

## Using to release

Make sure to keep your CHANGELOG.md up to date. When it's time to release,
create a new `Unreleased` section, and name the old one to match the release
tag. Commit this change, tag it with the release version, and push to master.
