// Command cruft is a single-binary cleaner for dev-laptop mess.
//
// Run `cruft` for the interactive TUI. Use subcommands for scripting.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/sachincool/cruft/internal/audit"
	"github.com/sachincool/cruft/internal/cleaner"
	_ "github.com/sachincool/cruft/internal/cleaner/all"
	"github.com/sachincool/cruft/internal/fsutil"
	"github.com/sachincool/cruft/internal/profile"
	"github.com/sachincool/cruft/internal/runner"
	"github.com/sachincool/cruft/internal/tombstone"
	"github.com/sachincool/cruft/internal/tui"
)

var (
	flagDryRun        bool
	flagYes           bool
	flagMaxParallel   int
	flagStaleDays     int
	flagTombstoneDays int
	// flagSafe replaces the older flagNoTombstone. Default off — most
	// of what cruft cleans regenerates on next use, so the tombstone
	// safety net was theatre for the common case (and consumed disk
	// before its auto-sweep). Opt in with --safe when you want the
	// recoverable-for-N-days behaviour.
	flagSafe         bool
	flagBudget       string
	flagProfile      string
	flagAuditDir     string
	flagJSON         bool
	flagSearchRoots  []string
	flagIncludeRisky bool
	flagReclaimSnaps bool

	versionString = "0.1.0-dev"
)

func main() {
	if err := root().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "cruft:", err)
		os.Exit(1)
	}
}

func root() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cruft",
		Short: "Reclaim disk from caches, build artifacts, and stale tool state.",
		Long: `cruft scans every cache, build artifact, and stale .terraform/ on your Mac
in parallel, shows you the receipts, and only deletes what you tick.

Deletes on confirm. Caches regenerate on next use. Use --dry-run to preview;
use --safe to keep a 7-day undo window (recoverable with 'cruft restore').

Quick start:
  cruft                              interactive TUI (pick, confirm, done)
  cruft run --all                    non-interactive — delete everything safe
  cruft run --all --dry-run          preview only, nothing changes
  cruft run --all --safe             delete via 7-day recycle (restorable)
  cruft doctor                       what's installed, what's in use
  cruft last                         what was deleted in the last run
  cruft glossary                     define every term cruft uses
  cruft explain <name>               docs for one cleaner (e.g. xcode-derived)`,
		RunE: runTUI,
		// Accept positional cleaner names (e.g. `cruft gomod pip`) and pass
		// them to runTUI. Without this, cobra rejects any non-subcommand arg
		// with "unknown command". Subcommands still match first.
		Args:              cobra.ArbitraryArgs,
		ValidArgsFunction: completeCleanerNames,
		SilenceUsage:      true,
		SilenceErrors:     true,
		// Reject --profile typos up front: profile.Parse falls back to
		// balanced, so `--profile agressive` would otherwise silently run
		// a different scope than the user asked for.
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if !profile.Valid(flagProfile) {
				return fmt.Errorf("unknown --profile %q — use conservative, balanced, or aggressive", flagProfile)
			}
			return nil
		},
	}
	cmd.PersistentFlags().BoolVar(&flagDryRun, "dry-run", false, "preview only — show what would happen, change nothing")
	cmd.PersistentFlags().BoolVarP(&flagYes, "yes", "y", false, "skip the confirm prompt")
	cmd.PersistentFlags().IntVar(&flagMaxParallel, "max-parallel", runtime.NumCPU(), "concurrent cleaners")
	cmd.PersistentFlags().IntVar(&flagStaleDays, "stale-days", 30, "consider .terraform/ etc stale after this many idle days")
	cmd.PersistentFlags().IntVar(&flagTombstoneDays, "tombstone-days", 7, "how long --safe deletions stay recoverable")
	cmd.PersistentFlags().BoolVar(&flagSafe, "safe", false, "route deletions through a 7-day recycle so `cruft restore` works")
	cmd.PersistentFlags().StringVar(&flagBudget, "budget", "", "stop once this much is reclaimed (e.g. 5GB)")
	cmd.PersistentFlags().StringVar(&flagProfile, "profile", "balanced", "conservative (re-download only) | balanced (default; safe set) | aggressive (incl. risky)")
	cmd.PersistentFlags().StringVar(&flagAuditDir, "audit-dir", defaultAuditDir(), "where to write JSONL audit logs")
	cmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "machine-readable output for non-TUI subcommands")
	cmd.PersistentFlags().StringSliceVar(&flagSearchRoots, "search-root", defaultSearchRoots(), "directories scanned for stale .terraform/ etc")
	cmd.PersistentFlags().BoolVar(&flagIncludeRisky, "include-risky", false, "include cleaners marked risky")
	cmd.PersistentFlags().BoolVar(&flagReclaimSnaps, "reclaim-snapshots", true, "after a live run, thin Time Machine local snapshots so deleted space actually frees (macOS)")

	// Journey-ordered groups so --help reads in the order a user
	// actually does things: explore → run → recover. Cobra defaults
	// to alphabetical (doctor → explain → glossary → history → …),
	// which doesn't match the mental model.
	cmd.AddGroup(
		&cobra.Group{ID: "explore", Title: "Explore:"},
		&cobra.Group{ID: "act", Title: "Clean:"},
		&cobra.Group{ID: "recover", Title: "Inspect & recover:"},
	)
	with := func(c *cobra.Command, group string) *cobra.Command {
		c.GroupID = group
		return c
	}
	cmd.AddCommand(
		with(cmdDoctor(), "explore"),
		with(cmdList(), "explore"),
		with(cmdExplain(), "explore"),
		with(cmdGlossary(), "explore"),
		with(cmdRun(), "act"),
		with(cmdHistory(), "recover"),
		with(cmdLast(), "recover"),
		with(cmdRestore(), "recover"),
		cmdVersion(),
	)
	return cmd
}

