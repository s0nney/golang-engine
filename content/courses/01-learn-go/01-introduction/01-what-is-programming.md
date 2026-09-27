---
title: What Is Programming?
quiz:
  - question: What is a program?
    options:
      - text: A list of precise instructions that a computer follows, in order
        correct: true
      - text: A website that you visit in a browser
      - text: A kind of computer chip that does maths very fast
      - text: A message you send to another programmer
    explanation: |
      A program is just a list of instructions. The computer runs them one
      after another, exactly as written. It never guesses what you meant, so
      the instructions have to be precise.
  - question: Which of these is one of the main reasons Go was created?
    options:
      - text: To make web pages look nicer
      - text: To replace every other programming language
      - text: To be simple to read and fast to compile and run, even in very large codebases
        correct: true
    explanation: |
      Go was designed at Google for big teams working on big programs. It keeps
      the language small and readable, compiles quickly, and produces fast
      programs. That simplicity is also what makes it a great first language.
---

Welcome to Textio! You've just joined a small startup that sends text messages (SMS) for its customers: appointment reminders, delivery updates, two-factor login codes, that sort of thing. The backend is written in Go, and by the end of this course you'll be able to read and write it.

But first, what *is* programming?

## A program is a list of instructions

A computer is extremely fast and extremely literal. It can do billions of tiny steps per second, but it has no idea what you *want*. It only does what you tell it.

A **program** is a list of instructions that tells the computer exactly what to do, step by step. **Programming** is the job of writing those instructions.

Here's what a tiny Go program looks like:

```go
package main

import "fmt"

func main() {
	fmt.Println("Sending message to Alice...")
	fmt.Println("Message sent!")
}
```

It prints:

```text
Sending message to Alice...
Message sent!
```

Don't worry about every word yet. For now, notice that the instructions run **from top to bottom**, one after another. The first line prints first and the second line prints second.

## Code is written for humans too

The computer only cares that your code is correct. But other people (including you, three months from now) have to *read* it. Good programmers write code that is easy to understand, not just code that works.

## Why Go?

There are hundreds of programming languages. Python, JavaScript, Rust, Java... so why learn Go?

- **It's small.** Go has very few keywords and features. You can hold most of the language in your head, which is perfect when you're starting out.
- **It's fast.** Go programs are compiled into machine code (more on that soon), so they run quickly.
- **It's practical.** Go powers real infrastructure: Docker, Kubernetes, and a huge number of backend servers at companies like Google, Uber and Cloudflare.
- **It's great at doing many things at once.** Textio needs to send thousands of messages at the same time, and Go was built for exactly that kind of work.
- **Its tools are built in.** Formatting, testing and dependency management all come with Go. No hunting for extra tools.

Go was created at Google in 2007 by Robert Griesemer, Rob Pike and Ken Thompson, and released publicly in 2009. Their goal was a language that stays readable and quick to compile even when a codebase has millions of lines and hundreds of engineers.

## How this course works

Each lesson teaches one idea with short examples. When you finish reading, answer the quiz on the side to check your understanding. Many questions will show you some code and ask what it prints, because *reading* code carefully is the most important programming skill of all.

Let's write some Go.
