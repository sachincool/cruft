package container

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// Earlier versions tried to "trim" the VM's diffdisk via
	// `limactl disk sparsify` — a command that doesn't exist (limactl
	// disk manages additional disks; its subcommands are create/list/
	// delete/unlock/resize/import), so the cleaner failed on every run
	// while claiming the whole image size as reclaimable. What IS
	// safely reclaimable is the download cache: base images and ISOs
	// colima/lima keep around, re-fetched on the next `colima start`
	// that needs them.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "colima",
		CategoryValue: cleaner.CategoryContainer,
		Desc:          "Colima/Lima download caches (base VM images, ISOs). Re-downloads on next `colima start` that needs them. The VM disk and its containers are NOT touched.",
		Paths:         []string{"~/Library/Caches/colima", "~/Library/Caches/lima"},
		DetectCmd:     "colima",
		DetectAnyPath: true,
		Reason:        "colima/lima download cache",
	})
}