func defaultAuditDir() string {
	return filepath.Join(fsutil.HomeDir(), ".local", "share", "cruft", "runs")
}

func defaultTombstoneDir() string {
	return filepath.Join(fsutil.HomeDir(), ".local", "share", "cruft", "tombstone")
}

func defaultSearchRoots() []string {
	candidates := []string{"~/Projects", "~/projects", "~/code", "~/work", "~/src", "~/dev"}
	var out []string
	var infos []os.FileInfo
	for _, c := range candidates {
		p := fsutil.Expand(c)
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		// On case-insensitive APFS ~/Projects and ~/projects are the same
		// directory; scanning both would double-count every finding and
		// then fail the second delete. Dedupe by identity, not by name.
		dup := false
		for _, seen := range infos {
			if os.SameFile(seen, info) {
				dup = true
				break
			}
		}
		if dup {
			continue
		}
		infos = append(infos, info)
		out = append(out, p)
	}
	if len(out) == 0 {
		out = []string{fsutil.HomeDir()}
	}
	return out
}

// buildRunner constructs a runner from the global flags.
func buildRunner() (*runner.Runner, error) {
	budget, err := runner.ParseSize(flagBudget)
	if err != nil {
		// Hide strconv internals — the only useful message is "this isn't a size".
		return nil, fmt.Errorf("invalid --budget %q — try a size like 5GB, 500MB, or 100000000 (bytes)", flagBudget)
	}
	tombstoneDir := defaultTombstoneDir()
	useTomb := flagSafe
	autoRisky := flagIncludeRisky || profile.Parse(flagProfile) == profile.Aggressive
	return runner.New(runner.Options{
		MaxParallel:      flagMaxParallel,
		StaleDays:        flagStaleDays,
		SearchRoots:      flagSearchRoots,
		SamplePaths:      5,
		DryRun:           flagDryRun,
		UseTombstone:     useTomb,
		TombstoneDir:     tombstoneDir,
		AuditDir:         flagAuditDir,
		Budget:           budget,
		AutoApproveRisky: autoRisky,
	})
}

