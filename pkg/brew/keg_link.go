package brew

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func installInfo(path string) {
    exec.Command("/usr/bin/install-info", "--quiet", path, filepath.Join(filepath.Dir(path), "dir")).Run()
}

func makeRelativeSymlink(src string, dst string) {
    dstinfo, err := os.Lstat(dst)
    if err != nil {
        if !os.IsNotExist(err) {
            // not IsNotExist - file exists, some other error
            panic(err)
        }
    } else if dstinfo.Mode() & os.ModeSymlink == os.ModeSymlink { // if no err (file present) and is symlink
        if target, _ := os.Readlink(dst); src == target { return }
    }

    os.MkdirAll(filepath.Dir(dst), 0755)
    fmt.Printf("link\t%s -> %s\n", src, dst)
    rel_path, err := filepath.Rel(filepath.Dir(dst), src)
    if err != nil {
        panic(err)
    }
    err = os.Symlink(rel_path, dst)
    if err != nil {
        panic(err)
    }
}

func linkToCellar(path string, relativeDir string, linkMethodSelector func(string) LinkStrategy) {
    root := filepath.Join(path, relativeDir)
    if _, err := os.Lstat(root); os.IsNotExist(err) { return }

    err := filepath.Walk(root, func(src string, info os.FileInfo, err error) error {
        if src == root { return nil }

        rel_path, err := filepath.Rel(path, src)
        if err != nil {
            panic(err)
        }

        dst := filepath.Join(CurrentPlatform().HomebrewPrefix(), rel_path)

        if info.Mode().IsRegular() || info.Mode() & os.ModeSymlink == os.ModeSymlink {
            if filepath.Base(src) == ".DS_Store" { return filepath.SkipDir }
            if info.Mode() & os.ModeSymlink == os.ModeSymlink {
                resolved_path, err := os.Readlink(src)
                if err != nil { panic(err) }
                if resolved_path == dst { return filepath.SkipDir }
            }

            if info.Mode().IsRegular() {
                filekind := CheckFileType(src)
                isInPatchelfKeg := strings.HasPrefix(src, filepath.Join(CurrentPlatform().HomebrewCellar(), "patchelf"))
                isInGlibcKeg := strings.HasPrefix(src, filepath.Join(CurrentPlatform().HomebrewCellar(), "glibc"))
                if !isInPatchelfKeg && !isInGlibcKeg {
                    if filekind == "elf" {
                        fmt.Printf("patch\t%s\n", src)
                        RelocateElfLinkage(src, "@@HOMEBREW_PREFIX@@", CurrentPlatform().HomebrewPrefix())
                    }
                }
            }

            if strings.Contains(src, "site-packages/") {
                for _, ext := range PYC_EXTENSIONS {
                    if filepath.Ext(src) == ext { return filepath.SkipDir }
                }
            }

            switch linkMethodSelector(strings.TrimPrefix(rel_path, relativeDir + "/")) {
            case SkipFile:
                return filepath.SkipDir
            case Info:
                if filepath.Base(src) == "dir" { return nil }
                makeRelativeSymlink(src, dst)
                installInfo(dst)
            default:
                makeRelativeSymlink(src, dst)
            }
        } else if info.IsDir() {
            dstInfo, err := os.Lstat(dst)
            if (!os.IsNotExist(err)) && (dstInfo.Mode() & os.ModeSymlink != os.ModeSymlink) {
                if filepath.Ext(src) == ".app" { return filepath.SkipDir }

                switch linkMethodSelector(strings.TrimPrefix(rel_path, relativeDir + "/")) {
                case SkipDir:
                    return filepath.SkipDir
                case Mkpath:
                    os.MkdirAll(dst, 0755)
                default:
                    makeRelativeSymlink(src, dst)
                    return filepath.SkipDir
                }
            }
        }

        return nil
    })

    if err != nil {
        panic(err)
    }

    return
}
