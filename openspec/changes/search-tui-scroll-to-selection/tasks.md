## 1. Dynamic Height in searchModel

- [x] 1.1 Add a `const listReservedRows = 3` in `search_model.go` to account for title, hint, and padding rows
- [x] 1.2 In `newSearchModel`, keep the existing static fallback height (`min(results*2, 20)`) for initial list construction
- [x] 1.3 Handle `tea.WindowSizeMsg` in `searchModel.Update`: set list height to `msg.Height - listReservedRows`, clamped to `[2, len(items)*2]`

## 2. Dynamic Height in removeModel

- [x] 2.1 Add a `const listReservedRows = 3` in `remove_model.go` (or share from search_model.go if in the same package)
- [x] 2.2 In `newRemoveModel`, keep the existing static fallback height (`min(skills, 20)`) for initial list construction
- [x] 2.3 Handle `tea.WindowSizeMsg` in `removeModel.Update`: set list height to `msg.Height - listReservedRows`, clamped to `[2, len(items)]`

## 3. Tests

- [x] 3.1 Add a test that sends a `tea.WindowSizeMsg` to `searchModel.Update` with a small height and verifies the list height is clamped correctly
- [x] 3.2 Add a test that sends a `tea.WindowSizeMsg` to `searchModel.Update` with a large height and verifies the list height does not exceed `len(items)*2`
- [x] 3.3 Add a test that sends a `tea.WindowSizeMsg` to `removeModel.Update` with a small height and verifies the list height is clamped correctly
- [x] 3.4 Add a test that sends a `tea.WindowSizeMsg` to `removeModel.Update` with a large height and verifies the list height does not exceed `len(items)`
