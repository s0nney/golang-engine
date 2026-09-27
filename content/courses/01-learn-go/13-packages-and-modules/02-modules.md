---
title: Modules and Dependencies
quiz:
  - question: You run `go mod init github.com/textio/smsapp` with Go 1.27 installed. Which `go` line does the new `go.mod` contain?
    options:
      - text: '`go 1.27.0`'
      - text: '`go 1.26.0`'
        correct: true
      - text: '`go 1.0`'
      - text: There is no `go` line
    explanation: |
      Since Go 1.26, `go mod init` writes the *previous* minor version into
      `go.mod`. That makes a new module usable by people who haven't
      upgraded to the very latest release yet. You can raise it later with
      `go get go@1.27.0`.
  - question: What is the job of `go.sum`?
    options:
      - text: It lists the functions your module exports
      - text: It records cryptographic checksums of your dependencies so downloads can be verified
        correct: true
      - text: It stores your test results
      - text: It's a backup copy of `go.mod`
    explanation: |
      `go.sum` holds a hash for every dependency version. If a download ever
      differs from what was recorded, the `go` command refuses to use it.
      Commit it alongside `go.mod`.
---

A package is a folder of code. A **module** is a collection of packages that are versioned and released together, like a whole project. Every Go project you build will be a module.

## `go mod init`

To start a new project, make a folder and run `go mod init` with the module's **path**, the name other code uses to import it. By convention, it's where the code lives online:

```text
$ mkdir smsapp && cd smsapp
$ go mod init github.com/textio/smsapp
go: creating new go.mod: module github.com/textio/smsapp
```

That creates a `go.mod` file:

```text
module github.com/textio/smsapp

go 1.26.0
```

- The `module` line gives the module path. Packages inside the module are imported by the module path plus their folder: `github.com/textio/smsapp/billing`.
- The `go` line says the minimum Go version needed to build the module.

Notice that Go 1.27 wrote `go 1.26.0`, not `go 1.27.0`. Since Go 1.26, `go mod init` deliberately picks the **previous** minor version, so that a brand-new module can be used by people who haven't upgraded yet. If you need a newer language feature, raise it with `go get go@1.27.0`. (Some distribution builds of Go, such as Arch Linux's, write the exact installed version instead, so don't panic if you see `go 1.27.1`.)

The module path doesn't have to be a real URL. For a program you'll never publish, something like `go mod init smsapp` works fine.

## Adding a dependency with `go get`

Code from other modules is called a **dependency**. Say Textio wants a third-party package for parsing phone numbers. Use `go get` with the module path:

```text
$ go get github.com/nyaruka/phonenumbers@latest
go: added github.com/nyaruka/phonenumbers v1.8.1
```

(Your version number will likely differ.) This does three things:

1. Downloads the module (into a shared cache on your machine, not your project folder).
2. Adds a `require` line to `go.mod`:

```text
module github.com/textio/smsapp

go 1.26.0

require github.com/nyaruka/phonenumbers v1.8.1
```

3. Records **checksums** of the downloaded code in a file called `go.sum`.

Then you import the package in your code by its path, just like your own packages:

```go
import "github.com/nyaruka/phonenumbers"
```

## `go.sum`

`go.sum` lists a cryptographic hash of every dependency version you use. If anyone (or anything) tampers with a dependency, the hashes won't match and the build fails. Commit both `go.mod` and `go.sum` to version control. Don't edit `go.sum` by hand.

## Versions

Go modules use **semantic versioning**: `v1.8.1` means major version 1, minor version 8, patch 1.

- **Patch** releases fix bugs.
- **Minor** releases add features without breaking anything.
- **Major** releases may break things. In Go, a new major version (v2 and up) gets a new module path ending in `/v2`, so old and new can coexist. That's why you'll see imports like `math/rand/v2` in the standard library.

You can ask for a specific version with `@`: `go get example.com/pkg@v1.2.3`. `@latest` gets the newest release.

## `go mod tidy`

As your code changes, you'll add and remove imports. `go mod tidy` syncs `go.mod` and `go.sum` with what your code actually imports: it adds anything missing and removes anything unused.

```text
$ go mod tidy
```

Run it before committing, and whenever the `go` command tells you requirements are missing.

## Prefer the standard library

Go's standard library is large and high quality. Before adding a dependency, check whether `strings`, `slices`, `maps`, `net/http`, `encoding/json` and friends already do what you need. Every dependency is code you're trusting and will need to keep up to date.

## Further reading

- [Tutorial: Create a Go module](https://go.dev/doc/tutorial/create-module)
- [Managing dependencies](https://go.dev/doc/modules/managing-dependencies)