// completeCleanerNames powers shell tab-completion for positional cleaner
// arguments (e.g. `cruft run <TAB>`, `cruft explain <TAB>`). It offers the
// registered cleaner names that aren't already on the command line.
func completeCleanerNames(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	chosen := map[string]bool{}
	for _, a := range args {
		chosen[a] = true
	}
	var names []string
	for _, c := range cleaner.All() {
		if !chosen[c.Name()] && strings.HasPrefix(c.Name(), toComplete) {
			names = append(names, c.Name())
		}
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

// pickCleaners resolves the cleaners to run from flags and positional args.
func pickCleaners(args []string) ([]cleaner.Cleaner, error) {
	all := cleaner.All()
	if len(args) > 0 {
		found, unknown := cleaner.ByNames(args)
		if len(unknown) > 0 {
			return nil, fmt.Errorf("unknown cleaner(s): %s", strings.Join(unknown, ", "))
		}
		return found, nil
	}
	p := profile.Parse(flagProfile)
	if flagIncludeRisky {
		p = profile.Aggressive
	}
	return profile.Filter(all, p), nil
}

func runTUI(_ *cobra.Command, args []string) error {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return fmt.Errorf("not a terminal — try `cruft run --all`")
	}
	// Positional args scope the TUI to the named cleaners (same resolution
	// as `cruft run <cleaner>...`). With none, fall back to the --profile
	// set. Lets `cruft gomod pip` open the picker on just those.
	cleaners, err := pickCleaners(args)
	if err != nil {
		return err
	}
	r, err := buildRunner()
	if err != nil {
		return err
	}
	defer r.Close()
	// Sweep any tombstone runs older than retention up-front. Use a
	// standalone store, not the runner's: the runner only has one when
	// --safe is set, and the promise "recoverable for N days" must be
	// enforced on every launch, not just --safe ones.
	_, _ = tombstone.New(defaultTombstoneDir()).Sweep(time.Duration(flagTombstoneDays) * 24 * time.Hour)

	m := tui.NewModel(context.Background(), r, cleaners, tui.Options{
		ReclaimSnapshots: flagReclaimSnaps,
		SkipConfirm:      flagYes,
	})
	prog := tea.NewProgram(m, tea.WithAltScreen())
	_, err = prog.Run()
	return err
}

// ---- subcommands ----

func cmdRun() *cobra.Command {
	var all bool
	c := &cobra.Command{
		Use:               "run [cleaner...]",
		Short:             "Run cleaners non-interactively.",
		ValidArgsFunction: completeCleanerNames,
		RunE: func(cc *cobra.Command, args []string) error {
			ctx := context.Background()
			var cleaners []cleaner.Cleaner
			var err error
			if all {
				cleaners = profile.Filter(cleaner.All(), profile.Parse(flagProfile))
				if flagIncludeRisky {
					cleaners = profile.Filter(cleaner.All(), profile.Aggressive)
				}
			} else {
				// Bare `cruft run` used to silently behave like `run --all`
				// — immediate live deletion for someone who typed it
				// expecting usage help. Make the full sweep an explicit
				// opt-in.
				if len(args) == 0 {
					return fmt.Errorf("specify cleaners to run (e.g. `cruft run npm gomod`) or pass --all")
				}
				cleaners, err = pickCleaners(args)
				if err != nil {
					return err
				}
			}
			r, err := buildRunner()
			if err != nil {
				return err
			}
			defer r.Close()
			_, _ = tombstone.New(defaultTombstoneDir()).Sweep(time.Duration(flagTombstoneDays) * 24 * time.Hour)

			// Free space before and after the run. The Total below is a
			// sum of finding sizes (an estimate of what was removed); this
			// pair lets the summary show the *measured* disk delta, which
			// can be far smaller — e.g. Docker's sparse VM image doesn't
			// shrink on prune, and APFS/Time Machine snapshots can pin
			// freed blocks until they're thinned.
			beforeFS := fsutil.FreeBytes("/")
			scans, err := r.Scan(ctx, cleaners, nil)
			if err != nil {
				return err
			}
			results := r.Execute(ctx, scans, nil)
			afterFS := fsutil.FreeBytes("/")

			// Reclaim snapshot-pinned space in the same run. Deletions can
			// "succeed" yet leave free space unchanged because Time Machine
			// local snapshots still reference the removed blocks. The thin
			// target is bounded by what this run freed, so older restore
			// points survive.
			var snapReclaimed int64
			if flagReclaimSnaps && !flagDryRun {
				var freed int64
				for _, x := range results {
					freed += x.Result.BytesFreed
				}
				snapReclaimed = fsutil.ReclaimLocalSnapshots(ctx, "/", freed)
				if snapReclaimed > 0 {
					afterFS = fsutil.FreeBytes("/")
				}
			}

			if flagJSON {
				if err := json.NewEncoder(os.Stdout).Encode(runJSONOutput{
					RunID:              r.RunID(),
					AuditLog:           r.AuditPath(),
					Tombstone:          r.TombstoneRoot(),
					FreeBefore:         beforeFS,
					FreeAfter:          afterFS,
					ActualFreed:        afterFS - beforeFS,
					SnapshotsReclaimed: snapReclaimed,
					Results:            resultsToJSON(results),
				}); err != nil {
					return err
				}
				return failuresAsError(results)
			}
			if err := printSummary(results, r, beforeFS, afterFS, snapReclaimed); err != nil {
				return err
			}
			// Scripted callers need the exit code to reflect failed
			// deletions; the summary alone is invisible to a script.
			return failuresAsError(results)
		},
	}
	c.Flags().BoolVar(&all, "all", false, "run all non-risky cleaners (per --profile)")
	return c
}

func cmdList() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List registered cleaners.",
		RunE: func(_ *cobra.Command, _ []string) error {
			all := cleaner.All()
			if flagJSON {
				rows := make([]listRow, 0, len(all))
				for _, c := range all {
					rows = append(rows, listRow{
						Name:     c.Name(),
						Category: string(c.Category()),
						Risky:    c.Risky(),
					})
				}
				return json.NewEncoder(os.Stdout).Encode(rows)
			}
			// Group by category so users see "Containers & VMs" / "IaC tool
			// state" instead of cryptic slugs. Inline the per-cleaner risk
			// reason so the (risky) tag stops being a mystery in the list view.
			fmt.Printf("%s  %s\n\n",
				tui.StyleTitle.Render(fmt.Sprintf("%d cleaners", len(all))),
				tui.StyleMuted.Render("· `cruft explain <name>` for details, `cruft glossary risky` for what risky means"),
			)
			byCat := map[cleaner.Category][]cleaner.Cleaner{}
			for _, c := range all {
				byCat[c.Category()] = append(byCat[c.Category()], c)
			}
			for _, cat := range []cleaner.Category{
				cleaner.CategoryLangPkg, cleaner.CategoryIaC,
				cleaner.CategoryContainer, cleaner.CategorySystem,
				cleaner.CategoryProject,
			} {
				cs := byCat[cat]
				if len(cs) == 0 {
					continue
				}
				fmt.Println(tui.StyleAccent.Render(cat.Title()))
				for _, c := range cs {
					risky := ""
					if c.Risky() {
						risky = "  " + tui.StyleWarn.Render("risky")
						if r := c.RiskReason(); r != "" {
							risky += tui.StyleMuted.Render(" · " + r)
						}
					}
					fmt.Printf("  %-22s%s\n", c.Name(), risky)
				}
				fmt.Println()
			}
			return nil
		},
	}
}

