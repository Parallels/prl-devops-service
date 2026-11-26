# pdfile

Parser and executor for PDFiles, the Dockerfile-like format for pulling, pushing, importing and cloning VMs.

- `main.go` — `PDFileService` with an ordered list of command processors. `load.go`/`process.go` parse lines, `validation.go` and `diagnostics/` report errors with line info, `pull.go`/`push.go`/`import*.go`/`list.go` execute.
- `processors/` — one processor per directive (`FROM`, `TO`, `AUTHENTICATE`, `PROVIDER`, `EXECUTE`, `CLAIM`, `ROLE`, `TAG`, ...), implementing `PDFileCommandProcessor` from `command_processor.go`. Add a directive: new processor file, register it in `NewPDFileService`, update validation, add tests.
- `models/` — the parsed `PDFile` structure.
- Processor order matters (the empty/comment processor first). Credentials in a PDFile (`AUTHENTICATE`, `EXECUTE PASSWORD`) are secrets: never log them (see `execute_credentials_test.go`).
- Docs: `docs/docs/devops/catalog/pdfile.md` — update it when directives change.
