package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// Model weights are often the single largest reclaimable directory on
	// a dev laptop (tens of GB), but re-pulling them is a slow, large
	// download — hence risky.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:       "ollama",
		CategoryValue:   cleaner.CategorySystem,
		Desc:            "Ollama downloaded model weights (~/.ollama/models). Often tens of GB.",
		IsRisky:         true,
		RiskReasonValue: "re-pulling models is a large, slow download (often tens of GB)",
		Paths:           []string{"~/.ollama/models"},
		DetectAnyPath:   true,
		BusyProcs:       []string{"ollama"},
		Reason:          "ollama model weights (risky — large re-download)",
	})
}
