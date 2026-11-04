# Task 01 — Split `virtual_machine.go` (ER-3 refactor-first, pure move)

## Context
Feature plan §6 ER-3: any file ≥ 1,000 lines must be split **before** tagging.
`src/data/models/virtual_machine.go` is **1,018 lines** (re-measured this round) holding
`VirtualMachine` + 31 nested `VirtualMachine*` config types with their helper methods
interleaved. Tasks 02–04 add GORM tags on top of the result; this PR changes zero behaviour.

## Objective
`virtual_machine.go` is split into 4 same-package files by cohesive cluster;
`cd src && go build ./... && go test ./... -count=1` is green **with zero diff to any type,
field, tag, method body, or comment** — the diff is `git mv`-class moves only.

## Files
- Modify: `src/data/models/virtual_machine.go` (keeps `VirtualMachine` + its methods/clusters)
- Create: `src/data/models/virtual_machine_hardware.go` — clusters: `VirtualMachineGuestTools`,
  `VirtualMachineHardware`, `VirtualMachineCPU`, `VirtualMachineMemory`, `VirtualMachineMemoryQuota`,
  `VirtualMachineHdd0`, `VirtualMachineCdrom0`, `VirtualMachineNet0`, `VirtualMachineSound0`,
  `VirtualMachineVideo` (grep anchors: 328–591)
- Create: `src/data/models/virtual_machine_config.go` — clusters: `VirtualMachineAdvanced`(220),
  `Coherence`(252), `Expiration`(276), `Fullscreen`(284), `GuestSharedFolders`(312),
  `MiscellaneousSharing`(592), `Modality`(608), `MouseAndKeyboard`(632), `Optimization`(656),
  `PrintManagement`(704), `SMBIOSSettings`(724), `Security`(744), `SharedApplications`(784),
  `SmartMount`(816), `StartupAndShutdown`(840), `TimeSynchronization`(880), `TravelMode`(904),
  `USBAndBluetooth`(924)
- Create: `src/data/models/virtual_machine_network.go` — clusters: `NetworkInformation`(952),
  `NetworkInformationBound`(1000), `NetworkInformationIPAddress`(984)

## File size guard (ER-4; `wc -l` at plan time / projected after)
| File | Now | After | Verdict |
|---|---|---|---|
| `virtual_machine.go` | **1,018** | ~230 → ~270 after task 04 tags | refactor-first (this task) |
| `virtual_machine_hardware.go` | new | ~270 (proj.) | ok |
| `virtual_machine_config.go` | new | ~470 (proj.) | ok (<500 guideline, «1,000 cap) |
| `virtual_machine_network.go` | new | ~80 (proj.) | ok |

## Cluster move map (source ranges by `type` anchors; re-derive with `grep -n "^type \|^func "` before each move)
| Target file | Anchors (line → type) |
|---|---|
| `virtual_machine.go` (stays) | 5 `VirtualMachine` struct + its method cluster (…→219) |
| `virtual_machine_config.go` | 220 Advanced, 252 Coherence, 276 Expiration, 284 Fullscreen, 312 GuestSharedFolders, 592 MiscellaneousSharing, 608 Modality, 632 MouseAndKeyboard, 656 Optimization, 704 PrintManagement, 724 SMBIOSSettings, 744 Security, 784 SharedApplications, 816 SmartMount, 840 StartupAndShutdown, 880 TimeSynchronization, 904 TravelMode, 924 USBAndBluetooth |
| `virtual_machine_hardware.go` | 328 GuestTools, 344 Hardware, 388 CPU, 424 Cdrom0, 448 Hdd0, 480 Memory, 500 MemoryQuota, 508 Net0, 532 Sound0, 552 Video |
| `virtual_machine_network.go` | 952 NetworkInformation, 984 NetworkInformationIPAddress, 1000 NetworkInformationBound |

## Implementation spec
0. Baseline first: `cd src && go build ./... && go test ./... -count=1`, record results (reference
   for step 2), and note `git rev-parse HEAD` — the revert anchor.
1. Work top-to-bottom through the file; the layout alternates
   `type + fields → its helper funcs/methods`. Move each **contiguous cluster** (doc comment,
   `type` block, every function whose receiver or subject is that type) into the target file.
   Cluster boundaries are the `type` line numbers listed above (verify with
   `grep -n "^type \|^func " src/data/models/virtual_machine.go` — cheap, allowed).
2. Each new file starts: `package models` + only the imports its code needs
   (`sort`,`strings`,`fmt`,`time`, …; let `goimports`-equivalent trim decide).
3. Lines 1–56 (`VirtualMachine` struct) and 57–219 (main-VM methods, lines before `Advanced`@220)
   stay in `virtual_machine.go`.
4. **No edits**: no field/tag reorder, no comment rewording, no func rewrite. If the compiler
   complains, the cluster moved to the wrong file — move again, don't modify.
5. Optional guard-net: `git diff --stat` should show 4 files, zero insertions except `package`/
   `import` headers; `git diff -w` review should be pure relocation.

## Out of scope
- Any `gorm:` tag (task 04), any rename, any test changes, touching `virtual_machines.go`
  (the `ParallelsVM` DTO file — 261 lines, stays whole, never a table root per plan D-3 row 24).

## Validation
1. `cd src && go build ./...` → clean.
2. `cd src && go test ./... -count=1` → identical results to pre-split `main` baseline (run first).
3. `wc -l src/data/models/virtual_machine*.go` → all 4 files < 1,000 (ER-2) and ≈ projections.
4. `grep -c "^type VirtualMachine" src/data/models/virtual_machine*.go` summed = 32 (nothing lost).
5. `go vet ./data/models/...` clean (catches mis-filed receivers that still compile).
6. `git diff <baseline> -- src/data/models/` — every `+`/`-` line is a `package`/`import` header or
   blank-line boundary; any type/func/tag line appearing on both sides fails review.

## Rollback
Single revert of the split commit; no other commit in this PR (ER-3: split commit and feature
commit separate — here there is only the split commit).

## PBI acceptance mapping
Moves AC1 (enabler) and satisfies ER-3 pre-seeded trigger; AC4 unregressed (behaviour-identical
proof = step 2).
