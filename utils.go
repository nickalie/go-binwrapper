package binwrapper

import (
	"runtime"
)

func osFilterObj(values []*Src) *Src {
	arches := []string{runtime.GOARCH}

	if runtime.GOARCH == "386" {
		arches = append(arches,"x86")
	} else if runtime.GOARCH == "amd64" {
		arches = append(arches,"x64")
	}

	platforms := []string{runtime.GOOS}

	if runtime.GOOS == "windows" {
		platforms = append(platforms,"win32")
	}

	for _, v := range values {
		if stringsContains(platforms, v.os) && stringsContains(arches, v.arch) {
			return v
		} else if stringsContains(platforms, v.os) && v.arch == "" {
			return v
		} else if stringsContains(arches, v.arch) && v.os == "" {
			return v
		} else if v.os == "" && v.arch == "" {
			return v
		}
	}

	return nil
}

func stringsContains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}

	return false
}
