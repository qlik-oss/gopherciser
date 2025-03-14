# Contributing to gopherciser

You are more than welcome to contribute to gopherciser! Follow these guidelines and you will be ready to start:

 - [Code of conduct](#code-of-conduct)
 - [Bugs](#bugs)
 - [Features](#features)
 - [Documentation](#documentation)
 - [Git guidelines](#git)
 - [Signing the CLA](#cla)

## <a name="code-of-conduct"></a> Code of conduct

Please read and follow our [Code of conduct](https://github.com/qlik-oss/open-source/blob/master/CODE_OF_CONDUCT.md).

## <a name="bugs"></a> Bugs

Bugs can be reported by adding issues in the repository. Submit your bug fix by creating a Pull Request, following the [GIT guidelines](#git).

## <a name="features"></a> Features

Features can be requested by adding issues in the repository. If the feature includes new designs or bigger changes,
please be prepared to discuss the changes with us so we can cooperate on how to best include them.

Submit your feature by creating a Pull Request, following the [GIT guidelines](#git). Include any related documentation changes.

## <a name="documentation"></a> Documentation changes

Documentation changes can be requested by adding issues in the repository.

Submit your documentation changes by creating a Pull Request, following the [GIT guidelines](#git).
If the change is minor, you can submit a Pull Request directly without creating an issue first.

## <a name="git"></a> Git Guidelines

Generally, development should be done directly towards a branch in your own fork.

### Workflow

1\. Fork and clone the repository

```sh
git clone git@github.com:YOUR-USERNAME/gopherciser.git
```

2\. Create a branch in the fork

The branch should be based on the `master` branch in the master repository.

```sh
git checkout -b my-feature-or-bugfix master
```

3\. Commit changes on your branch

Commit changes to your branch, following the commit message format.

```sh
git commit -m "Add new action 'thinktime', used to simulate user thinking between clicks."
```

4\. Push the changes to your fork

```sh
git push -u myfork my-feature-or-bugfix
```

5\. Create a Pull Request

Before creating a Pull Request, make sure the following items are satisfied:

- CircleCI is green
- Commit message format is followed
- [CLA](#cla) is signed

In the Github UI of your fork, create a Pull Request to the `master` branch of the master repository.

If the branch has merge conflicts or has been outdated, please do a rebase against the `master` branch.

_WARNING: Squashing or reverting commits and force-pushing thereafter may remove GitHub comments on code that were previously made by you or others in your commits. Avoid any form of rebasing unless necessary._


### <a name="commit"></a> Commit message format

There are currently no conventions on how to format commit messages. We'd like you to follow some rules on the content however:

- Use the present form, e.g. _Add cleanup of socket on timeout_
- Be descriptive and avoid messages like _Minor fix_.
- If the change is breaking existing scenarios, add a _[breaking]_ tag in the message.

## <a name="cla"></a> Signing the CLA

We need you to sign our Contributor License Agreement (CLA) before we can accept your Pull Request. Visit this link for more information: https://github.com/qlik-oss/open-source/blob/master/sign-cla.md.
