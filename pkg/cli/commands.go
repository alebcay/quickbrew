package cli

import (
	"fmt"
	"runtime"

	"github.com/alebcay/quickbrew/pkg/brew"
)

func Info(formulaName string) {
	formula := brew.GetLatestFormula(formulaName)

	fmt.Printf("%s %s\n%s\n%s\nLicense: %s\n", formula.Name, formula.Version, formula.Desc, formula.Homepage, formula.License)
}

func Pour(formulaName string) {
	formula := brew.GetLatestFormula(formulaName)
	formula.PourBottle()
}

func Install(formulaName string) {
	formula := brew.GetLatestFormula(formulaName)
	if runtime.GOOS == "linux" {
		brew.SymlinkLdSo()
		patchelfBare := brew.GetBareFormula("patchelf")
		if !patchelfBare.IsInstalled() {
			patchelf := brew.GetLatestFormula("patchelf")
			patchelf.PourBottle()
			brew.GetKegFromFormula(patchelf).Link()
		}
	}
	formula.PourBottle()
	brew.GetKegFromFormula(formula).Link()
}

func Remove(formulaName string) {
	formula := brew.GetBareFormula(formulaName)
	formula.Remove()
}