func cmdDoctor() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Show each cleaner's detect/busy status.",
		RunE: func(_ *cobra.Command, _ []string) error {
			ctx := context.Background()

			// Compact key UP TOP. Earlier this lived at the bottom, but
			// users hit "risky" 25 rows before reaching the explanation
			// and gave up — anchor terms before the table, not after.
			fmt.Println(tui.StyleMuted.Render("Legend  ") +
				tui.StyleAccent.Render("ready") + tui.StyleMuted.Render(" = safe to clean   ·   ") +
				tui.StyleMuted.Render("in use") + tui.StyleMuted.Render(" = skipped (a process is running)   ·   ") +
				tui.StyleWarn.Render("risky") + tui.StyleMuted.Render(" = unchecked by default; run `cruft explain <name>` for why"))
			fmt.Println()

			for _, c := range cleaner.All() {
				detected := c.Detect(ctx)
				busy := ""
				if detected {
					busy = fsutil.AnyProcessRunning(ctx, c.BusyProcesses())
				}
				status := "  -"
				switch {
				case !detected:
					status = "  not installed"
				case busy != "":
					status = "  in use (" + busy + ")"
				default:
					status = tui.StyleAccent.Render("  ready")
				}
				risky := ""
				if c.Risky() {
					risky = tui.StyleWarn.Render(" risky")
					if r := c.RiskReason(); r != "" {
						risky += tui.StyleMuted.Render(" · " + r)
					}
				}
				fmt.Printf("  %-22s  %s%s\n", c.Name(), status, risky)
			}
			fmt.Println()
			fmt.Println(tui.StyleMuted.Render("Search roots:"))
			for _, root := range flagSearchRoots {
				fmt.Println("  " + root)
			}
			ts := tombstone.New(defaultTombstoneDir())
			fmt.Println()
			fmt.Printf("%s %s\n", tui.StyleMuted.Render("Tombstone size:"), tui.HumanBytes(ts.Size()))
			fmt.Println(tui.StyleMuted.Render("Tip: `cruft glossary <term>` defines any word here · `cruft explain <name>` for any cleaner"))
			return nil
		},
	}
}

