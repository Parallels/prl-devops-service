# models

API-facing DTOs: request/response bodies, error envelopes (`ApiErrorResponse`, `ApiCommonResponse`), events (`event_message.go`), orchestrator host/resource views, requirement checks.

- These are the public REST contract and appear in swagger/Postman. Renaming a JSON tag is a breaking API change; add fields instead and update the docs.
- Do not put persistence tags or DB logic here; persistence types live in `../data/models`. Convert with `../mappers`.
- Never add secret fields to response types (passwords, keys, tokens); there are shape tests guarding this elsewhere, keep to the same intent.
