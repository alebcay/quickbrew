package brew

import (
    "os"
    "path/filepath"
    "regexp"
    "strings"

    "github.com/alebcay/quickbrew/pkg/utils"
)

type Keg struct {
    Formula *Formula
}

type LinkStrategy int

const (
    SkipFile LinkStrategy = iota
    SkipDir
    Mkpath
    Link
    Info
)

var PYC_EXTENSIONS = []string{".pyc", ".pyo"}
var INFOFILE_RX = regexp.MustCompile(`(?m)info/([^.].*?\.info|dir)$`)
var ICONTHEMECACHE_RX = regexp.MustCompile(`(?m)^icons/.*/icon-theme\.cache$`)
var LOCALEDIR_RX = regexp.MustCompile(`(?m)(locale|man)/([a-z]{2}|C|POSIX)(_[A-Z]{2})?(\.[a-zA-Z\-0-9]+(@.+)?)?`)
var SHARE_PATHS = []string{
    "aclocal", "doc", "info", "java", "locale", "man",
    "man/man1", "man/man2", "man/man3", "man/man4",
    "man/man5", "man/man6", "man/man7", "man/man8",
    "man/cat1", "man/cat2", "man/cat3", "man/cat4",
    "man/cat5", "man/cat6", "man/cat7", "man/cat8",
    "applications", "gnome", "gnome/help", "icons",
    "mime-info", "pixmaps", "sounds", "postgresql",
}
var PYTHON23_RX = regexp.MustCompile(`^python[23]\.\d`)
var FRAMEWORKSVERSIONS_RX = regexp.MustCompile(`[^/]*\.framework(/Versions)?`)
var KEG_LINK_DIRECTORIES = []string{"bin", "etc", "include", "lib", "sbin", "share", "var"}

func (k *Keg) CellarLocation() string {
    return k.Formula.GetFormulaPrefix()
}

func (k *Keg) OptRecord() string {
    return filepath.Join(CurrentPlatform().HomebrewPrefix(), "opt", k.Formula.Name)
}

func (k *Keg) LinkedKegRecord() string {
    return filepath.Join(CurrentPlatform().HomebrewLinkedKegs(), k.Formula.Name)
}

func (k *Keg) Link() {
    if k.Formula.KegOnly { return }
    cellarLocation := k.CellarLocation()
    optRecord := k.OptRecord()
    if _, err := os.Lstat(optRecord); !os.IsNotExist(err) { os.Remove(optRecord) }
    makeRelativeSymlink(cellarLocation, optRecord)

    linkToCellar(cellarLocation, "etc", func(path string) LinkStrategy {
        return Mkpath
    })

    linkToCellar(cellarLocation, "bin", func(path string) LinkStrategy {
        return SkipDir
    })

    linkToCellar(cellarLocation, "sbin", func(path string) LinkStrategy {
        return SkipDir
    })

    linkToCellar(cellarLocation, "include", func(path string) LinkStrategy {
        return Link
    })

    linkToCellar(cellarLocation, "share", func(path string) LinkStrategy {
        switch {
        case INFOFILE_RX.MatchString(path):
            return Info
        case path == "locale/locale.alias", ICONTHEMECACHE_RX.MatchString(path):
            return SkipFile
        case LOCALEDIR_RX.MatchString(path), strings.HasPrefix(path, "icons/"), strings.HasPrefix(path, "zsh"),
        strings.HasPrefix(path, "fish"), strings.HasPrefix(path, "lua/"), strings.HasPrefix(path, "guile/"),
        utils.StringSliceContains(path, SHARE_PATHS):
            return Mkpath
        default:
            return Link
        }
    })

    linkToCellar(cellarLocation, "lib", func(path string) LinkStrategy {
        switch {
        case path == "charset.alias":
            return SkipFile
        case path == "pkgconfig", path == "cmake", path == "dtrace", strings.HasPrefix(path, "gdk-pixbuf"),
        path == "ghc", strings.HasPrefix(path, "gio"), path == "lua", strings.HasPrefix(path, "mecab"),
        strings.HasPrefix(path, "node"), strings.HasPrefix(path, "ocaml"), strings.HasPrefix(path, "perl5"),
        path == "php", PYTHON23_RX.MatchString(path), strings.HasPrefix(path, "R"), strings.HasPrefix(path, "ruby"):
            return Mkpath
        default:
            return Link
        }
    })

    linkToCellar(cellarLocation, "Frameworks", func(path string) LinkStrategy {
        switch {
        case FRAMEWORKSVERSIONS_RX.MatchString(path):
            return Mkpath
        default:
            return Link
        }
    })

    makeRelativeSymlink(cellarLocation, k.LinkedKegRecord())
}

func GetKegFromFormula(f *Formula) *Keg {
    return &Keg{f}
}
