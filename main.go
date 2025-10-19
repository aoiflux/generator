package main

import (
	"flag"
	"fmt"
	manifestpkg "generator/manifest"
	playbookpkg "generator/playbook"
	"generator/util"
	"io/fs"
	"os"
	"path/filepath"
)

func main() {
	seed := flag.Int64("seed", 1, "PRNG seed for deterministic generation")
	manifestPath := flag.String("manifest", "", "Path to a YAML manifest that defines specific artifacts and actions")
	playbookPath := flag.String("playbook", "", "Path to a YAML playbook that describes a modus operandi (high-level timeline)")
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
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  fsagen --seed 42 --manifest basic.yaml ./output")
	fmt.Println("  fsagen --seed 100 --playbook adversary.yaml ./crime-scene")
	fmt.Println()
	fmt.Println("Either --manifest or --playbook is required.")
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