func cmdExplain() *cobra.Command {
	return &cobra.Command{
		Use:               "explain <cleaner>",
		Short:             "Show what a cleaner does, without running anything.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeCleanerNames,
		RunE: func(_ *cobra.Command, args []string) error {
			c := cleaner.ByName(args[0])
			if c == nil {
				return fmt.Errorf("unknown cleaner: %s", args[0])
			}
			fmt.Println(tui.StyleTitle.Render(c.Name()))
			fmt.Println(tui.StyleMuted.Render(string(c.Category())))
			if c.Risky() {
				reason := c.RiskReason()
				if reason == "" {
					reason = "unchecked by default"
				}
				fmt.Println(tui.StyleWarn.Render("⚠  risky · " + reason))
			}
			fmt.Println()
			fmt.Println(c.Description())
			if procs := c.BusyProcesses(); len(procs) > 0 {
				fmt.Println()
				fmt.Println(tui.StyleMuted.Render("Skipped while these are running: " + strings.Join(procs, ", ")))
			}
			return nil
		},
	}
}

func cmdHistory() *cobra.Command {
	var limit int
	c := &cobra.Command{
		Use:   "history",
		Short: "Show past runs from the audit log.",
		RunE: func(_ *cobra.Command, _ []string) error {
			runs, err := audit.History(flagAuditDir)
			if err != nil {
				return err
			}
			if limit > 0 && len(runs) > limit {
				runs = runs[:limit]
			}
			if flagJSON {
				return json.NewEncoder(os.Stdout).Encode(runs)
			}
			if len(runs) == 0 {
				fmt.Println(tui.StyleMuted.Render("no runs yet — try `cruft run --all` to see what's reclaimable"))
				return nil
			}
			for _, r := range runs {
				marker := ""
				if r.DryRun {
					marker = " (dry)"
				}
				fmt.Printf("%s   %s freed   %d files%s\n",
					r.RunID, tui.StyleAccent.Render(tui.HumanBytes(r.BytesFreed)),
					r.Successes, marker,
				)
			}
			return nil
		},
	}
	c.Flags().IntVar(&limit, "limit", 20, "max runs to show")
	return c
}

func cmdLast() *cobra.Command {
	return &cobra.Command{
		Use:   "last",
		Short: "Show the most recent run, with per-cleaner detail.",
		RunE: func(_ *cobra.Command, _ []string) error {
			runs, err := audit.History(flagAuditDir)
			if err != nil {
				return err
			}
			if len(runs) == 0 {
				fmt.Println(tui.StyleMuted.Render("no runs yet — try `cruft run --all` to see what's reclaimable"))
				return nil
			}
			r := runs[0]
			fmt.Println(tui.StyleTitle.Render(r.RunID))
			fmt.Printf("Freed: %s   Files: %d   Failures: %d\n",
				tui.StyleAccent.Render(tui.HumanBytes(r.BytesFreed)),
				r.Successes, r.Failures,
			)
			fmt.Println()
			for name, bytes := range r.PerCleaner {
				fmt.Printf("  %-22s  %s\n", name, tui.HumanBytes(bytes))
			}
			fmt.Println()
			fmt.Println(tui.StyleMuted.Render(r.Path))
			return nil
		},
	}
}

