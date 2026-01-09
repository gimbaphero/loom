# Branch Protection Rules Configuration

This document explains how to configure branch protection rules for the Loom repository to enforce quality standards and prevent direct pushes to main.

## Table of Contents

- [Overview](#overview)
- [Configuration Steps](#configuration-steps)
- [Main Branch Protection](#main-branch-protection)
- [Additional Settings](#additional-settings)
- [Verification](#verification)

## Overview

Branch protection rules enforce:
- ✅ All code must go through pull requests
- ✅ All CI checks must pass before merge
- ✅ Code review required from maintainers
- ✅ No force pushes to main
- ✅ No direct commits to main
- ✅ Up-to-date branch before merge

## Configuration Steps

### Access Repository Settings

1. Go to the Loom repository: https://github.com/teradata-labs/loom
2. Click **Settings** (requires admin access)
3. Navigate to **Branches** in the left sidebar
4. Click **Add branch protection rule** (or edit existing rule)

## Main Branch Protection

### Branch Name Pattern

```
main
```

### Required Settings

#### 1. Require a Pull Request Before Merging

✅ **Enable**: Require a pull request before merging

**Settings:**
- ✅ **Require approvals**: `1` (at least one approval from maintainer)
- ✅ **Dismiss stale pull request approvals when new commits are pushed**
- ✅ **Require review from Code Owners** (if CODEOWNERS file exists)
- ❌ **Require approval of the most recent reviewable push** (optional, recommended for critical repos)

**Purpose**: Enforces code review process and prevents unreviewed code from reaching main.

---

#### 2. Require Status Checks to Pass Before Merging

✅ **Enable**: Require status checks to pass before merging

**Required Status Checks** (all must pass):

Select all checks from `.github/workflows/ci.yml`:
- ✅ `Proto Lint & Breaking Changes`
- ✅ `Go Lint`
- ✅ `Unit Tests` (with race detection)
- ✅ `Build ${{ matrix.os }}` (all OS matrix combinations)
  - `Build ubuntu-latest`
  - `Build macos-latest`
  - `Build windows-latest`
- ✅ `Security Scan`
- ✅ `All CI Checks Passed`

**Additional Settings:**
- ✅ **Require branches to be up to date before merging**
  - Forces branches to merge latest main before CI runs
  - Prevents merge conflicts
  - Ensures CI runs against final merge state

**Purpose**: Ensures all automated checks pass before code reaches main.

---

#### 3. Require Conversation Resolution Before Merging

✅ **Enable**: Require conversation resolution before merging

**Purpose**: Ensures all reviewer feedback is addressed before merge.

---

#### 4. Require Signed Commits

⚠️ **Optional but Recommended**: Require signed commits

**Purpose**: Verifies commit authenticity via GPG signatures.

**Note**: Requires contributors to set up GPG signing. May be too restrictive for open-source projects. Recommended for enterprise/internal projects.

---

#### 5. Require Linear History

✅ **Enable**: Require linear history

**Purpose**: Enforces clean git history by requiring rebase or squash merges (no merge commits).

**Recommended Merge Strategy:**
- Squash and merge (keeps history clean, one commit per PR)
- OR Rebase and merge (preserves individual commits)

---

#### 6. Require Deployments to Succeed Before Merging

❌ **Disable**: (unless you have deployment workflows)

**Purpose**: For repos with deployment pipelines. Not applicable to Loom initially.

---

#### 7. Lock Branch

❌ **Disable**: Lock branch

**Purpose**: Makes branch read-only. Only use for archived branches.

---

#### 8. Do Not Allow Bypassing the Above Settings

✅ **Enable**: Do not allow bypassing the above settings

**CRITICAL**: Prevents administrators from bypassing protection rules.

**Exception**: Keep disabled if you need emergency hotfix capability (maintainers can bypass in emergencies). Enable for maximum safety.

---

#### 9. Restrict Who Can Push to Matching Branches

✅ **Enable**: Restrict who can push to matching branches

**Allowed to push:**
- Leave empty (no one can push directly)
- OR specify CI/CD service accounts only (for automated releases)

**Allowed to force push:**
- Leave empty (no one can force push)

**Purpose**: Prevents direct pushes to main. All code must go through PRs.

---

#### 10. Allow Force Pushes

❌ **Disable**: Allow force pushes

**Purpose**: Prevents history rewriting on main branch.

---

#### 11. Allow Deletions

❌ **Disable**: Allow deletions

**Purpose**: Prevents accidental branch deletion.

---

## Summary Configuration

```yaml
# Branch Protection Rule: main
require_pull_request:
  required_approving_review_count: 1
  dismiss_stale_reviews: true
  require_code_owner_reviews: false  # Enable if CODEOWNERS exists

require_status_checks:
  strict: true  # Require branches to be up to date
  contexts:
    - Proto Lint & Breaking Changes
    - Go Lint
    - Unit Tests
    - Build ubuntu-latest
    - Build macos-latest
    - Build windows-latest
    - Security Scan
    - All CI Checks Passed

require_conversation_resolution: true
require_signed_commits: false  # Optional
require_linear_history: true
enforce_admins: true  # CRITICAL
restrictions:
  users: []  # No direct push access
  teams: []
allow_force_pushes: false
allow_deletions: false
```

## Additional Settings

### Repository Settings

Navigate to **Settings** → **General**:

#### Pull Requests

✅ **Allow squash merging** (recommended)
- Default commit message: `Pull request title`
- Keeps history clean (one commit per PR)

❌ **Allow merge commits** (optional)
- Only if you want to preserve all individual commits

✅ **Allow rebase merging** (optional)
- Preserves individual commits, linear history

✅ **Automatically delete head branches**
- Cleans up merged PR branches automatically

#### Merge Button

✅ **Allow auto-merge**
- PRs can be set to auto-merge once checks pass

### Code Security and Analysis

Navigate to **Settings** → **Code security and analysis**:

✅ **Dependency graph**: Enabled
✅ **Dependabot alerts**: Enabled
✅ **Dependabot security updates**: Enabled
✅ **Dependabot version updates**: Enabled (uses `.github/dependabot.yml`)
✅ **Secret scanning**: Enabled
✅ **Push protection**: Enabled (prevents pushing secrets)

## Verification

### Test Branch Protection

1. **Try to push directly to main** (should fail):
   ```bash
   git checkout main
   echo "test" > test.txt
   git add test.txt
   git commit -m "test direct push"
   git push origin main
   # Expected: Error - protected branch
   ```

2. **Create a test PR**:
   ```bash
   git checkout -b test/branch-protection
   echo "test" > test.txt
   git add test.txt
   git commit -m "test: verify branch protection"
   git push origin test/branch-protection
   # Create PR on GitHub
   # Verify: Cannot merge until checks pass and review obtained
   ```

3. **Verify status check requirements**:
   - Open any PR
   - Check "Merge" button is disabled until:
     - ✅ All CI jobs pass
     - ✅ At least 1 approval
     - ✅ All conversations resolved
     - ✅ Branch is up to date

### Common Issues

**Issue**: "Required status checks not found"
**Solution**: Make sure `.github/workflows/ci.yml` is merged to main first, then configure branch protection.

**Issue**: "Cannot merge - branch is out of date"
**Solution**: Click "Update branch" button or rebase locally:
```bash
git fetch origin
git rebase origin/main
git push --force-with-lease
```

**Issue**: "Cannot merge - pending status checks"
**Solution**: Wait for all CI jobs to complete. Check Actions tab for failures.

## Maintainer Exceptions

### Emergency Hotfixes

If `enforce_admins` is disabled, maintainers can:
1. Temporarily disable branch protection
2. Push emergency hotfix
3. Re-enable branch protection

**Process:**
```bash
# 1. Disable branch protection (GitHub UI)
# 2. Push hotfix
git checkout main
git cherry-pick <commit>
git push origin main
# 3. Re-enable branch protection
# 4. File post-mortem issue explaining emergency bypass
```

### Legitimate Bypass Scenarios

- **Emergency security patches** (document in issue)
- **CI infrastructure failures** (document in issue)
- **Critical production outages** (document in issue)

**Always file an issue explaining why bypass was necessary.**

## References

- [GitHub Branch Protection Documentation](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/defining-the-mergeability-of-pull-requests/about-protected-branches)
- [GitHub Status Checks](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/collaborating-on-repositories-with-code-quality-features/about-status-checks)
- [GitHub Code Review](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/reviewing-changes-in-pull-requests/about-pull-request-reviews)

---

## Quick Setup Checklist

Use this checklist when configuring branch protection:

- [ ] Navigate to Settings → Branches
- [ ] Add protection rule for `main`
- [ ] ✅ Require pull request with 1 approval
- [ ] ✅ Dismiss stale reviews
- [ ] ✅ Require status checks (all 7 CI jobs)
- [ ] ✅ Require up-to-date branch
- [ ] ✅ Require conversation resolution
- [ ] ✅ Require linear history
- [ ] ✅ Restrict push access (empty list)
- [ ] ✅ Enforce for administrators
- [ ] ❌ Disable force pushes
- [ ] ❌ Disable deletions
- [ ] Save changes
- [ ] Test with PR (verify cannot merge without approval + passing CI)
- [ ] Enable Dependabot, secret scanning, push protection
- [ ] Document any emergency bypass in issue tracker

---

**Once configured, all code changes will require:**
1. Feature branch + PR
2. Passing CI (proto, lint, test, build, security)
3. Code review approval
4. Conversation resolution
5. Up-to-date with main

**No one (including admins) can push directly to main.** 🔒
