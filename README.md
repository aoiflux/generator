# File System Artifacts Generator (fsagen)

Deterministic generator for creating diverse file-system artifacts to test forensic tools (Autopsy, EnCase, FTK, etc.). Use YAML manifests for simple operations, playbooks for complex modus operandi simulation, or the bulk generator for a quick synthetic corpus with no YAML required.

## Features

- Deterministic output with `--seed`
- Manifest mode: simple, declarative file operations (create/update/append/delete/mace/rename/truncate/rotate/ads/motw)
- Playbook mode: complex timelines with actors, timed steps, and templating
- Bulk mode: super-simple synthetic corpus generation with `--bulk` and `--depth`
- Extensive file type support: documents, logs, archives, media, emails, Windows artifacts
- MACE (atime/mtime) timestamp control
- Windows-specific: NTFS ADS and Mark-of-the-Web (MoTW)

## Install

Build from source:

```pwsh
go build -v -o fsagen.exe
```

## Usage

fsagen can generate artifacts using a manifest, a playbook, or the bulk generator:

```pwsh
fsagen [OPTIONS] <output-path>
```

**Options:**
- `--seed N` - PRNG seed for deterministic generation (default: 1)
- `--manifest FILE` - Execute a YAML manifest (simple file operations)
- `--playbook FILE` - Execute a YAML playbook (complex modus operandi)
- `--bulk N` - Super-simple bulk generation: N items per level (no YAML)
- `--depth D` - Bulk generation depth (default: 1)
- `--timeline FILE` - Generate forensic timeline after execution (formats: csv, txt, bodyfile, macb)

**Examples:**

Simple bulk generation with manifest:
```pwsh
fsagen --seed 42 --manifest examples/manifest-bulk-simple.yaml ./output
```

Complex adversary simulation with playbook:
```pwsh
fsagen --seed 100 --playbook examples/playbook-adversary-data-theft.yaml ./crime-scene
```

Generate artifacts with forensic timeline:
```pwsh
fsagen --seed 42 --playbook examples/playbook-comprehensive-ransomware.yaml --timeline timeline.csv ./output
```

Quick synthetic corpus with bulk generator (no YAML):
```pwsh
fsagen --seed 7 --bulk 3 --depth 2 ./quick-bulk
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

- **start**: RFC3339 or "now"
- **variables**: Global variables for templating (map of key-value pairs)
- **actors**: List of { name, base, variables }
	- name: Actor identifier
	- base: Base directory for this actor's files
	- variables: Actor-specific variables (override global variables)
- **steps**: Timeline steps
	- actor: Actor name
	- offset: time.Duration from start for first occurrence (e.g., 5m, 2h)
	- every: Repeat interval (optional)
	- repeat: Number of occurrences (default 1)
	- condition: Step-level conditional execution ("odd", "even", "first", "last")
	- batch_count: Generate N files in this step (multiplies actions)
	- actions: List of operations with extras:
		- offset: time.Duration relative to the step occurrence
		- condition: Action-level conditional execution
		- template: Predefined content template ("email", "log", "script", "doc")
		- All standard manifest fields (action, path, content, etc.)

Operations supported: `create|update|append|truncate|rotate|delete|mace|rename|ads|motw` (all operations work in both manifest and playbook). `ads` and `motw` are Windows-only. Timestamps are computed from the timeline unless explicitly provided in the action.

**Playbook templating:**
- `${SEQ}` - Monotonic sequence counter
- `${RND:N}` or `${RANDOM:N}` - Deterministic random string of length N
- `${DATE:layout}` - Current time formatted with Go layout (e.g., `${DATE:2006-01-02T15:04:05Z07:00}`)
- `${ACTOR}` - Current actor name
- `${VAR:name}` - Variable substitution (from global or actor-specific variables)
- `${UUID}` - Deterministic UUID based on sequence
- `${IP}` - Deterministic IP address (192.168.x.x range)
- `${HASH:N}` - Deterministic hash-like hex string of length N
- `${BATCH}` - Current batch index (when using batch_count)
- `${ITER}` - Current iteration index (when using repeat)

**Advanced Playbook Features:**

1. **Variables**: Define reusable values at global and actor scope
```yaml
variables:
  campaign_id: "OP-2024-001"
  target_org: "ACME Corp"
actors:
  - name: attacker
    base: users/victim/Downloads
    variables:
      ip_addr: "192.0.2.42"
