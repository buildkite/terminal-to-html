# Releasing terminal-to-html

Releases are published manually using the GitHub CLI (`gh`) and the archives
produced by [Buildkite](https://buildkite.com/buildkite/terminal-to-html).
Use Bash for the commands below, with `gh` authenticated to an account that can
create releases in this repository.

> [!IMPORTANT]
> This repository uses **immutable GitHub releases**. Attach and verify every
> binary asset **while the release is a draft**, then publish it. After
> publication, assets cannot be added, replaced, or deleted. Deleting the release
> does not allow its tag name to be reused.

## 1. Prepare the version and build

1. Open and merge a PR updating `baseVersion` in [`version.go`](version.go) to the
   next version. Do not reuse a published version.
2. From a clean checkout, update `main` and record the release version and commit:

   ```bash
   git switch main
   git pull --ff-only
   VERSION=$(make version)
   TAG="v$VERSION"
   COMMIT=$(git rev-parse HEAD)
   printf 'Release %s from %s\n' "$TAG" "$COMMIT"
   ```

3. Find the Buildkite build on `main` whose commit is exactly `$COMMIT`, and wait
   for the whole build to pass. Do not use the PR build or a build of an older
   commit. The pipeline runs tests, benchmarks, and `make dist`, then stores the
   release archives as build artifacts.

The publishing steps in [`.buildkite/pipeline.yml`](.buildkite/pipeline.yml) are
commented out. CI builds the archives but does **not** attach them to GitHub
releases. The legacy [release script](.buildkite/steps/release) is not an active
publishing workflow; use the draft-first procedure below.

## 2. Download and verify the artifacts

Download all eight archives from that build's **Build** job into `dist/`, keeping
their filenames unchanged. Use the CI artifacts rather than rebuilding locally.
The expected filenames are defined by [`Makefile`](Makefile):

```bash
ASSETS=(
  "dist/terminal-to-html-$VERSION-linux-amd64.gz"
  "dist/terminal-to-html-$VERSION-linux-i386.gz"
  "dist/terminal-to-html-$VERSION-linux-armel.gz"
  "dist/terminal-to-html-$VERSION-linux-arm64.gz"
  "dist/terminal-to-html-$VERSION-darwin-amd64.gz"
  "dist/terminal-to-html-$VERSION-darwin-arm64.gz"
  "dist/terminal-to-html-$VERSION-windows-i386.zip"
  "dist/terminal-to-html-$VERSION-windows-amd64.zip"
)

for asset in "${ASSETS[@]}"; do
  test -s "$asset" || { echo "Missing or empty asset: $asset"; exit 1; }
done
```

Before uploading:

- Compare each downloaded file's SHA-1 checksum with Buildkite's artifact
  metadata (`sha1sum` on Linux or `shasum` on macOS).
- Check archive integrity with `gzip -t` for each `.gz` and `unzip -t` for each
  `.zip`.
- Extract the binary for your platform and run it with `--version`; it must
  report `terminal-to-html version $VERSION`.

## 3. Create and review a draft

Write release notes in `/tmp/terminal-release-notes.md`, covering the changes
since the previous release. Follow the previous release's headings, link the
relevant PRs and full changelog comparison, and call out minimum Go version
changes when applicable.

Create the draft with all eight archives attached and pin it to the verified
commit, rather than whichever commit `main` happens to point to later:

```bash
gh release create "$TAG" "${ASSETS[@]}" \
  --draft \
  --target "$COMMIT" \
  --title "$TAG" \
  --notes-file /tmp/terminal-release-notes.md

gh release view "$TAG" \
  --json isDraft,targetCommitish,body,assets,url
```

Confirm that `isDraft` is `true`, the target commit is correct, and all eight
assets have the expected names and sizes. Review the release notes and download
the attached assets to verify their checksums match the original CI artifacts.
If a draft already exists, inspect it instead of creating another one; use
`gh release upload "$TAG"` followed by only the missing archive paths to attach
them while it is still a draft. If any upload fails, inspect the draft's asset
list and resolve the failure before publishing.

## 4. Publish only after verification

Publishing is the final step, after the notes and all assets are verified:

```bash
gh release edit "$TAG" --draft=false
gh release view "$TAG" --json isDraft,publishedAt,assets,url
```

Confirm the release is published and still lists all eight binary assets.
GitHub's automatically generated source archives are not substitutes for these
binaries. No separate Packagecloud publication is part of this procedure; its
pipeline step is also commented out.

## If a release was published without binaries

Do not delete or retag the release, disable immutability, or try to replace its
assets. Prepare a new patch version through the process above, attach its own
matching CI binaries to a draft, and publish only after verification. If useful,
edit the earlier release's notes to point users to the replacement release.

See [GitHub's immutable release documentation](https://docs.github.com/en/code-security/concepts/supply-chain-security/immutable-releases)
for the restrictions that apply after publication.
