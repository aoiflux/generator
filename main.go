package main

import (
	"flag"
	"fmt"
	manifestpkg "generator/manifest"
	playbookpkg "generator/playbook"
	timelinepkg "generator/timeline"
	"generator/util"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	seed := flag.Int64("seed", 1, "PRNG seed for deterministic generation")
	manifestPath := flag.String("manifest", "", "Path to a YAML manifest that defines specific artifacts and actions")
	playbookPath := flag.String("playbook", "", "Path to a YAML playbook that describes a modus operandi (high-level timeline)")
	timelineOutput := flag.String("timeline", "", "Generate forensic timeline after execution (formats: csv, txt, bodyfile, macb)")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		return
	}

	// Initialize deterministic RNG
	util.Seed(*seed)

	path := args[0]
	absPath, err := getRootPath(path)
	handle(err)

	// Require either manifest or playbook
	if *playbookPath == "" && *manifestPath == "" {
		fmt.Println("Error: either --manifest or --playbook is required")
		printUsage()
		os.Exit(1)
	}

	fmt.Println("Generating artifacts...")

	if *playbookPath != "" {
		if err := playbookpkg.ExecutePlaybook(absPath, *playbookPath); err != nil {
			handle(err)
		}
	} else if *manifestPath != "" {
		if err := manifestpkg.ExecuteManifest(absPath, *manifestPath); err != nil {
			handle(err)
		}
	}

	fmt.Println("Done!")

	// Generate timeline if requested
	if *timelineOutput != "" {
		fmt.Println("\nGenerating forensic timeline...")
		if err := generateTimeline(absPath, *timelineOutput); err != nil {
			fmt.Printf("Warning: timeline generation failed: %v\n", err)
		} else {
			fmt.Printf("Timeline written to: %s\n", *timelineOutput)
		}
	}
}

func printUsage() {
	fmt.Println("Usage: fsagen [OPTIONS] <output-path>")
	fmt.Println()
	fmt.Println("Generate deterministic filesystem artifacts for forensic testing.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --seed N           PRNG seed for deterministic generation (default: 1)")
	fmt.Println("  --manifest FILE    Execute a YAML manifest (simple file operations)")
	fmt.Println("  --playbook FILE    Execute a YAML playbook (complex modus operandi)")
	fmt.Println("  --timeline FILE    Generate forensic timeline after execution")
	fmt.Println()
	fmt.Println("Timeline formats (detected by file extension):")
	fmt.Println("  .csv               CSV format with full metadata")
	fmt.Println("  .txt               Human-readable text format")
	fmt.Println("  .bodyfile          Bodyfile format (compatible with mactime)")
	fmt.Println("  .macb              MACB timeline format")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  fsagen --seed 42 --manifest basic.yaml ./output")
	fmt.Println("  fsagen --seed 100 --playbook adversary.yaml ./crime-scene")
	fmt.Println("  fsagen --seed 100 --playbook adversary.yaml --timeline timeline.csv ./output")
	fmt.Println()
	fmt.Println("Either --manifest or --playbook is required.")
}

func generateTimeline(root string, outputPath string) error {
	// Generate the timeline
	tl, err := timelinepkg.Generate(root)
	if err != nil {
		return fmt.Errorf("generate timeline: %w", err)
	}

	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create timeline file: %w", err)
	}
	defer f.Close()

	// Determine format from file extension
	ext := strings.ToLower(filepath.Ext(outputPath))

	switch ext {
	case ".csv":
		return tl.WriteCSV(f)
	case ".txt":
		return tl.WriteTXT(f)
	case ".bodyfile":
		return tl.WriteBodyfile(f)
	case ".macb":
		return tl.WriteMACB(f)
	default:
		// Default to CSV if extension not recognized
		return tl.WriteCSV(f)
	}
}

func getRootPath(path string) (string, error) {
	finfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.Mkdir(path, fs.ModePerm)
		if err != nil {
			return "", err
		}
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	if finfo != nil {
		if !finfo.IsDir() {
			return filepath.Dir(absPath), nil
		}
	}

	return absPath, nil
}

func handle(err error) {
	if err != nil {
		fmt.Printf("\n\n%v\n\n", err)
		os.Exit(1)
	}
}
