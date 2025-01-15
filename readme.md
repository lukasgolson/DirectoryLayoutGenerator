# Directory Layout Generator

`dirlayout` is a command-line tool for generating complex directory structures based on a simple, expressive DSL.

## Installation

Clone the repository and build the executable:

```bash
git clone https://github.com/lukasgolson/DirectoryLayoutGenerator.git
cd dirlayout
go build -o dirlayout main.go
```

Or run directly using:

```bash
go run main.go
```

## Usage

```bash
dirlayout --layout "<DSL_STRING>" --output "<OUTPUT_DIRECTORY>"
```

### Required Flags

- `--layout` (`-l`): A string describing the directory structure in DSL format.
- `--output` (`-o`): (Optional) The base directory where the layout will be created. Defaults to the current directory (`.`).

### Example

```bash
dirlayout --layout "site:3 > tree:5 > branch:North,South" --output "/tmp"
```

This creates the following structure in `/tmp`:

```
/tmp/site 1/tree 1/branch North/
/tmp/site 1/tree 1/branch South/
/tmp/site 1/tree 2/branch North/
/tmp/site 1/tree 2/branch South/
...
/tmp/site 3/tree 5/branch South/
```

---

## DSL Language Primer

The Directory Structure Language (DSL) allows you to define nested folder layouts concisely.

### Syntax

- **Basic Structure**: `name[:count]`
  - `name`: Base name for directories.
  - `count` (optional): Number of directories to create with sequential numbering.
    - Example: `site:3` creates `site 1`, `site 2`, `site 3`.

- **Nesting**: Use `>` to define subdirectories.
  - Example: `site:3 > tree:5` creates 3 `site` directories, each with 5 `tree` subdirectories.

- **Static Lists**: Use `:` with comma-separated names for explicit subdirectories.
  - Example: `branch:North,South` creates `branch North` and `branch South`.

- **Combining Features**: Combine counts and static lists for complex hierarchies.
  - Example: `site:3 > tree:5 > branch:North,South` creates:
    ```
    site 1/tree 1/branch North
    site 1/tree 1/branch South
    site 1/tree 2/branch North
    ...
    ```

### Examples

1. **Simple Layout**:
   ```bash
   dirlayout --layout "project:2" --output "/projects"
   ```
   Creates:
   ```
   /projects/project 1/
   /projects/project 2/
   ```

2. **Nested Layout**:
   ```bash
   dirlayout --layout "department:HR,Finance > team:3" --output "/company"
   ```
   Creates:
   ```
   /company/HR/team 1/
   /company/HR/team 2/
   /company/HR/team 3/
   /company/Finance/team 1/
   /company/Finance/team 2/
   /company/Finance/team 3/
   ```

3. **Multi-Level Nesting**:
   ```bash
   dirlayout --layout "site:3 > tree:4 > branch:North,South,East" --output "/tmp"
   ```
   Creates:
   ```
   /tmp/site 1/tree 1/branch North/
   /tmp/site 1/tree 1/branch South/
   ...
   /tmp/site 3/tree 4/branch East/
   ```

---

## Features

- **Recursive Nesting**: Define deeply nested structures.
- **Dynamic Counts**: Generate numbered directories automatically.
- **Explicit Lists**: Create specific folder names.
- **Flexible Base Path**: Control where the directories are generated.

---

## Troubleshooting

- **Error: No layout string provided**  
  Ensure you specify the `--layout` flag with a valid DSL string.

- **Error: Invalid count**  
  Verify that counts (e.g., `site:5`) are numeric values.

- **No Directories Created**  
  Check the specified `--output` path and ensure it exists or is writable.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE.txt) file for details.
```