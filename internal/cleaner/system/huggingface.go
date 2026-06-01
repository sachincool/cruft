package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// The HF hub cache holds downloaded model/dataset snapshots — can be
	// very large, and re-downloading is slow.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:       "huggingface",
		CategoryValue:   cleaner.CategorySystem,
		Desc:            "Hugging Face hub cache (~/.cache/huggingface/hub) — downloaded models + datasets.",
		IsRisky:         true,
		RiskReasonValue: "re-downloading model/dataset snapshots is large and slow",
		Paths:           []string{"~/.cache/huggingface/hub"},
		DetectAnyPath:   true,
		Reason:          "huggingface hub cache (risky — large re-download)",
	})
}
