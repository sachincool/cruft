package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sachincool/cruft/internal/tui"
)

// glossaryEntry is one term cruft uses, in plain language.
//
// Keep entries short. Anything longer than two short paragraphs belongs
// in `cruft explain <cleaner>` or the README, not here. The point of
// the glossary is to anchor terms a first-time user hits in doctor / TUI
// / flag help, without making them leave the tool to learn the tool.
type glossaryEntry struct {
	Name  string `json:"name"`
	Short string `json:"short"`
	Long  string `json:"long,omitempty"`
}

// Source of truth for vocabulary. If a term shows up anywhere in the UX
// (TUI, flag help, doctor, explain, errors) and isn't obvious from
// English, it belongs here.
var glossaryEntries = []glossaryEntry{
	{
		Name:  "audit log",
		Short: "JSONL receipt of every action cruft took (or would have taken).",
		Long:  "One line per finding written to ~/.local/share/cruft/runs/<run-id>.jsonl. Includes path, bytes, timestamp, dry-run flag, success/error. Use `cruft history` or `cruft last` to read; `jq` works fine for ad-hoc queries.",
	},
	{
		Name:  "budget",
		Short: "Stop reclaiming once this much disk is freed. Default: unlimited.",
		Long:  "Set with --budget (e.g. --budget 5GB). Useful when you only want a quick win and don't care which cleaner gets you there.",
	},
	{
		Name:  "busy processes",
		Short: "Process names that, if running, make a cleaner skip itself.",
		Long:  "Each cleaner declares its own list — e.g. npm skips if `node` is running, terraform skips if `terraform` is running. Prevents corrupting a cache mid-build. Detected with pgrep.",
	},
	{
		Name:  "cleaner",
		Short: "One unit of cleanup — one per tool (npm, brew, xcode-derived, …).",
		Long:  "Run `cruft list` for the full set. Each cleaner declares the paths it owns, what counts as busy, whether it's risky, and how to actually clean.",
	},
	{
		Name:  "dry run",
		Short: "Preview only — show what would happen, change nothing.",
		Long:  "Pass --dry-run to preview without deleting. Default OFF: cruft deletes on confirm. The preview is for caution or scripting (e.g. `cruft run --all --dry-run --json | jq` to see reclaim without acting). In the TUI press `d` to toggle preview mid-session.",
	},
	{
		Name:  "finding",
		Short: "One specific path a cleaner identified as reclaimable.",
		Long:  "A single cleaner can return many findings — e.g. terraform returns one per stale .terraform/ directory it found. Each finding carries its own bytes, age, and reason.",
	},
	{
		Name:  "in use",
		Short: "A relevant process is running; cleaner will skip this run.",
		Long:  "Shown in doctor as `in use (procname)`. Not an error — re-run when the process exits, or stop the process and re-run. Cruft never touches a cache while its owning tool is live.",
	},
	{
		Name:  "not installed",
		Short: "The tool this cleaner targets isn't on this machine.",
		Long:  "Hidden in the TUI; shown in doctor for transparency. Nothing to do.",
	},
	{
		Name:  "profile",
		Short: "Preset filter over cleaners: conservative | balanced | aggressive.",
		Long:  "conservative = only caches that regenerate in seconds. balanced (default) = everything safe. aggressive = also includes risky cleaners. Pass --profile to override.",
	},
	{
		Name:  "ready",
		Short: "Tool detected, idle, safe to clean.",
	},
	{
		Name:  "risky",
		Short: "Unchecked by default — may lose recoverable state or take long to regenerate.",
		Long:  "Each risky cleaner now states its specific reason inline (e.g. \"may corrupt VM if running\", \"forces multi-minute re-index\"). Run `cruft explain <name>` or hit ? in the TUI for the per-cleaner risk.",
	},
	{
		Name:  "stale days",
		Short: "How many idle days before a .terraform/ etc counts as cleanable. Default: 30.",
		Long:  "IaC cleaners (terraform, terragrunt, pulumi) only flag directories whose parent project hasn't been touched in this long. Tune with --stale-days. Active projects are never touched.",
	},
	{
		Name:  "safe mode",
		Short: "Route deletions through a 7-day recycle so `cruft restore` works.",
		Long:  "Off by default — most of what cruft cleans regenerates on next use (npm cache, Xcode DerivedData, brew downloads), so the recycle is theatre and wastes disk. Opt in with --safe (CLI) or press `s` in the TUI when you want a real undo window. Restorable with `cruft restore <run-id>` until the next sweep (default 7 days, configurable with --tombstone-days). Risky cleaners that shell out (homebrew, docker) can't be recycled — the upstream tool unlinks files directly.",
	},
	{
		Name:  "tombstone",
		Short: "Alias for 'safe mode' — the internal name for the recycle bin directory.",
		Long:  "Lives at ~/.local/share/cruft/tombstone/<run-id> when --safe is on. See `cruft glossary safe-mode`.",
	},
	{
		Name:  "whitelist",
		Short: "Each cleaner declares the exact paths it's allowed to touch.",
		Long:  "Every deletion runs through fsutil.SafeRemove, which resolves symlinks first and refuses any path that isn't under a registered prefix. You can't trick a cleaner into escaping its declared scope.",
	},
}

func cmdGlossary() *cobra.Command {
	return &cobra.Command{
		Use:   "glossary [term]",
		Short: "Define the vocabulary cruft uses (risky, tombstone, dry-run, …).",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			// Deterministic order; users skim alphabetically.
			sort.Slice(glossaryEntries, func(i, j int) bool {
				return glossaryEntries[i].Name < glossaryEntries[j].Name
			})

			if flagJSON {
				if len(args) == 1 {
					if e := lookupTerm(args[0]); e != nil {
						return json.NewEncoder(os.Stdout).Encode(e)
					}
					return fmt.Errorf("unknown term: %s", args[0])
				}
				return json.NewEncoder(os.Stdout).Encode(glossaryEntries)
			}

			if len(args) == 1 {
				e := lookupTerm(args[0])
				if e == nil {
					return fmt.Errorf("unknown term: %s (try `cruft glossary` for the full list)", args[0])
				}
				printGlossaryEntry(*e, true)
				return nil
			}

			fmt.Println(tui.StyleTitle.Render("cruft glossary"))
			fmt.Println(tui.StyleMuted.Render("Every term cruft uses, in plain language. `cruft glossary <term>` for full detail."))
			fmt.Println()
			for _, e := range glossaryEntries {
				printGlossaryEntry(e, false)
			}
			return nil
		},
	}
}

// lookupTerm finds an entry by case-insensitive prefix or exact match.
// Tolerates the dash/space split — "dry-run" and "dry run" both match.
func lookupTerm(needle string) *glossaryEntry {
	n := normalise(needle)
	for i := range glossaryEntries {
		if normalise(glossaryEntries[i].Name) == n {
			return &glossaryEntries[i]
		}
	}
	for i := range glossaryEntries {
		if strings.HasPrefix(normalise(glossaryEntries[i].Name), n) {
			return &glossaryEntries[i]
		}
	}
	return nil
}

func normalise(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	return s
}

func printGlossaryEntry(e glossaryEntry, includeLong bool) {
	fmt.Printf("%s\n", tui.StyleAccent.Render(e.Name))
	fmt.Printf("  %s\n", e.Short)
	if includeLong && e.Long != "" {
		fmt.Println()
		// Indent long body for visual grouping.
		for _, line := range strings.Split(e.Long, "\n") {
			fmt.Printf("  %s\n", tui.StyleMuted.Render(line))
		}
	}
	fmt.Println()
}
