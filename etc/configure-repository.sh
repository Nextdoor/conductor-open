#!/bin/bash -e

REPO_NAME='conductor'

# Create symlinks under .git/hooks to hook scripts in the githooks
# directory.
hooks="pre-commit"
git_hooks_dir=$(git rev-parse --git-dir)/hooks
target_dir=../../etc/githooks
for hook in $hooks; do
    ln -sf $target_dir/$hook $git_hooks_dir/$hook
done

# Configure remote tracking branches for rebase.
git config branch.master.rebase true
git config branch.autosetuprebase remote

# Configure aliases.
git config alias.pre-commit '!$(git rev-parse --git-dir)/hooks/pre-commit'
