# Contributing

For general contribution and community guidelines, please see the [community repo](https://github.com/cyberark/community).

1. [Fork the project](https://help.github.com/en/github/getting-started-with-github/fork-a-repo)
1. [Clone your fork](https://help.github.com/en/github/creating-cloning-and-archiving-repositories/cloning-a-repository)
1. Make local changes to your fork by editing files
1. [Commit your changes](https://help.github.com/en/github/managing-files-in-a-repository/adding-a-file-to-a-repository-using-the-command-line)
1. [Push your local changes to the remote server](https://help.github.com/en/github/using-git/pushing-commits-to-a-remote-repository)
1. [Create new Pull Request](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/creating-a-pull-request-from-a-fork)

From here your pull request will be reviewed and once you've responded to all
feedback it will be merged into the project. Congratulations, you're a
contributor!

## Releasing

Releases should be created by maintainers only. To create and promote a
release, follow the instructions in this section.

### Update the changelog and notices

**NOTE:** If the Changelog and NOTICES.txt are already up-to-date, skip this
step and promote the desired build from the master branch.

1. Create a new branch for the version bump.
1. Based on the changelog content, determine the new version number and update.
1. Review the git log and ensure the [changelog](CHANGELOG.md) contains all
   relevant recent changes with references to GitHub issues or PRs, if possible.
1. Review the changes since the last tag, and if the dependencies have changed
   revise the [NOTICES](NOTICES.txt) to correctly capture the included
   dependencies and their licenses / copyrights.
1. Commit these changes - `Bump version to x.y.z` is an acceptable commit
   message - and open a PR for review.

### Release and Promote

1. Merging into the master branch will automatically trigger a release.
   If successful, this release can be promoted at a later time.
1. Jenkins build parameters can be utilized to promote a successful release
   or manually trigger aditional releases as needed.
1. Reference the [internal automated release doc](https://github.com/conjurinc/docs/blob/master/reference/infrastructure/automated_releases.md#release-and-promotion-process) for releasing and promoting.
1. Copy the `conjur-cli.rb` homebrew formula output by goreleaser
   to the [homebrew formula for Conjur CLI](https://github.com/cyberark/homebrew-tools/blob/main/conjur-cli.rb)
   and submit a PR to update the version of Conjur-Cli available in brew.
