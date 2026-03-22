# Workflow

## 0. Learn from the past

Read .gemini/learnings.md to see what you've previously mistaken what you should do to improve.

## 1. Implement the code

Implement the code based on the plan. 

## 2. Run the linter

Spawn a subagent "lint-master" that lints the code with `just lint`.

See the error and fix the syntax errors.

## 3. Run the test

Spawn another subagent "code-tester" that runs `just test` to test the code.

## 4. Add changes and commit

Spawn another subagent "git-master" to do the following:
  Run `git add <files>` to add files that should be tracked by git. Run `git commit -m <commit_message>` to commit.

## 5. Push the changes

Run `git push` to push the changes to the current branch.

## 6. Review the workflow

Review what you've done. Note down what you've learned to .gemini/learnings.md. 

# MANDATORY: Use td for Task Management

You must run td usage --new-session at conversation start (or after /clear) to see current work.
Use td usage -q for subsequent reads.
