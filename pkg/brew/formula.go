package brew

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "path/filepath"
    "strings"

    "github.com/mholt/archiver/v3"
    "github.com/tidwall/gjson"
)

type Formula struct {
    Name              string   `json:"name"`
    VersionedFormulae []string `json:"versioned_formulae"`
    Desc              string   `json:"desc"`
    License           string   `json:"license"`
    Homepage          string   `json:"homepage"`
    Version           string
    Revision          int      `json:"revision"`
    VersionScheme     int      `json:"version_scheme"`
    KegOnly           bool     `json:"keg_only"`
    BottleDisabled    bool     `json:"bottle_disabled"`
    BottleURL         string
    BottleChecksum    string
    BuildDependencies []string `json:"build_dependencies"`
    Dependencies      []string `json:"dependencies"`
    UsesFromMacos     []string `json:"uses_from_macos"`
    Requirements      []string `json:"requirements"`
    ConflictsWith     []string `json:"conflicts_with"`
    Caveats           []string `json:"caveats"`
    Installed         []struct {
        Version             string        `json:"version"`
        RuntimeDependencies []struct {
            FullName string `json:"full_name"`
            Version  string `json:"version"`
        } `json:"runtime_dependencies"`
        InstalledAsDependency bool `json:"installed_as_dependency"`
        InstalledOnRequest    bool `json:"installed_on_request"`
    } `json:"installed"`
    Deprecated        bool        `json:"deprecated"`
    DeprecationDate   string `json:"deprecation_date"`
    DeprecationReason string `json:"deprecation_reason"`
    Disabled          bool        `json:"disabled"`
    DisableDate       string `json:"disable_date"`
    DisableReason     string `json:"disable_reason"`
}

func GetBareFormula(name string) *Formula {
    return &Formula{
        Name: name,
    }
}

func GetLatestFormula(name string) *Formula {
    formulaJSONFile := FetchFormula(name)
    var result Formula

    formulaJSON, err := ioutil.ReadFile(formulaJSONFile)
    if err != nil {
        panic(err)
    }

    json.Unmarshal([]byte(formulaJSON), &result)
    result.BottleURL = gjson.Get(string(formulaJSON), fmt.Sprintf("bottle.stable.files.%s.url", CurrentPlatform().PlatformString())).String()
    result.BottleChecksum = gjson.Get(string(formulaJSON), fmt.Sprintf("bottle.stable.files.%s.sha256", CurrentPlatform().PlatformString())).String()
    result.Version = gjson.Get(string(formulaJSON), "versions.stable").String()
    return &result
}

func (f *Formula) GetKegPath() string {
    return filepath.Join(CurrentPlatform().HomebrewCellar(), f.Name)
}

func (f *Formula) GetFormulaPrefix() string {
    switch f.Revision {
    case 0:
        return filepath.Join(f.GetKegPath(), f.Version)
    default:
        return filepath.Join(f.GetKegPath(), fmt.Sprintf("%s_%d", f.Version, f.Revision))
    }
}

func FetchFormula(name string) string {
    resp, err := http.Head(fmt.Sprintf("%s%s.json", CurrentPlatform().HomebrewFormulaAPIRoot(), name))
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
        return FetchResource(fmt.Sprintf("%s%s.json", CurrentPlatform().HomebrewFormulaAPIRoot(), name), false)
    } else {
        alias_resp, err := http.Get(fmt.Sprintf("%s%s", CurrentPlatform().HomebrewFormulaAliasesRoot(), name))
        if err != nil {
            panic(err)
        }
        defer alias_resp.Body.Close()
        if alias_resp.StatusCode >= 200 && alias_resp.StatusCode <= 299 {
            bodyBytes, err := ioutil.ReadAll(alias_resp.Body)
            if err != nil {
                panic(err)
            }
            real_name := strings.TrimSuffix(filepath.Base(string(bodyBytes)), ".rb")
            return FetchResource(fmt.Sprintf("%s%s.json", CurrentPlatform().HomebrewFormulaAPIRoot(), real_name), false)
        } else {
            panic("No such formula!")
        }
    }
}

func (f *Formula) FetchBottle() string {
    return FetchResourceWithChecksum(f.BottleURL, f.BottleChecksum, true)
}

func (f *Formula) PourBottle() {
    bottlePath := f.FetchBottle()
    if f.IsInstalled() { return }
    fmt.Printf("pour\t%s\n", bottlePath)
    err := archiver.Unarchive(bottlePath, CurrentPlatform().HomebrewCellar())
    if err != nil {
        panic(err)
    }
}

func (f *Formula) IsInstalled() bool {
    _, err := os.Stat(f.GetKegPath())
    return err == nil
}

func (f *Formula) IsUpToDate() bool {
    if !f.IsInstalled() { return false }
    _, err := os.Stat(f.GetFormulaPrefix())
    return err == nil
}

func (f *Formula) Remove() {
    os.RemoveAll(f.GetKegPath())
}
