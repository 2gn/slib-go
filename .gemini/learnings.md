# Learnings

## 2026-03-22

- `termshot` is a great tool for capturing CLI screenshots, but it might panic on certain ANSI escape sequences or special characters (like `---` as separators).
- Always ensure the target directory for screenshots exists before running `termshot`.
- When starting a session, always check the `td` focused issue to ensure you don't overwrite or conflict with ongoing work.
- `td review --minor` and `td approve` are useful for self-completing smaller tasks.
