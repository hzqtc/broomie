package scanner

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Reason string

const (
	Cache              Reason = "Cache"
	Log                Reason = "Log"
	Temp               Reason = "Temporary Files"
	DeletedAppData     Reason = "Deleted App Data"
	LeftOverUpdateData Reason = "Leftover Update Data"
	IphoneBackup       Reason = "iPhone Backup"
	XcodeCache         Reason = "Xcode Cache"
	XcodeSimulator     Reason = "Xcode Simulator"
)

type ScannerResult struct {
	Path         string
	SizeKbs      int64
	ModifiedDate time.Time
	Reason       Reason
}

func ScanForJunk() []ScannerResult {
	scanners := []scanner{
		&cacheScanner,
		&logScanner,
		&tempScanner,
		// &deletedAppDataScanner,
		&leftOverUpdateDataScanner,
		&iphoneBackupScanner,
		&xcodeCacheScanner,
		&xcodeSimulatorScanner,
	}
	results := []ScannerResult{}
	for _, s := range scanners {
		results = append(results, s.scan()...)
	}
	return results
}

type scanner interface {
	scan() []ScannerResult
}

var homeDir = func() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("failed to get user's home dir: %v", err))
	}
	return homeDir
}()

var (
	cacheScanner = pathScannerWithFilter{
		paths: []string{
			filepath.Join(homeDir, "Library", "Caches"),
			filepath.Join(homeDir, ".cache"),
		},
		filters: []filter{
			sizeFilter(10 * 1024), // 10M+
		},
		reason: Cache,
	}

	logScanner = pathScannerWithFilter{
		paths: []string{
			filepath.Join(homeDir, "Library", "Logs"),
		},
		filters: []filter{
			sizeFilter(10 * 1024), // 10M+
		},
		reason: Log,
	}

	tempScanner = pathScannerWithFilter{
		paths: []string{
			"/tmp",
			"/private/var/tmp/",
		},
		filters: []filter{
			sizeFilter(10 * 1024), // 10M+
		},
		reason: Temp,
	}

	deletedAppDataScanner = pathScannerWithFilter{
		paths: []string{
			filepath.Join(homeDir, "Library", "Application Support"),
			filepath.Join(homeDir, "Library", "Containers"),
			filepath.Join(homeDir, "Library", "Saved Application State"),
			filepath.Join(homeDir, "Library", "Preferences"),
		},
		filters: []filter{
			// TODO: filter by deleted app
		},
		reason: DeletedAppData,
	}

	leftOverUpdateDataScanner = pathScannerWithFilter{
		paths: []string{
			"/Library/Updates/",
			"/macOS Install Data/",
			filepath.Join(homeDir, "Library", "InstallerSandboxes"),
			filepath.Join(homeDir, "iTunes", "iPhone Software Updates"),
		},
		reason: LeftOverUpdateData,
	}

	iphoneBackupScanner = pathScannerWithFilter{
		paths: []string{
			filepath.Join(homeDir, "Library", "Application Support", "MobileSync", "Backup"),
		},
		reason: IphoneBackup,
	}

	xcodeCacheScanner = pathScannerWithFilter{
		paths: []string{
			filepath.Join(homeDir, "Library", "Developer", "Xcode", "DerivedData"),
			filepath.Join(homeDir, "Library", "Developer", "Xcode", "DocumentationCache"),
			filepath.Join(homeDir, "Library", "Developer", "Xcode", "DocumentationIndex"),
		},
		filters: []filter{
			sizeFilter(1 * 1024), // 1M+
		},
		reason: XcodeCache,
	}

	xcodeSimulatorScanner = pathScannerWithFilter{
		paths: []string{
			filepath.Join(homeDir, "Library", "Developer", "CoreSimulator", "Devices"),
		},
		filters: []filter{
			sizeFilter(1 * 1024), // 1M+
		},
		reason: XcodeSimulator,
	}
)

type filter func(ScannerResult) bool

type pathScannerWithFilter struct {
	paths   []string
	filters []filter
	reason  Reason
}

func sizeFilter(size int64) filter {
	return func(s ScannerResult) bool {
		return s.SizeKbs >= size
	}
}

func (s *pathScannerWithFilter) scan() []ScannerResult {
	var results []ScannerResult

	for _, p := range s.paths {
		entries, err := os.ReadDir(p)
		if err != nil {
			log.Printf("failed to read dir %s: %v", p, err)
			continue
		}

		for _, entry := range entries {
			childPath := filepath.Join(p, entry.Name())
			info, err := os.Stat(childPath)
			if err != nil {
				log.Printf("failed to stat path %s: %v", childPath, err)
				continue
			}
			if !info.Mode().IsRegular() && !info.Mode().IsDir() {
				continue
			}

			r := ScannerResult{
				Path:         childPath,
				SizeKbs:      fetchDirSize(childPath),
				ModifiedDate: info.ModTime(),
				Reason:       s.reason,
			}

			passedAllFilters := true
			for _, filter := range s.filters {
				if !filter(r) {
					passedAllFilters = false
					break
				}
			}

			if passedAllFilters {
				results = append(results, r)
			}
		}
	}

	return results
}

func fetchDirSize(path string) int64 {
	// -k: output in KB
	// -s: output the total size
	args := []string{"-k", "-s"}
	args = append(args, path)
	cmd := exec.Command("du", args...)
	output, err := cmd.Output()

	if err == nil {
		fields := strings.Fields(string(output))
		if len(fields) == 2 {
			size, _ := strconv.ParseInt(fields[0], 10, 64)
			return size
		}
	}
	return 0
}
