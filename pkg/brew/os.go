package brew

import (
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "strings"

    "github.com/hashicorp/go-version"
)

type OSVariant int

const (
    Unknown OSVariant = iota
    Mojave
    Catalina
    BigSur
    Arm64BigSur
    X8664Linux
)

func (o OSVariant) String() string {
    return [...]string{
        "Unknown",
        "Mojave",
        "Catalina",
        "Big Sur (Intel)",
        "Big Sur (Apple Silicon)",
        "x86_64 Linux",
    }[o]
}

func (o OSVariant) PlatformString() string {
    return [...]string{
        "unknown",
        "mojave",
        "catalina",
        "big_sur",
        "arm64_big_sur",
        "x86_64_linux",
    }[o]
}

func (o OSVariant) HomebrewCache() string {
    switch o {
    case Mojave, Catalina, BigSur, Arm64BigSur:
        return filepath.Join(os.Getenv("HOME"), "Library", "Caches", "Homebrew", "quickbrew")
    case X8664Linux:
        return filepath.Join(os.Getenv("HOME"), ".cache", "Homebrew", "quickbrew")
    default:
        return ""
    }
}

func (o OSVariant) HomebrewPrefix() string {
    switch o {
    case Mojave, Catalina, BigSur:
        return filepath.Join("/usr", "local")
    case Arm64BigSur:
        return filepath.Join("/opt", "homebrew")
    case X8664Linux:
        return filepath.Join("/home", "linuxbrew", ".linuxbrew")
    default:
        return ""
    }
}

func (o OSVariant) HomebrewCellar() string {
    return filepath.Join(o.HomebrewPrefix(), "Cellar")
}

func (o OSVariant) HomebrewLinkedKegs() string {
    return filepath.Join(o.HomebrewPrefix(), "var", "homebrew", "linked")
}

func (o OSVariant) HomebrewFormulaAPIRoot() string {
    switch o {
    case X8664Linux:
        return "https://formulae.brew.sh/api/formula-linux/"
    default:
        return "https://formulae.brew.sh/api/formula/"
    }
}

func (o OSVariant) HomebrewFormulaAliasesRoot() string {
    switch o {
    case X8664Linux:
        return "https://raw.githubusercontent.com/Homebrew/linuxbrew-core/master/Aliases/"
    default:
        return "https://raw.githubusercontent.com/Homebrew/homebrew-core/master/Aliases/"
    }
}

func CurrentPlatform() OSVariant {
    if runtime.GOOS == "darwin" {
        output, err := exec.Command("defaults", "read", "loginwindow", "SystemVersionStampAsString").Output()
        if err != nil {
            panic(err)
        }

        systemVersion := version.Must(version.NewVersion(strings.TrimSpace(string(output))))
        switch {
        case systemVersion.Compare(version.Must(version.NewVersion("11.0"))) >= 1:
            switch runtime.GOARCH {
            case "amd64":
                return BigSur
            case "arm64":
                return Arm64BigSur
            default:
                return Unknown
            }
        case systemVersion.Compare(version.Must(version.NewVersion("10.15"))) >= 1:
            return Catalina
        case systemVersion.Compare(version.Must(version.NewVersion("10.14"))) >= 1:
            return Mojave
        default:
            return Unknown
        }
    } else if runtime.GOOS == "linux" {
        switch runtime.GOARCH {
        case "amd64":
            return X8664Linux
        default:
            return Unknown
        }
    } else {
        return Unknown
    }
}
