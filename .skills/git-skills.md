---
name: git-skills
description: Opinionated constraints for better git commits and workflows.
---

# Git Skills

When invoked, apply these constraints for git operations and commit messages.


## Commit Format

- format(action):subject - Brief, lowercase subject
- Use present tense, imperative mood
- Keep subject under 50 characters
- Separate subject from body with blank line

## Commit Types

- feat: New feature
- fix: Bug fix
- docs: Documentation changes
- style: Formatting, no code change
- refactor: Code restructuring
- test: Adding tests
- chore: Maintenance tasks

## Content

- Explain why, not what
- Include breaking changes
- Use bullet points in body

## Branching

- Use kebab-case for branch names
- Include type: feature/x, fix/x, chore/x
- Keep branches short-lived
- Delete merged branches

## Workflow

- Rebase before merging
- Squash fixup commits
- Force push only to feature branches
- Never rewrite published history
