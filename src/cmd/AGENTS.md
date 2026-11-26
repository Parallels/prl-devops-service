# cmd

CLI entry points dispatched from `main.go`: `api` (run the service), `catalog` (push/pull/list from the CLI), `install`/`uninstall` (launchd/systemd service), `reverse-proxy`, `register_with_orchestrator`, `init_orchestrator_client`, `generate_security_key`, `root_password`, `test_providers`, `version`, `help`.

- `main.go` here parses `os.Args` and routes to the command files; add a command by adding a file and a case there, plus a line in `help.go`.
- Commands are thin: parse flags, build a `basecontext`, call into `startup`, `catalog`, `install`, etc. Keep logic out of this package.
- Install/service-unit code lives in `../install/`, not here.
- Docs: `docs/docs/devops/getting-started/` (installation, configuration).
