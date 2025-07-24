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

### Working with Unreleased Changes in `conjur-api-go`

#### For CI/Jenkins Builds
1. Jenkins automatically switches dependencies to internal Enterprise versions, allowing CI to build against private repos without requiring public releases.
- *Note:* Changes still need to be merged to `main` in the internal `conjur-api-go` repository before the downstream repositories will be able to use them.
2. In the downstream project (e.g., this project), add replace statements to the bottom of `go.mod` to ensure that the internal dependencies are pulled in when running the CI pipeline:
  ```
   replace github.com/cyberark/conjur-api-go => github.com/cyberark/conjur-api-go latest
   ```
- *Note:* the custom replace statements and CI business logic are specific to CyberArk internal contributors
- See the [secrets provider go.mod](https://github.com/cyberark/secrets-provider-for-k8s/blob/main/go.mod) for examples of proper replace statements

#### For Local Development
1. Locally, you need to follow standard Go practice of replacing the dependency in `go.mod ` with the version in a local directory.
- See [Go Documentation: Requiring Module Code in a Local Directory](https://go.dev/doc/modules/managing-dependencies#local_directory)

## Releasing

Releases should be created by maintainers only. To create and promote a
release, follow the instructions in this section.

### Update the changelog and notices (if necessary)
1. Update the `CHANGELOG.md` file with the new version and the changes that are included in the release.
1. Update `NOTICES.txt`
    ```sh-session
    go install github.com/google/go-licenses@latest
    # Verify that dependencies fit into supported licenses types.
    # If there is new dependency having unsupported license, that license should be
    # included to notices.tpl file in order to get generated in NOTICES.txt.
    $(go env GOPATH)/bin/go-licenses check ./... \
      --allowed_licenses="MIT,ISC,Apache-2.0,BSD-3-Clause,BSD-2-Clause,MPL-2.0" \
      --ignore $(go list std | awk 'NR > 1 { printf(",") } { printf("%s",$0) } END { print "" }')
    # If no errors occur, proceed to generate updated NOTICES.txt
    $(go env GOPATH)/bin/go-licenses report ./... \
      --template notices.tpl \
      --ignore github.com/cyberark/conjur-cli-go \
      --ignore $(go list std | awk 'NR > 1 { printf(",") } { printf("%s",$0) } END { print "" }') \
      > NOTICES.txt
    ```
1. Commit these changes - `Bump version to x.y.z` is an acceptable commit
   message - and open a PR for review.

### Release and Promote

1. Merging into the master branch will automatically trigger a release.
   If successful, this release can be promoted at a later time.
1. Jenkins build parameters can be utilized to promote a successful release
   or manually trigger aditional releases as needed.
1. Reference the [internal automated release doc](https://github.com/conjurinc/docs/blob/master/reference/infrastructure/automated_releases.md#release-and-promotion-process) for releasing and promoting.
