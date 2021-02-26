package brew

import (
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    "github.com/alebcay/quickbrew/pkg/utils"
    "github.com/h2non/filetype"
)

func CheckFileType(filename string) string {
    result, err := filetype.MatchFile(filename)
    if err != nil {
        panic(err)
    }

    switch result.Extension {
    case "elf":
        return "elf"
    case "macho":
        return "macho"
    default:
        return "other"
    }
}

func SymlinkLdSo() {
    DYNAMIC_LINKERS := []string {
      "/lib64/ld-linux-x86-64.so.2",
      "/lib64/ld64.so.2",
      "/lib/ld-linux.so.3",
      "/lib/ld-linux.so.2",
      "/lib/ld-linux-aarch64.so.1",
      "/lib/ld-linux-armhf.so.3",
      "/system/bin/linker64",
      "/system/bin/linker",
    }

    brew_ld_so := filepath.Join(CurrentPlatform().HomebrewPrefix(), "lib", "ld.so")
    if !utils.IsFileReadable(brew_ld_so) {
        ld_so := filepath.Join(CurrentPlatform().HomebrewPrefix(), "opt", "glibc", "lib", "ld-linux-x86-64.so.2")
        if !utils.IsFileReadable(ld_so) {
            ld_so = ""
            for _, location := range DYNAMIC_LINKERS {
                if utils.IsFileExecutable(location) {
                    ld_so = location
                    break
                }
            }
            if ld_so == "" { panic("Unable to locate the system's dynamic linker") }
        }

        os.MkdirAll(filepath.Join(CurrentPlatform().HomebrewPrefix(), "lib"), 0755)
        err := os.Symlink(ld_so, brew_ld_so)
        if err != nil {
            panic(err)
        }
    }
}

func RelocateElfLinkage(filename string, oldPrefix string, newPrefix string) {
    patchelfBinary := filepath.Join(CurrentPlatform().HomebrewPrefix(), "opt", "patchelf", "bin", "patchelf")

    err := exec.Command(patchelfBinary, filename, "--print-rpath").Run()
    if err != nil {
        return
    }

    oldRpathBytes, err := exec.Command(patchelfBinary, filename, "--print-rpath").CombinedOutput()
    if err != nil {
        panic(err)
    }

    oldRpathPathString := strings.TrimSpace(string(oldRpathBytes))
    newRpathArray := []string{}

    oldRpathArray := strings.Split(oldRpathPathString, ":")
    for _, oldRpath := range oldRpathArray {
        newRpath := strings.Replace(oldRpath, oldPrefix, newPrefix, -1)
        if strings.HasPrefix(newRpath, newPrefix) || strings.HasPrefix(newRpath, "$ORIGIN") {
            newRpathArray = append(newRpathArray, newRpath)
        }
    }

    if !utils.StringSliceContains(filepath.Join(newPrefix, "lib"), newRpathArray) {
        newRpathArray = append(newRpathArray, filepath.Join(newPrefix, "lib"))
    }
    newRpathPathString := strings.Join(newRpathArray, ":")

    newInterpreterString := ""
    oldInterpreterBytes, err := exec.Command(patchelfBinary, filename, "--print-interpreter").CombinedOutput()
    oldInterpreterString := strings.TrimSpace(string(oldInterpreterBytes))

    if oldInterpreterString == "" {
        newInterpreterString = ""
    } else if utils.IsFileReadable(filepath.Join(CurrentPlatform().HomebrewPrefix(), "lib", "ld.so")) {
        newInterpreterString = filepath.Join(CurrentPlatform().HomebrewPrefix(), "lib", "ld.so")
    } else {
        newInterpreterString = strings.Replace(oldInterpreterString, oldPrefix, newPrefix, -1)
    }

    args := []string{filename, "--set-rpath", newRpathPathString}
    if newInterpreterString != "" { args = append(args, "--set-interpreter", newInterpreterString) }

    exec.Command(patchelfBinary, args...).CombinedOutput()
}
