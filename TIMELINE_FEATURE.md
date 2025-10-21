# Timeline Generation Feature - Summary

## Overview

Added comprehensive forensic timeline generation capability to fsagen. After generating artifacts with manifests or playbooks, users can automatically create timeline files in multiple forensic formats.

## New Components

### 1. Timeline Package (`timeline/`)
- **timeline.go**: Core timeline generation logic
- **timeline_windows.go**: Windows-specific timestamp extraction
- **timeline_unix.go**: Unix/Linux timestamp extraction

### 2. Timeline Formats

#### CSV Format (`.csv`)
Structured spreadsheet format with columns:
- Path, Size, Mode, Accessed, Modified, Changed/Created, MD5, Type, ADS

#### Text Format (`.txt`)
Human-readable report with:
- Header with generation info
- Detailed file entries with all metadata
- Indented details for readability

#### Bodyfile Format (`.bodyfile`)
Compatible with The Sleuth Kit's mactime:
- Format: `MD5|name|inode|mode|UID|GID|size|atime|mtime|ctime|crtime`
- Can be processed with: `mactime -b file.bodyfile -d`

#### MACB Format (`.macb`)
Modified/Accessed/Changed/Birth timeline:
- Separates each timestamp type (M..., .A.., ..C.)
- Shows temporal patterns clearly
- Sorted chronologically

### 3. Features

✅ **Complete Timestamp Capture**
- Access time (atime)
- Modification time (mtime)
- Change/Creation time (ctime) - platform-specific
- Proper Windows creation time vs Unix change time

✅ **MD5 Hash Calculation**
- Automatic for all files
- Skips files > 100MB for performance
- Useful for integrity verification

✅ **NTFS ADS Detection** (Windows)
- Detects alternate data streams
- Lists stream names in CSV/TXT output
- Identifies Zone.Identifier, metadata streams

✅ **Platform-Aware**
- Uses syscall.Win32FileAttributeData on Windows
- Uses syscall.Stat_t on Unix/Linux
- Handles platform differences in timestamp semantics

✅ **Sorted Output**
- Chronologically ordered by modification time
- MACB format shows all timestamp events separately
- Easy to identify temporal sequences

✅ **Deterministic**
- Same seed produces identical timelines
- Reproducible for validation
- Perfect for testing and verification

## Usage

### Basic Usage
```bash
fsagen --playbook scenario.yaml --timeline output.csv ./artifacts
```

### Format Selection (by file extension)
```bash
# CSV format
fsagen --playbook attack.yaml --timeline timeline.csv ./scene

# Text format
fsagen --playbook attack.yaml --timeline report.txt ./scene

# Bodyfile format
fsagen --playbook attack.yaml --timeline evidence.bodyfile ./scene

# MACB format
fsagen --playbook attack.yaml --timeline macb-timeline.macb ./scene
```

### Integration with Forensic Tools

#### The Sleuth Kit
```bash
fsagen --playbook malware.yaml --timeline evidence.bodyfile ./case
mactime -b evidence.bodyfile -d > timeline.txt
```

#### Timeline Analysis
1. Generate scene with timeline
2. Import CSV to Autopsy/EnCase/FTK
3. Analyze temporal patterns
4. Identify attack phases

## Implementation Details

### File Walking
- Uses `filepath.Walk()` for complete directory traversal
- Gracefully handles permission errors
- Includes files and directories

### Timestamp Extraction
- Platform-specific build tags (`// +build windows`, `// +build !windows`)
- Direct syscall access for accurate timestamps
- Fallback to `ModTime()` if syscalls fail

### MD5 Calculation
- Streaming hash computation with `io.Copy()`
- Memory-efficient for large files
- Size limit (100MB) to avoid performance impact

### ADS Detection (Windows)
- Checks common stream names (Zone.Identifier, metadata, content)
- Uses `os.Stat()` on stream path syntax (path:stream)
- Returns list of found streams

## Code Structure

