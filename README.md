# Seiso

[![dockeri.co](https://dockeri.co/image/appuio/seiso)](https://hub.docker.com/r/appuio/seiso)

[![](https://img.shields.io/github/workflow/status/appuio/seiso/Build)](https://github.com/appuio/seiso/actions)
[![](https://img.shields.io/github/v/release/appuio/seiso?include_prereleases)](https://github.com/appuio/seiso/releases)
[![](https://img.shields.io/github/issues-raw/appuio/seiso)](https://github.com/appuio/seiso/issues)
[![](https://img.shields.io/github/issues-pr-raw/appuio/seiso)](https://github.com/appuio/seiso/pulls)
[![](https://img.shields.io/github/license/appuio/seiso)](https://github.com/appuio/seiso/blob/master/LICENSE)

Inspired by Robert C. Martin's book, [Clean Code](https://www.investigatii.md/uploads/resurse/Clean_Code.pdf), foreword, page xx:

> Seisō (清掃), Japanese for “cleaning” (think “shine” in English):
> Keep the workplace free of hanging wires, grease, scraps, and waste.

## Clean up your container registry and Kubernetes resources

*Seiso* is a CLI client that helps you clean up container resources.

* Unused images in your container registry (identified by Image Stream Tags
  in an [Image Stream](https://blog.openshift.com/image-streams-faq/) in OpenShift)

## Why should I use this tool?

*Seiso* is intended to be used in application lifecycle management on application
container platforms and with automatic deployment processes (CI/CD pipelines).

Kubernetes distributions and tools allow you to easily create or add new
resources that allow sophisticated, well-designed deployment processes.
For example, you may want to tag every application container image you build
with the [Git commit SHA-1 hash](https://git-scm.com/book/en/v2/Git-Tools-Revision-Selection)
corresponding to the revision in source control that it was built from.

While this is all convenient, those features were designed to create resources
but not to clean them up again. As a result, those resources will pile up, incur
additional cost for storage, cause confusion, or on (shared) application container
environments you may hit a quota limit after a while.

```
The ImageStream "application" is invalid: []: Internal error: ImageStream.image.openshift.io "application" is forbidden:
 requested usage of openshift.io/image-tags exceeds the maximum limit per openshift.io/ImageStream (51 > 50)
```

## How does it work?

*Seiso* uses different strategies to identify resources to be cleaned up.

1. It analyzes a Git repository and compares its history with the target
   image registry, removing old and unused image tags according to customizable rules.

1. It can be used more aggressively by deleting dangling image tags, "orphans",
   that happen to exist when the Git history is altered (e.g. by force-pushing).

*Seiso* is opinionated, e.g. with respect to naming conventions of image tags,
either by relying on a long Git SHA-1 value (`namespace/app:a3d0df2c5060b87650df6a94a0a9600510303003`)
or a Git tag following semantic versioning (`namespace/app:v1.2.3`).

The cleanup **runs in dry-mode by default**. Only when the `--force` flag is specified,
it will actually delete the identified resources. This should prevent accidental
deletions during verifications or test runs.

## Caveats and known issues

* Currently only OpenShift image registries are supported. In future, more
  resource types are planned to be supported for cleanup.

* **Please watch out for shallow clones**, as the Git history might be missing,
  it would in some cases also undesirably delete image tags.

## Usage

The following examples assume the namespace `namespace`, and and image stream
named `app`. For the image cleanup to work, you need to be logged in to the target
cluster, as the tool will indirectly read your kubeconfig file.

For the following examples, we will assume the following Git history,
with `c6` being the latest commit in branch `c`:

```
a1
a2
- b3
- b4
a5
- c6
```

In all cases, it is assumed that the Git repository is already checked out in the desired branch.

### Example: Keep the latest 2 image tags

Let's assume target branch is `a`:

```console
seiso images history namespace/app --keep 2
```
Only the image tag `a1` will be deleted (only current branch is compared).

### Example: Keep no image tags

```console
seiso images history namespace/app --keep 0
```
This would delete `a1` and `a2`, but *not* `a5`, as this image is being actively used by a Pod.

### Example: Delete orphaned images

```console
seiso images orphans namespace/app --older-than 7d
```
This will delete `a1`, `a2`, `b3` and `b4`, if we assume that `a5` is being actively used, and `c6` is younger than 7d
(image tag push date, not commit date).

That means it will also look in other branches too. It also deletes amended/force-pushed commits, which do not show up
in the history anymore, but would still be available in the registry.

This is very useful in cases where the images from feature branches are being pushed to a `dev` namespace, but need to
be cleaned up after some time. In the `production` namespace, we can apply different cleanup rules.

### Example: Delete versioned image tags

Let's assume we have image tagged according to semver:
```
v1.9.3
v1.10.0
v1.10.1
```

```console
seiso images history namespace/app --keep 2 --tags
```
This would delete `v1.9.3` as expected, since the `--sort` flag is `version` by default (including support for v prefix).
If `alphabetic`, the order for semver tags is reversed (probably undesired). For date-based tags, `alphabetic` sorting
flag might be better suitable, e.g. `2020-03-17`.

## Migrate from legacy cleanup plugin

Projects using the legacy `oc` cleanup plugin can be migrated to `seiso` as follows

```console
oc -n "$OPENSHIFT_PROJECT" plugin cleanup "$APP_NAME" -p "$PWD" -f=y
```
becomes:
```console
seiso images history "$OPENSHIFT_PROJECT/$APP_NAME" --force
```

## Development

### Requirements

* go 1.13
* goreleaser
* Docker

### Build

```
goreleaser release --snapshot --rm-dist --skip-sign
```

### Run
```
./dist/seiso_linux_amd64/seiso --help
# or
docker run --rm -it appuio/seiso:<tag>
```

## Release

Push a git tag with the scheme `vX.Y.Z` (semver).

## License

This project is BSD 3-Clause licensed (see LICENSE file for details).

This project uses other OpenSource software as listed in `go.mod` and indirectly in `go.sum` files. All their licenses apply too.
