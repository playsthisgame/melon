## 1. Update Confirmation Prompt

- [x] 1.1 In `internal/cli/search_cmd.go` `offerAddMany`, change the prompt string from `[y/N]` to `[Y/n]`
- [x] 1.2 Update the input-acceptance condition to also accept empty string as a "yes" (i.e. `input == "" || input == "y" || input == "yes"`)

## 2. Tests

- [x] 2.1 Add a test for `offerAddMany` that verifies empty input (Enter) proceeds with installation
- [x] 2.2 Add a test that verifies `n` input cancels without installing