```
timeline/
├── timeline.go          # Core logic, format writers
├── timeline_windows.go  # Windows syscall timestamp extraction
└── timeline_unix.go     # Unix syscall timestamp extraction

main.go                  # Added --timeline flag and generateTimeline()

examples/
├── TIMELINE_EXAMPLES.md      # Timeline usage examples
└── INVESTIGATION_WORKFLOW.md # Complete forensic workflow
```

## Example Output

### CSV Sample
```csv
Path,Size,Mode,Accessed,Modified,Changed/Created,MD5,Type,ADS
users/victim/Downloads/malware.exe,4096,-rwxrwxrwx,2024-09-01T00:00:00Z,2024-09-01T00:00:00Z,2024-09-01T00:00:00Z,5d41402abc4b2a76b9719d911017c592,file,Zone.Identifier
users/victim/AppData/Local/Temp/staging/data.zip,262144,-rw-rw-rw-,2024-09-01T01:00:00Z,2024-09-01T01:00:00Z,2024-09-01T01:00:00Z,098f6bcd4621d373cade4e832627b4f6,file,metadata
```

### Text Sample
```
Forensic Timeline for: ./crime-scene
Generated: 2024-10-21T12:00:00Z
Total entries: 156
========================================================================================================================

2024-09-01 00:00:00 [FILE] users/victim/Downloads/malware.exe
         Size: 4096 bytes | Mode: -rwxrwxrwx | MD5: 5d41402abc4b2a76b9719d911017c592
         Access: 2024-09-01 00:00:00 | Modified: 2024-09-01 00:00:00 | Changed: 2024-09-01 00:00:00
         ADS: Zone.Identifier
```

### MACB Sample
```
Timestamp            Type   Path                                                         Size MD5
------------------------------------------------------------------------------------------------------------------------
2024-09-01 00:00:00  M...   users/victim/Downloads/malware.exe                           4096 5d41402abc4b2a76b9719d911017c592
2024-09-01 00:00:00  .A..   users/victim/Downloads/malware.exe                           4096 5d41402abc4b2a76b9719d911017c592
2024-09-01 00:00:30  M...   users/victim/AppData/Local/Temp/.beacon.log                   156 098f6bcd4621d373cade4e832627b4f6
```

## Testing

### Determinism Test
```bash
# Generate twice with same seed
fsagen --seed 42 --playbook test.yaml --timeline t1.csv ./out1
fsagen --seed 42 --playbook test.yaml --timeline t2.csv ./out2

# Compare (should be identical)
diff t1.csv t2.csv
```

### Format Test
```bash
# Generate all formats from same seed
fsagen --seed 100 --playbook scenario.yaml --timeline out.csv ./test
fsagen --seed 100 --playbook scenario.yaml --timeline out.txt ./test
fsagen --seed 100 --playbook scenario.yaml --timeline out.bodyfile ./test
fsagen --seed 100 --playbook scenario.yaml --timeline out.macb ./test

# All should contain same files/hashes with different formatting
```

## Benefits

1. **Integrated Workflow**: Generate and analyze in one command
2. **Format Flexibility**: Multiple formats for different tools
3. **Forensic Tool Compatibility**: Works with standard toolchains
4. **Verification**: Built-in checksums for integrity
5. **Temporal Analysis**: Clear visualization of time-based patterns
6. **Educational**: Perfect for teaching timeline analysis
7. **Reproducible**: Deterministic output for validation

## Future Enhancements (Potential)

- JSON timeline format for programmatic processing
- Super timeline format (log2timeline/plaso compatible)
- Timeline filtering options (date range, file types)
- Statistical summary (file counts, size distributions)
- Anomaly detection (unusual timestamp patterns)
- Visual timeline generation (HTML/SVG output)
- Integration with timeline visualization tools

## Documentation

- README.md: Updated with --timeline flag, formats, examples
- examples/TIMELINE_EXAMPLES.md: Timeline-specific usage examples
- examples/INVESTIGATION_WORKFLOW.md: Complete forensic investigation scenario

## Compatibility

- ✅ Windows (with NTFS ADS detection, creation time)
- ✅ Linux (with proper change time handling)
- ✅ macOS (with Unix timestamp extraction)
- ✅ The Sleuth Kit (bodyfile format)
- ✅ Standard forensic tools (CSV import)