func cmdRestore() *cobra.Command {
	return &cobra.Command{
		Use:   "restore <run-id>",
		Short: "Recover files from a tombstoned run.",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			ts := tombstone.New(defaultTombstoneDir())
			restored, skipped, err := ts.Restore(args[0])
			if err != nil {
				// Wrap raw filesystem errors so a noob sees actionable text
				// instead of "open /path/manifest.jsonl: no such file or directory".
				if os.IsNotExist(err) {
					return fmt.Errorf(
						"no recycled run with id %q.\n\n"+
							"Reasons this can happen:\n"+
							"  · the run id is wrong         (`cruft history` lists known runs)\n"+
							"  · the run didn't use --safe   (only --safe runs are recoverable)\n"+
							"  · the recycle window expired  (default %d days)",
						args[0], flagTombstoneDays,
					)
				}
				return fmt.Errorf("restore failed: %w", err)
			}
			fmt.Printf("Restored %d, skipped %d.\n", len(restored), len(skipped))
			for _, p := range skipped {
				fmt.Println(tui.StyleWarn.Render("  skipped (already exists): " + p))
			}
			return nil
		},
	}
}

func cmdVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the cruft version.",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(versionString)
		},
	}
}

// failuresAsError converts failed deletions into a non-nil error so the
// process exits 1 when any approved finding could not be cleaned.
func failuresAsError(results []runner.ExecResult) error {
	var failed int
	for _, x := range results {
		failed += x.Result.Failed
	}
	if failed > 0 {
		return fmt.Errorf("%d deletion(s) failed — see the summary above or the audit log", failed)
	}
	return nil
}

// ---- output helpers ----

type listRow struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Risky    bool   `json:"risky"`
}

type runJSONOutput struct {
	RunID              string          `json:"run_id"`
	AuditLog           string          `json:"audit_log"`
	Tombstone          string          `json:"tombstone"`
	FreeBefore         int64           `json:"free_before_bytes"`
	FreeAfter          int64           `json:"free_after_bytes"`
	ActualFreed        int64           `json:"actual_freed_bytes"`
	SnapshotsReclaimed int64           `json:"snapshots_reclaimed_bytes,omitempty"`
	Results            []runResultJSON `json:"results"`
}

type runResultJSON struct {
	Cleaner      string   `json:"cleaner"`
	Findings     int      `json:"findings"`
	BytesFreed   int64    `json:"bytes_freed"`
	Succeeded    int      `json:"succeeded"`
	Failed       int      `json:"failed"`
	DurationMS   int64    `json:"duration_ms"`
	NotInstalled bool     `json:"not_installed,omitempty"`
	BusyProcess  string   `json:"busy_process,omitempty"`
	Errors       []string `json:"errors,omitempty"`
}

func resultsToJSON(rs []runner.ExecResult) []runResultJSON {
	out := make([]runResultJSON, 0, len(rs))
	for _, r := range rs {
		var errs []string
		for _, e := range r.Result.Errors {
			errs = append(errs, e.Error())
		}
		out = append(out, runResultJSON{
			Cleaner:      r.Cleaner.Name(),
			Findings:     r.Result.Findings,
			BytesFreed:   r.Result.BytesFreed,
			Succeeded:    r.Result.Succeeded,
			Failed:       r.Result.Failed,
			DurationMS:   r.Result.DurationMS,
			NotInstalled: r.NotInstalled,
			BusyProcess:  r.BusyProcess,
			Errors:       errs,
		})
	}
	return out
}

