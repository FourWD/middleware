# CLAUDE.md

## General Behavior

- Be concise and direct.
- Do not over-explain.
- Focus on producing working, clean code.

## Code Style

- Write clean, readable, and idiomatic code.
- Prefer simplicity over cleverness.
- Follow standard conventions of the language (e.g., Go, TypeScript).

## Comments Policy

- Write comments ONLY when necessary.
- Avoid obvious comments that restate the code.
- Use comments for:
  - Complex logic
  - Non-obvious decisions
  - Edge cases

- All comments MUST be written in English.
- Keep comments short and precise.

## Restrictions

- Do NOT add unnecessary explanations outside the code unless explicitly requested.
- Do NOT generate verbose documentation.

## Example

### Bad

```go
// This function adds two numbers
func Add(a, b int) int {
    return a + b
}
```

### Good

```go
func Add(a, b int) int {
    return a + b
}
```

### Acceptable (when needed)

```go
func CalculateDiscount(price float64, userType string) float64 {
    // Apply special rate for premium users due to business rule v2
    if userType == "premium" {
        return price * 0.8
    }
    return price
}
```
