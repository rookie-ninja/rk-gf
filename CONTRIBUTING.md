<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Contributing](#contributing)
  - [Setup](#setup)
  - [Making Changes](#making-changes)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Contributing

We'd love your help make rk-gf the very best structured GoFrame bootstrapper library in Go!

If you'd like to add new features, please [open an issue][open-issue]
describing your proposal problem; discussing feature changes ahead of time makes
pull request review much smoother. In your issue, pull request, and any other
communications, please remember to treat your fellow contributors with
respect! We take our [code of conduct](CODE_OF_CONDUCT.md) seriously.

## Setup

[Fork][fork], then clone the repository:

```
git clone https://github.com/rookie-ninja/rk-gf.git
cd rk-gf
git remote add upstream https://github.com/rookie-ninja/rk-gf.git
git fetch upstream
```

Install rk-gf's dependencies:

```
go mod tidy
```

## Making Changes

Start by creating a new branch for your changes:

```
git checkout master
git fetch upstream
git rebase upstream/master
git checkout -b cool_new_feature
```

Make your changes, then ensure that `make lint` and `make test` still pass. If
you're satisfied with your changes, push them to your fork.

```
git push origin cool_new_feature
```

Then use the GitHub UI to open a pull request.

At this point, you're waiting on us to review your changes. We **try** to respond
to issues and pull requests within a few business days, and we may suggest some
improvements or alternatives. Once your changes are approved, one of the
project maintainers will merge them.

We're much more likely to approve your changes if you:

* Add tests for new functionality.
* Write a [good commit message][commit-message].
* Maintain backward compatibility.

[fork]: https://github.com/rookie-ninja/rk-gf/fork
[open-issue]: https://github.com/rookie-ninja/rk-gf/issues/new
[cla]: https://cla-assistant.io/rookie-ninja/rk-gf
[commit-message]: http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html