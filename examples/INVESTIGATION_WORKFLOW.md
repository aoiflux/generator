# Complete Forensic Investigation Scenario

This example demonstrates the full workflow: generating a sophisticated crime scene and analyzing it with timeline tools.

## Scenario: APT Intrusion with Data Exfiltration

### Step 1: Generate the Crime Scene

```bash
fsagen --seed 12345 \
  --playbook examples/playbook-malware-lifecycle.yaml \
  --timeline apt-timeline.csv \
  ./apt-investigation
```

This creates:
- 48-hour malware infection timeline
- Initial dropper with MoTW
- Persistence mechanisms
- Credential dumps
- Lateral movement artifacts
- Data staging and exfiltration
- Anti-forensics activities
- Timeline in CSV format

### Step 2: Generate Multiple Timeline Formats

```bash
# CSV for spreadsheet analysis
fsagen --seed 12345 --playbook examples/playbook-malware-lifecycle.yaml \
  --timeline apt-timeline.csv ./apt-investigation

# Human-readable report
fsagen --seed 12345 --playbook examples/playbook-malware-lifecycle.yaml \
  --timeline apt-report.txt ./apt-investigation

# Bodyfile for mactime
fsagen --seed 12345 --playbook examples/playbook-malware-lifecycle.yaml \
  --timeline apt-evidence.bodyfile ./apt-investigation

# MACB timeline for temporal analysis
fsagen --seed 12345 --playbook examples/playbook-malware-lifecycle.yaml \
  --timeline apt-macb.macb ./apt-investigation
```

### Step 3: Analyze with The Sleuth Kit

```bash
# Convert bodyfile to human-readable timeline
mactime -b apt-evidence.bodyfile -d > apt-mactime.txt

# Filter to specific time window (first 2 hours of infection)
mactime -b apt-evidence.bodyfile -d 2024-09-01 2024-09-01-02:00:00 > apt-initial-infection.txt
```

### Step 4: Import to Forensic Tools

The generated timeline files can be imported into:

- **Autopsy**: Import CSV as timeline data source
- **EnCase**: Use bodyfile format with timeline module
- **FTK**: Import CSV for timeline analysis
- **log2timeline/plaso**: Process bodyfile format
- **Timesketch**: Import CSV for collaborative investigation

### Step 5: Analyze Patterns

Look for forensic indicators in the timeline:

1. **Initial Compromise** (T+0):
   - Dropper execution with Zone.Identifier ADS
   - Unusual file creation in Temp directory

2. **Persistence** (T+2m):
   - Startup folder modifications
   - Registry export artifacts

3. **Credential Harvesting** (T+5m):
   - Multiple creds-*.txt files created
   - Suspicious file creation patterns

4. **Lateral Movement** (T+10m-20m):
   - SMB scan artifacts in Windows/Temp
   - Network enumeration files

5. **Data Staging** (T+45m):
   - Large number of .dat files in staging directory
   - Encrypted chunks with UUIDs

6. **Exfiltration** (T+1h):
   - Large ZIP file creation
   - C2 communication logs

7. **Anti-Forensics** (T+2h):
   - Log truncation
   - File deletions
   - MACE timestamp manipulation

8. **Persistence Verification** (T+24h):
   - Health check files
   - Beacon log updates

9. **Ransomware Deployment** (T+48h):
   - Mass .locked file creation
   - Ransom note appearance

## Expected Timeline Output (CSV Sample)

```csv
Path,Size,Mode,Accessed,Modified,Changed/Created,MD5,Type,ADS
users/alice/AppData/Local/Temp/abc12345.exe,4096,-rwxrwxrwx,2024-09-01T00:00:00Z,2024-09-01T00:00:00Z,2024-09-01T00:00:00Z,5d41402abc4b2a76b9719d911017c592,file,Zone.Identifier
users/alice/AppData/Local/Temp/.beacon.log,15600,-rw-rw-rw-,2024-09-01T01:40:00Z,2024-09-01T01:40:00Z,2024-09-01T00:00:30Z,098f6bcd4621d373cade4e832627b4f6,file,
users/alice/AppData/Roaming/Microsoft/Windows/Start Menu/Programs/Startup/WindowsUpdate.lnk,256,-rw-rw-rw-,2024-09-01T00:02:00Z,2024-08-15T10:00:00Z,2024-09-01T00:02:00Z,ad0234829205b9033196ba818f7a872b,file,
...
```

## Expected MACB Output (Sample)

```
MACB Timeline for: ./apt-investigation
Generated: 2024-10-21T12:00:00Z
========================================================================================================================
Timestamp            Type   Path                                                         Size MD5
------------------------------------------------------------------------------------------------------------------------
2024-09-01 00:00:00  M...   users/alice/AppData/Local/Temp/abc12345.exe                    4096 5d41402abc4b2a76b9719d911017c592
2024-09-01 00:00:00  .A..   users/alice/AppData/Local/Temp/abc12345.exe                    4096 5d41402abc4b2a76b9719d911017c592
2024-09-01 00:00:30  M...   users/alice/AppData/Local/Temp/.beacon.log                      156 098f6bcd4621d373cade4e832627b4f6
2024-09-01 00:01:00  M...   users/alice/AppData/Local/Temp/.beacon.log                      312 098f6bcd4621d373cade4e832627b4f6
...
```

## Deterministic Verification

Generate the same scene twice and compare:

```bash
# First generation
fsagen --seed 12345 --playbook examples/playbook-malware-lifecycle.yaml --timeline timeline1.csv ./scene1

# Second generation with same seed
fsagen --seed 12345 --playbook examples/playbook-malware-lifecycle.yaml --timeline timeline2.csv ./scene2

# Compare timelines (should be identical)
diff timeline1.csv timeline2.csv
```

No differences = perfect determinism achieved!

## Use Cases

This workflow is perfect for:

- **Training**: Create realistic scenarios for forensic analysts
- **Tool Testing**: Verify forensic tool capabilities with known datasets
- **Research**: Generate controlled datasets for timeline analysis research
- **Demonstrations**: Show temporal attack patterns in presentations
- **Validation**: Verify forensic tool timeline accuracy
- **Education**: Teach timeline analysis with reproducible examples
