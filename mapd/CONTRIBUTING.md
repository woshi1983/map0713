# Contributing to MAP

## Welcome

MAP is developed in the open, and is constantly being improved by our **users, contributors, and maintainers**.
It is because of you that we can bring great software to the community.

This guide provides information on filing issues and guidelines for open source contributors. **Please leave comments / suggestions if you find something is missing or incorrect.**

## Reporting Issues

It is a great way to contribute to MAP by reporting an issue. Well-written and complete bug reports are always welcome! Please open an issue and follow the template to fill in required information.

Before opening any issue, please look up the existing [issues](https://gitlab.eng.vmware.com/cn-se-solution/mapd/-/issues) to avoid submitting a duplication.
If you find a match, you can "subscribe" it to get notified on updates. If you have additional helpful information about the issue, please leave a comment.

When reporting issues, always include:

* Version of docker engine and docker-compose
* Configuration files of MAP
* Log files in /var/log/map/ 

Because the issues are open to the public, when submitting the log and configuration files, be sure to remove any sensitive information, e.g. user name, password, IP address, and company name. You can
replace those parts with "REDACTED" or other strings like "****".

Be sure to include the steps to reproduce the problem if applicable. It can help us understand and fix your issue faster.

## Contribution Guidelines

### Repository Structure

Here is the basic structure of code base. Some of the key folders / files are commented for your references.
```
.
...
├── cmd           # Command-line main packages
├── docs          # Contain documents, schema definitions, and other helpful things which are contributed by the community
├── example       # Resource for building and setting up Harbor environment
├── pkg           # Primary Golang packages. Source code for the main business logic.
├── scripts       # Contains scripts
├── src           # Web source code folder
└── tests         # Test cases for API / e2e testings
...
```

The folder graph below shows the structure of the source code folder `pkg`, which will be your primary working directory. The key folders are also commented.

```
.
├── auth               # handle authentication & authorization requests
├── conductor          # primary controllers for managing cluster (and related) resources
│   └── cluster        # handlers for provisioning cluster & extension
├── graphql            # source code for the GraphQL (GQL) API. 
│   ├── abstract       # GQL scalar types
│   ├── model          # GQL schema definitions
│   └── resolver       # GQL resolver definitions.
├── liberr             # helpers for handling error stream
├── model              # SQL model definitions
├── submodule          # entrypoint of the main business logic.
└── util               # helpful utilities to handle data or interact with external systems
    ├── avi            # AVI client
    ├── compressutil
    ├── cryptoutil
    ├── harbor         # Harbor client
    ├── httputil
    ├── identicon      # utility for generating avatars
    ├── k8s            # kubernetes client
    ├── netutil
    ├── randutil
    ├── sshutil
    ├── tanzu          # customized tanzu client
    └── textutil
```

### Setup Development Environment

#### Go

MAP backend is written in [Go](http://golang.org/). If you don't have a backend service development environment, please [set one up](https://golang.org/doc/install).

System requirement for mapd:

|Items|Requirements|
|:-|:-|
|OS|Apple macOS 10.12+, GNU Linux 2.6.23+|
|Platform|`darwin`, `linux`|
|Arch|`amd64`|
|Processor|dual-core 1.5GHz 64-bit processor or above|
|Memory|2GB+|

#### Web

### Pull requests

Pull requests (PR) are always welcome, even they are small fixes like typos or a few lines of code changes. If there will be significant effort, please first document as an issue and get the discussion going before starting to work on it.

Please submit a PR to contain changes bit by bit. A PR consisting of a lot features and code changes may be hard to review. It is recommended to submit PRs in a incremental fasion.

### Design new features

You can propose new designs for existing features. You can also design entirely new features. Please do open an issue for discussion first. This is necessary to ensure the overall architecture is consistent and to avoid duplicated work in the roadmap.

### Conventions

Fork MAP's repository and make changes on your own fork in a new branch. The branch should be named  XXX-description where XXX is the number of the issue. Please run the full test suite on your branch before creating a PR.

Always run [golint](https://github.com/golang/lint) on source code before
committing your changes.

Pull requests should be rebased on top of master without multiple branches mixed into the PR. If your pull requests do not merge cleanly, use `rebase master` in your branch rather than `merge master`.

Update the documentation if you are creating or changing features. Good documentation is as important as the code itself.

Before you submitting any pull request, always squash your commits into logical units of change. A logical unit of change is defined as a set of codes and documents that should be treated as a whole. When possible, compact your commits into one. The commands to use are `git rebase -i` and/or `git push -f`. 

Description of a pull request should refer to all the issues that it addresses. Remember to put a reference to issues (such as `Closes #XXX` and `Fixes #XXX`) in commits so that the issues can be closed when the PR is merged.

### Signing CLA
If it is the first time you are making a PR, be sure to sign the contributor license agreement (CLA) online.
