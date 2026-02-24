## Mass Project Updater

Mass Project Updater is a small Go CLI that automates dependency updates across many GitLab UI projects and opens merge requests for you.

It:
- Finds projects in a GitLab group whose names start with a given prefix (default `ui-spa`).
- Applies version changes from `deps.json` to each project's `package.json`.
- Runs `npm install` and `npm build`, commits the changes, and opens an MR back to `develop`.

### Requirements

- Go installed (for building).
- `npm` available on your PATH.
- A GitLab access token with permission to read/write the target projects.

### Configuration

You can configure the GitLab connection in two ways:

- **Environment variables** (non‑interactive / CI):
  - `GITLAB_BASE_URL` – e.g. `https://gitlab.example.com`
  - `GITLAB_TOKEN` – personal access token
  - `GITLAB_GROUP_ID` – numeric or path group ID

- **Interactive prompts** (double‑click or run with no args):
  - If any of the above are missing, the CLI will ask you to enter them.
  - You will also be asked whether to run in dry‑run mode and to confirm the list of projects before proceeding.

### Building

From the project root:

```bash
go build -o ./build/mass-project-updater ./cmd/updater
```

On Windows the binary will be `build/mass-project-updater.exe`.

### Usage

Interactive (recommended when trying it out – shows a menu and prompts):

```bash
./build/mass-project-updater   # or double‑click the .exe on Windows
```

Non‑interactive (for scripts/CI):

```bash
export GITLAB_BASE_URL="https://gitlab.example.com"
export GITLAB_TOKEN="your-token"
export GITLAB_GROUP_ID="1234"

mass-project-updater update-deps \
  --deps-file deps.json \
  --project-prefix ui-spa \
  --max-workers 5 \
  --dry-run
```

