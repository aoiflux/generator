# File System Artifacts Generator (fsagen)

Deterministic generator for creating diverse file-system artifacts to test forensic tools (Autopsy, EnCase, FTK, etc.). Uses YAML manifests for simple operations and playbooks for complex modus operandi simulation.

## Features

- Deterministic output with `--seed`
- Manifest mode: simple, declarative file operations (create/update/append/delete/mace/rename/truncate/rotate/ads/motw)
- Playbook mode: complex timelines with actors, timed steps, and templating
- Extensive file type support: documents, logs, archives, media, emails, Windows artifacts
- MACE (atime/mtime) timestamp control
- Windows-specific: NTFS ADS and Mark-of-the-Web (MoTW)

## Install

Build from source:

```pwsh
go build -v -o fsagen.exe
```

## Usage

fsagen requires either a manifest or playbook file to generate artifacts:

```pwsh
fsagen [OPTIONS] <output-path>
```

**Options:**
- `--seed N` - PRNG seed for deterministic generation (default: 1)
- `--manifest FILE` - Execute a YAML manifest (simple file operations)
- `--playbook FILE` - Execute a YAML playbook (complex modus operandi)

**Examples:**

Simple bulk generation with manifest:
```pwsh
fsagen --seed 42 --manifest examples/manifest-bulk-simple.yaml ./output
```

Complex adversary simulation with playbook:
```pwsh
fsagen --seed 100 --playbook examples/playbook-adversary-data-theft.yaml ./crime-scene
```

## Manifest schema

YAML with a sequence of operations:

- action: `create|update|append|truncate|rotate|delete|mace|rename|ads|motw`
- path: target path relative to output root
- type: `file|dir` (for create)
- ext: file extension to append if `path` has no extension
- content: literal content (optional)
- content_len: size of deterministic random content (fallback when content not provided)
- atime/mtime: RFC3339 timestamps for MACE control
- new_path: new location for `rename` or `rotate`
- stream: ADS stream name (for `ads` action, Windows-only)
- zone_id, host_url, referrer_url: for `motw` action (Windows-only)

**Examples:**
- `examples/manifest-basic.yaml` - Basic create/update/delete operations
- `examples/manifest-bulk-simple.yaml` - Quick bulk file generation across multiple types

## Playbook schema

YAML with a timeline and actors:

- start: RFC3339 or "now"
- actors: list of { name, base } to scope paths per actor
- steps:
	- actor: actor name
	- offset: time.Duration from start for first occurrence (e.g., 5m, 2h)
	- every: repeat interval (optional)
	- repeat: number of occurrences (default 1)
	- actions: list of operations like in manifest with extras:
		- offset: time.Duration relative to the step occurrence
		- templating in fields: `${SEQ}` (monotonic counter), `${RND:N}` for deterministic random strings of length N

Operations supported: `create|update|append|truncate|rotate|delete|mace|rename|ads|motw` (all operations work in both manifest and playbook). `ads` and `motw` are Windows-only. Timestamps are computed from the timeline unless explicitly provided in the action.

**Playbook templating:**
- `${SEQ}` - Monotonic sequence counter
- `${RND:N}` - Deterministic random string of length N
- `${DATE:layout}` - Current time formatted with Go layout
- `${ACTOR}` - Current actor name

**Example playbooks:**

- `examples/playbook-basic.yaml` - Simple two-actor workflow
- `examples/playbook-adversary-data-theft.yaml` - Stages documents, archives, writes exfil logs, backdates
- `examples/playbook-log-tampering.yaml` - Creates baseline logs, injects tampered entries, backdates, deletes
- `examples/playbook-persistence-artifacts.yaml` - Drops startup-like files and .reg exports
- `examples/playbook-email-and-archive.yaml` - Creates emails/images, archives, deletes originals
- `examples/playbook-log-rotate-and-truncate.yaml` - Demonstrates log rotation and truncation
- `examples/playbook-windows-ads-motw.yaml` - Adds NTFS ADS and Mark-of-the-Web (Windows-only)

## Notes on timestamps

- Sets mtime/atime via `os.Chtimes`. ctime is not directly settable on most systems and will reflect metadata change time.
- To emulate directory timestamp skew on deletion, `delete` can include `atime/mtime` which will be applied to the parent directory after removal.

## Reproducibility

- All random values (names, synthetic content) come from a seeded PRNG. Use the same `--seed` to reproduce identical output on the same platform and file system.
- Race-condition free: concurrent operations use proper synchronization while maintaining determinism.

## Supported File Types

The generator can create artifacts with proper structure for:

- **Documents**: .txt, .md, .docx, .pdf
- **Data**: .csv, .json, .jsonl, .xml, .html
- **Logs**: .log, .syslog, .jsonl
- **Media**: .png, .mp4
- **Archives**: .zip
- **Email**: .eml, .mbox
- **Windows**: .reg, .exe, NTFS ADS, MoTW
