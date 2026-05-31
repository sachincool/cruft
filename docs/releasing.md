# Releasing cruft

End-to-end flow for cutting a release and pushing the Homebrew formula.

## One-time setup

1. Create the tap repo on GitHub: **`sachincool/homebrew-tap`** (public, empty,
   with `Formula/` directory will be added by goreleaser).
2. Create a Personal Access Token with **`contents: write`** on `homebrew-tap`.
   Classic PAT or a fine-grained token, repo-scoped to `homebrew-tap`.
3. Add it as a secret on **`sachincool/cruft`** named **`HOMEBREW_TAP_GITHUB_TOKEN`**.
   The release workflow (`.github/workflows/release.yml`) exposes it as an env var
   that goreleaser reads via `{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}` in
   `.goreleaser.yml`'s `brews:` section.

## Cutting a release

```sh
# Bump version, push the tag.
git tag v0.2.0 -m "v0.2.0"
git push origin v0.2.0
```

The `release` workflow fires on `v*` tags:

1. Builds `darwin/arm64`, `darwin/amd64`, `linux/arm64`, `linux/amd64`.
2. Uploads `.tar.gz` archives + `checksums.txt` to the GitHub release.
3. Generates `Formula/cruft.rb` and pushes it to `sachincool/homebrew-tap`.

Users then get the new version via `brew upgrade cruft` (or fresh install via
`brew install sachincool/tap/cruft`).

## Verifying before tagging

```sh
make snapshot          # builds dist/ locally, doesn't publish
ls dist/               # archives + checksums + formula preview
goreleaser check       # validates .goreleaser.yml
```

## If the brew push fails

Most common: the token is missing, expired, or scoped wrong.

- Workflow log: look for the `brews` step in the goreleaser output.
- Verify the secret exists on `sachincool/cruft`:
  `gh secret list --repo sachincool/cruft`.
- Verify the token has `contents: write` on `homebrew-tap`.
- Re-run the workflow without re-tagging:
  `gh workflow run release.yml -F ref=v0.2.0` (only works if the workflow
  declares `workflow_dispatch`; otherwise delete + recreate the tag).

## Manual formula push (fallback)

If you need to ship without the bot:

```sh
make snapshot
cp dist/homebrew/Formula/cruft.rb ../homebrew-tap/Formula/cruft.rb
cd ../homebrew-tap && git commit -am "cruft: v0.2.0" && git push
```
