# mappers

Pure conversion functions between API DTOs (`../models`) and persistence models (`../data/models`), e.g. `MapDtoVirtualMachineFromApi` / `MapDtoVirtualMachineToApi`. One file per resource.

- No I/O, no DB, no config reads. Keep them deterministic and unit-testable.
- When a field is added to a model on either side, update both directions here and the test; a silently dropped field is the usual bug.
- Computed/effective values (e.g. `effective_claims.go`) are the only place logic belongs, and it needs a test.
