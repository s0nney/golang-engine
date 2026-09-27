---
title: Git Essentials
quiz:
  - question: |
      You edit `amount.go` and run `git add amount.go`. Now `git diff` prints
      nothing. Where did your change go?
    options:
      - text: '`git add` committed it'
      - text: It's staged. `git diff` compares your files with the staging area, so staged changes only show up in `git diff --staged`
        correct: true
      - text: '`git add` discarded it'
      - text: '`git diff` only works on branches'
    explanation: |
      Git has three places: your working files, the staging area (what the
      next commit will contain) and the commits themselves. `git diff` shows
      unstaged changes, and `git diff --staged` shows what `git commit`
      would record.
  - question: Which of these should Ledgerly's `.gitignore` **not** ignore?
    options:
      - text: '`cover.out`'
      - text: '`ledgerly.test`'
      - text: '`testdata/fuzz/FuzzParseAmount/0a2a901998b6f11e`'
        correct: true
      - text: '`cpu.out`'
    explanation: |
      Failing inputs the fuzzer saved under `testdata/fuzz` are regression
      tests, and every plain `go test` replays them. Commit them. Coverage
      profiles, CPU profiles and test binaries are generated output: ignore
      those.
---

Tests tell you whether the code works *now*. **Git** remembers how it got there. It records snapshots of your project so you can see what changed, undo mistakes, and try ideas on a branch without breaking what works. Nearly every Go project, including the standard library, lives in a Git repository.

## A repository and the first commit

```text
$ cd ledgerly
$ git init
Initialized empty Git repository in /home/you/ledgerly/.git/
$ git status
On branch main
Untracked files:
	amount.go
	amount_test.go
	go.mod
$ git add .
$ git commit -m "Parse and format amounts"
```

A commit happens in two steps:

1. `git add` puts changes in the **staging area**: "these go in the next commit".
2. `git commit` records everything staged as one snapshot with a message.

The two steps let you commit part of your work. After running `go fix`, for example, you can stage and commit the mechanical changes on their own, separately from a behaviour change. The go fix lesson in the next chapter recommends exactly that.

## Seeing what changed

```text
$ git status              # which files changed, and which are staged
$ git diff                # unstaged changes, line by line
$ git diff --staged       # what the next commit contains
$ git log --oneline       # history, one commit per line
3f9c2a1 Reject overflowing amounts
b41d0e7 Parse and format amounts
```

Read `git diff --staged` before every commit. It's a thirty-second code review of yourself, and it catches debugging `fmt.Println`s and accidental edits.

## Branches

A **branch** is a line of work. Make one before an experiment:

```text
$ git switch -c faster-import     # create a branch and switch to it
  ... edit, test, commit ...
$ git switch main                 # back to the untouched version
$ git merge faster-import         # bring the work in once it's proven
```

This fits the benchmarking workflow well. Record a baseline on `main`, switch to the branch, record again, and compare with `benchstat`. If the numbers don't improve, delete the branch with `git branch -D faster-import`, and nothing on `main` ever changed. (For a quick before and after of uncommitted work, the benchmarks chapter used `git stash`, which sets changes aside and brings them back with `git stash pop`.)

## .gitignore

Some files should never be committed: build output and anything a tool can regenerate. List patterns in a `.gitignore` file at the root of the repository:

```text
# binaries from go build
/ledgerly
*.exe
*.test

# profiles and coverage
*.out
*.prof
coverage.html
```

Git then leaves those files out of `git status` and `git add .`. Things you **should** commit include `go.mod`, `go.sum`, everything under `testdata/` (golden files and fuzz corpus entries are tests) and the `.gitignore` itself.

## Undoing things

- `git restore amount.go` throws away unstaged changes to a file.
- `git restore --staged amount.go` unstages it but keeps the edit.
- `git revert <commit>` makes a new commit that undoes an old one, which is safe even after you've shared your history.

Git has far more to offer (remotes, rebasing, bisecting), but these commands cover a day of real work.

## Further reading

- [Pro Git](https://git-scm.com/book/en/v2), free online, chapters 1 to 3.
