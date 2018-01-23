# About

`klone` is a command line tool that makes it easy to fork and clone a repository locally.

# Installing

```bash
go get -u github.com/kris-nova/klone
klone klone
```

# Example

```bash
klone kubernetes
```

1. Klone will look for a git server for the project.
2. In the case of `kubernetes` we will detect **github.com**.
3. Klone will then attempt to find the organization via the GitHub API, which will be the same value: `kubernetes`.
4. Klone will then attempt to find the repository via the GitHub API, which will be the same value: `kubernetes`.
5. After finding `github.com/kubernetes/kubernetes` klone check and see if you (the authenticated user) has forked the repository yet.
6. If needed, klone will use the GitHub API to fork the repo for you.
7. Klone will detect the `Go` programming language, and use the offical `Go` implementation for checking out the program.
8. Klone will then check out the repository with the following remotes:


After a `klone` you should have the following `git remote -v` configuration

| Remote        | URL                                         |
| ------------- | ------------------------------------------- |
| origin        | git@github.com:$you/$repo                 |
| upstream      | git@github.com:$them/$repo                |


# GitHub Credentials

Klone will prompt you the first time you use the program for needed credentials.
On a first run klone will cache local configuration to `~/.klone/auth` after it creates a token via the API.
Klone will *never* store your passwords in plaintext.

# Testing

Export `TEST_KLONE_GITHUBUSER` and `TEST_KLONE_GITHUBPASS` with a GitHub user/pass for a test account.
I use handy dandy @knovabot for my testing.

Run the test suite

```
make test
```

# Environmental variables

| Variable                              | Behaviour                                              |
| ------------------------------------- | ------------------------------------------------------ |
|KLONE_WORKSPACE                        | If set, klone will klone here for simple klones only   |
|KLONE_GITHUBTOKEN                      | GitHub acccess token to use with GitHub.com            |
|KLONE_GITHUBUSER                       | GitHub user name to authenticate with                  |
|KLONE_GITHUBPASS                       | GitHub password to authenticate with                   |
|TEST_KLONE_GITHUBTOKEN                 | (Testing) GitHub acccess token to use with GitHub.com  |
|TEST_KLONE_GITHUBUSER                  | (Testing) GitHub user name to authenticate with        |
|TEST_KLONE_GITHUBPASS                  | (Testing) GitHub password to authenticate with         |