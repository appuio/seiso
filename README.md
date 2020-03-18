# Image Cleanup Client

[![dockeri.co](https://dockeri.co/image/appuio/image-cleanup)](https://hub.docker.com/r/appuio/image-cleanup)

![](https://img.shields.io/github/workflow/status/appuio/image-cleanup/Build)
![](https://img.shields.io/github/v/release/appuio/image-cleanup?include_prereleases)
![](https://img.shields.io/github/issues-raw/appuio/image-cleanup)
![](https://img.shields.io/github/issues-pr-raw/appuio/image-cleanup)
![](https://img.shields.io/github/license/appuio/image-cleanup)

## General

The image cleanup client is used to clean up container images in an image registry (currently only OpenShift image registries are supported).

The tool is intended to be used closely within an application lifecycle management (Review apps). It analyzes a git repository and compares the history with the target image registry, removing old and unused image tags according to customizable rules.

The tool can also be used more aggressively by deleting unkown image tags too ("orphans"). See more usage examples below.

This tool is opinionated in the naming of image tags, as the image tags have to follow certain naming rules for this to work. E.g. stick to tagging with Git SHA value: `namespace/app:a3d0df2c5060b87650df6a94a0a9600510303003` or Git Tag: `namespace/app:v1.2.3`.

This helps to save space (e.g. your private registry with billed S3 storage) or quota threshold because obsolete images are being removed from the registry.

The cleanup **runs in dry-mode by default**, only when the `--force` flag is specified, it will delete the actual image tags. This prevents accidental deletions when testing.

## Usage

The following examples assume the namespace `namespace`, and and image stream named `app`. For the image cleanup to work, you need to be logged in to the target cluster, as the tool will indirectly read your kubeconfig file.

For the following examples, we will assume the following Git history, with `c6` being the latest commit in branch `c`:

```
a1
a2
- b3
- b4
a5
- c6
```

In all cases, it is assumed that the git repository is already checked out in the desired branch.

**Please watch out for shallow clones**, as the Git history might be missing, it would in some cases also undesirably delete image tags.

### Example: Keep the latest 2 image tags

Let's assume target branch is `a`:

```console
image-cleanup history namespace/app --keep 2
```
Only the image tag `a1` will be deleted (only current branch is compared).

### Example: Keep no image tags

```console
image-cleanup history namespace/app --keep 0
```
This would delete `a1` and `a2`, but *not* `a5`, as this image is being actively used by a Pod.

### Example: Delete orphaned images

```console
image-cleanup orphans namespace/app --older-than 7d
```
This will delete `a1`, `a2`, `b3` and `b4`, if we assume that `a5` is being actively used, and `c6` is younger than 7d (image tag push date, not commit date).

That means it will also look in other branches too. It also deletes amended/force-pushed commits, which do not show up in the history anymore, but would still be available in the registry.

This is very useful in cases where the images from feature branches are being pushed to a `dev` namespace, but need to be cleaned up after some time. In the `production` namespace, we can apply different cleanup rules.

### Example: Delete versioned image tags

Let's assume we have image tagged according to semver:
```
v1.9.3
v1.10.0
v1.10.1
```

```console
image-cleanup history namespace/app --keep 2 --tags
```
This would delete `v1.9.3` as expected, since the `--sort` flag is `version` by default (including support for v prefix). If `alphabetic`, the tool might find that `v1.9.3` is newer than `v1.10.0`. For date-based tags, `alphabetic` sorting flag might be better suitable, e.g. `2020-03-17`.

## Migrate from legacy cleanup plugin

Projects using the legacy `oc` cleanup plugin can be migrated to `image-cleanup` as follows

```console
oc -n "$OPENSHIFT_PROJECT" plugin cleanup "$APP_NAME" -p "$PWD" -f=y
```
becomes:
```console
image-cleanup history "$OPENSHIFT_PROJECT/$APP_NAME" --force
```

## Development

### Requirements

* go 1.13
* goreleaser

### Build

```
goreleaser --snapshot --rm-dist
```

### Run
```
./dist/image-cleanup_linux_amd64/image-cleanup --help
# or
docker run --rm -it appuio/image-cleanup:<tag>
```

## Release

Push a git tag with the scheme `vX.Y.Z` (semver).

## License

This project is BSD 3-Clause licensed (see LICENSE file for details).

This project uses other OpenSource software as listed in `go.mod` and indirectly in `go.sum` files. All their licenses apply too.