```

2. **Conditional Execution**: Control when steps/actions run
```yaml
steps:
  - actor: malware
    repeat: 10
    condition: even  # Only runs on even iterations (0, 2, 4, ...)
    actions:
      - action: create
        path: file-${ITER}.txt
        condition: odd  # Further filtering at action level
```

3. **Batch Operations**: Generate multiple files in one step
```yaml
steps:
  - actor: ransomware
    batch_count: 100  # Creates 100 files
    actions:
      - action: create
        path: encrypted-${BATCH}.locked
        content_len: 2048
```

4. **Content Templates**: Use predefined realistic content
```yaml
actions:
  - action: create
    path: message.eml
    template: email  # Generates realistic email structure
```

Available templates: `email`, `log`, `script`, `doc`

**Example playbooks:**

- `examples/playbook-basic.yaml` - Simple two-actor workflow
- `examples/playbook-adversary-data-theft.yaml` - Stages documents, archives, writes exfil logs, backdates
- `examples/playbook-log-tampering.yaml` - Creates baseline logs, injects tampered entries, backdates, deletes
- `examples/playbook-persistence-artifacts.yaml` - Drops startup-like files and .reg exports
- `examples/playbook-email-and-archive.yaml` - Creates emails/images, archives, deletes originals
- `examples/playbook-log-rotate-and-truncate.yaml` - Demonstrates log rotation and truncation
- `examples/playbook-windows-ads-motw.yaml` - Adds NTFS ADS and Mark-of-the-Web (Windows-only)
- `examples/playbook-comprehensive-ransomware.yaml` - **Advanced**: Full ransomware attack with variables, batching, conditions, and templates
- `examples/playbook-insider-threat-exfil.yaml` - **Advanced**: 7-day insider threat scenario with repeated access patterns
- `examples/playbook-malware-lifecycle.yaml` - **Advanced**: 48-hour malware infection lifecycle with beaconing and anti-forensics

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

## Forensic Timeline Generation

After generating artifacts, fsagen can automatically create forensic timelines for analysis:

```pwsh
fsagen --playbook scenario.yaml --timeline output.csv ./artifacts
```

**Timeline Formats:**

- **CSV** (`.csv`): Structured data with all metadata (path, size, mode, timestamps, MD5, type, ADS)
- **TXT** (`.txt`): Human-readable format with detailed file information
- **Bodyfile** (`.bodyfile`): Compatible with The Sleuth Kit's mactime tool
- **MACB** (`.macb`): Modified/Accessed/Changed/Birth timeline showing all timestamp events separately

**Timeline Features:**

- MD5 hash calculation for all files (except files > 100MB)
- Full timestamp capture (access, modify, change/create times)
- NTFS Alternate Data Stream detection (Windows)
- Chronologically sorted by modification time
- Deterministic output (same seed = same timeline)

**Example workflows:**

```pwsh
# Generate ransomware scenario with CSV timeline
fsagen --seed 999 --playbook examples/playbook-comprehensive-ransomware.yaml --timeline ransomware.csv ./scene

# Create timeline compatible with mactime
fsagen --playbook examples/playbook-malware-lifecycle.yaml --timeline evidence.bodyfile ./analysis
mactime -b evidence.bodyfile -d > detailed-timeline.txt

# Generate MACB timeline for temporal analysis
fsagen --playbook examples/playbook-insider-threat-exfil.yaml --timeline investigation.macb ./case
```

See `examples/TIMELINE_EXAMPLES.md` for more timeline generation examples.

## Bulk Generation (no YAML)

Use the bulk generator when you need a fast, synthetic corpus without describing a scenario:

```pwsh
# N items per level, depth D directories deep
fsagen --seed 7 --bulk 3 --depth 2 ./quick-bulk

# You can also emit a timeline for bulk output
fsagen --seed 7 --bulk 3 --depth 2 --timeline timeline.csv ./quick-bulk
```

What it does:
- Creates a directory fan-out up to `--depth` with `--bulk` sub-branches per level
- Populates each level with many file types (txt, docx, png, pdf, mp4, csv, json, xml, html, log, reg, zip, exe, jsonl, syslog, md, eml, mbox)
- Deterministic names and contents from `--seed`

Intended use:
- Quickly produce a sizeable, diverse dataset for tool demos, performance tests, or classroom exercises
- Warm-up data for timeline/report pipelines when a complex MO isn’t needed

Notes:
- Output size grows quickly with `--bulk` and `--depth`. Start small (e.g., `--bulk 2 --depth 1` or `--bulk 3 --depth 2`).
- Bulk mode is structure/content focused; if you need precise timelines, actors, or conditions, prefer Playbooks.
