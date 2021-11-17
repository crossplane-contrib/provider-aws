---
name: Release
about: Issue to track a release.
labels: release
---
<!--
Thank you for helping to improve Crossplane!

Please be sure to search for open issues before raising a new one. We use issues
for bug reports and feature requests. Please find us at https://slack.crossplane.io
for questions, support, and discussion.
-->

### Checklist

* [ ] Create the new release branch with minor version, i.e. `release-0.21` for `v0.21.0`.
  * You can use the existing branch for patch releases.
* [ ] Tag release by running `Tag` action in Github UI against release branch, i.e. `release-X` branch.
  * Use v-prefixed version, like `v0.21.0`, for both description and tag.
* [ ] Run `CI` action against release branch.
* [ ] Tag the next pre-release by running `Tag` action in Github UI against `master` branch, if it's not a patch release.
  * The tag should be `v0.22.0-rc.0` if you're releasing `v0.21.x`.
* [ ] Validate that you can install the published version, basic sanity check.
* [ ] Run `Promote` action to promote it in `alpha` channel if it's pre-`1.0`.
* [ ] Use Github UI to generate release notes.
  * Make sure to scan all PRs marked with `breaking-change` and add instructions for users to handle them.
* [ ] Create the next release issue with a title that has its date, **4 weeks** after the current release day.
* [ ] Announce in Slack, Twitter etc.