func printSummary(results []runner.ExecResult, r *runner.Runner, beforeFS, afterFS, snapReclaimed int64) error {
	type riskyOffer struct {
		name   string
		reason string
		bytes  int64
	}
	var (
		totalFreed   int64
		busy         []string // "name (proc in use)"
		notInstalled []string
		riskyShown   []string     // risky cleaners that ran (auto-approved)
		riskyOffered []riskyOffer // risky scanned but not approved — surface as opt-in
	)
	for _, x := range results {
		totalFreed += x.Result.BytesFreed
		switch {
		case x.NotInstalled:
			notInstalled = append(notInstalled, x.Cleaner.Name())
		case x.BusyProcess != "":
			busy = append(busy, fmt.Sprintf("%s (%s in use)", x.Cleaner.Name(), x.BusyProcess))
		}
		// Look at SCAN findings, not execute result.Findings — when nothing
		// gets approved, the runner never invokes Execute at all, so
		// Result.Findings stays 0. The risky-offer must read from the
		// scan phase.
		if x.Cleaner.Risky() && len(x.Findings) > 0 {
			if x.Result.Succeeded > 0 {
				riskyShown = append(riskyShown, x.Cleaner.Name())
			} else {
				var bytes int64
				for _, f := range x.Findings {
					bytes += f.Bytes
				}
				riskyOffered = append(riskyOffered, riskyOffer{
					name:   x.Cleaner.Name(),
					reason: x.Cleaner.RiskReason(),
					bytes:  bytes,
				})
			}
		}
	}
	mode := "dry run"
	if !r.IsDryRun() {
		mode = "live"
	}
	fmt.Println()
	fmt.Println(tui.StyleTitle.Render("cruft summary") + tui.StyleMuted.Render("  ("+mode+")"))

	// Risky-included banner: triggers whenever a risky cleaner produced
	// findings (whether the user selected --profile aggressive or
	// --include-risky). Previously this was invisible; a noob couldn't
	// tell that aggressive was destructive-by-comparison.
	if len(riskyShown) > 0 {
		fmt.Println(tui.StyleWarn.Render("⚠  includes risky cleaners: "+strings.Join(riskyShown, ", ")) +
			tui.StyleMuted.Render("  · `cruft explain <name>` for each"))
	}

	for _, x := range results {
		if x.NotInstalled || x.BusyProcess != "" {
			continue
		}
		// Keep rows that did no work but were skipped or errored (e.g.
		// budget exhausted) — selected items must not vanish from the
		// receipt.
		if x.Result.Findings == 0 && x.Result.Skipped == 0 && len(x.Result.Errors) == 0 {
			continue
		}
		marker := tui.StyleAccent.Render("✓")
		failedRow := x.Result.Failed > 0 || (x.Result.Succeeded == 0 && len(x.Result.Errors) > 0)
		switch {
		case failedRow:
			marker = tui.StyleDanger.Render("✗")
		case x.Result.Skipped > 0 && x.Result.Succeeded == 0:
			marker = tui.StyleMuted.Render("⏸")
		}
		fmt.Printf("  %s %-22s %s\n", marker, x.Cleaner.Name(),
			tui.StyleAccent.Render(tui.HumanBytes(x.Result.BytesFreed)))
		if failedRow && len(x.Result.Errors) > 0 {
			fmt.Println(tui.StyleMuted.Render("      ↳ " + x.Result.Errors[0].Error()))
		}
	}

	// Risky-but-unapproved findings: scanned, found stuff, but not
	// executed because they're risky and the user didn't opt in.
	// Surface as an explicit offer so the user knows what's available
	// AND knows exactly how to act on it.
	if len(riskyOffered) > 0 {
		var totalRisky int64
		for _, ro := range riskyOffered {
			totalRisky += ro.bytes
		}
		fmt.Println()
		fmt.Println(tui.StyleWarn.Render(
			fmt.Sprintf("⚠  %s more available from risky cleaners (not executed):", tui.HumanBytes(totalRisky)),
		))
		for _, ro := range riskyOffered {
			fmt.Printf("    %-22s %s   %s\n",
				ro.name,
				tui.StyleAccent.Render(tui.HumanBytes(ro.bytes)),
				tui.StyleMuted.Render("· "+ro.reason),
			)
		}
		fmt.Println(tui.StyleMuted.Render("    to also clean these, re-run with `--include-risky` or `--profile aggressive`"))
	}

	// Show what was skipped. Without this, the noob sees "Total: 6.8 GB"
	// and thinks that's all there is — when in fact (e.g.) xcode-derived
	// (5 GB) was skipped because Xcode is running. Surface it explicitly.
	if len(busy) > 0 {
		fmt.Println()
		fmt.Println(tui.StyleWarn.Render("⏸  skipped because a related tool is running:"))
		for _, s := range busy {
			fmt.Println("    " + tui.StyleMuted.Render(s))
		}
		fmt.Println(tui.StyleMuted.Render("    re-run after the tool exits to capture this reclaim."))
	}
	if len(notInstalled) > 0 {
		fmt.Println()
		fmt.Println(tui.StyleMuted.Render("not installed on this machine: " + strings.Join(notInstalled, ", ")))
	}

	fmt.Println()
	fmt.Printf("Total: %s %s\n",
		tui.StyleAccent.Render(tui.HumanBytes(totalFreed)),
		tui.StyleMuted.Render("(sum of what was removed)"),
	)

	// Measured disk delta — the honest "what actually changed on the
	// volume" number, from statfs before-scan vs after-execute. In a
	// live run this is what the user should trust: the Total above is a
	// sum of file sizes, but freed bytes can stay claimed by APFS / Time
	// Machine local snapshots, and Docker's prune never shrinks its
	// sparse VM image. When the two diverge, this line shows it instead
	// of hiding it behind an optimistic headline.
	if beforeFS > 0 && afterFS > 0 {
		if r.IsDryRun() {
			predicted := beforeFS + totalFreed
			fmt.Printf("Disk free: %s → %s   %s\n",
				tui.HumanBytes(beforeFS),
				tui.StyleAccent.Render(tui.HumanBytes(predicted)),
				tui.StyleMuted.Render("("+tui.HumanBytes(totalFreed)+" would be freed)"),
			)
		} else {
			freed := max(afterFS-beforeFS, 0)
			fmt.Printf("Disk free: %s → %s   %s\n",
				tui.HumanBytes(beforeFS),
				tui.StyleAccent.Render(tui.HumanBytes(afterFS)),
				tui.StyleAccent.Render("("+tui.HumanBytes(freed)+" actually freed)"),
			)
			if snapReclaimed > 0 {
				fmt.Println(tui.StyleMuted.Render(
					"   incl. " + tui.HumanBytes(snapReclaimed) +
						" recovered by thinning Time Machine local snapshots",
				))
			}
			// Loud, only when it still matters after the snapshot thin:
			// deletions "succeeded" but the volume barely moved. The usual
			// remaining cause on a dev box is Docker — prune reports space
			// reclaimed but never shrinks its sparse VM image on macOS.
			if totalFreed >= 1<<30 && freed < totalFreed/2 {
				fmt.Println(tui.StyleWarn.Render(
					"⚠  still less freed than removed — likely Docker's sparse image (prune doesn't shrink it)",
				))
				fmt.Println(tui.StyleMuted.Render(
					"   or snapshots pinned by an active backup. Check: `tmutil listlocalsnapshots /`",
				))
			}
		}
	}
	if p := r.AuditPath(); p != "" {
		fmt.Println(tui.StyleMuted.Render("Audit log: " + p))
	}
	if r.UsesTombstone() && !r.IsDryRun() {
		fmt.Println(tui.StyleMuted.Render(
			"Tombstone: " + r.TombstoneRoot() + "  (restore with `cruft restore " + r.RunID() + "`)",
		))
	}
	// Tip line — seeds discovery of glossary & explain. Cheap nudge
	// so users learn the tool from inside the tool.
	fmt.Println(tui.StyleMuted.Render(
		"Tip: `cruft glossary` defines every term · `cruft explain <name>` for any cleaner",
	))
	return nil
}
