package scanner

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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

type ScanResult struct {
	Selected     bool
	Path         string
	Size         uint64
	ModifiedDate time.Time
	Reason       Reason
}

func ScanForJunk() []*ScanResult {
	scanners := []scanner{
		&cacheScanner,
		&logScanner,
		&tempScanner,
		&deletedAppDataScanner,
		&leftOverUpdateDataScanner,
		&iphoneBackupScanner,
		&xcodeCacheScanner,
		&xcodeSimulatorScanner,
	}
	results := []*ScanResult{}
	var wg sync.WaitGroup
	ch := make(chan []*ScanResult)

	for _, s := range scanners {
		wg.Add(1)
		go func(s scanner) {
			defer wg.Done()
			ch <- s.scan()
		}(s)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for result := range ch {
		results = append(results, result...)
	}

	return results
}

type scanner interface {
	scan() []*ScanResult
}

var homeDir = func() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("failed to get user's home dir: %v", err))
	}
	return homeDir
}()

const (
	size10M = 10 * 1024 * 1024
	size1M  = 1024 * 1024
	size1K  = 1024
)

var (
	cacheScanner = pathScannerWithFilter{
		paths: []string{
			filepath.Join(homeDir, "Library", "Caches"),
			filepath.Join(homeDir, ".cache"),
		},
		filters: []filter{
			sizeFilter(size10M),
		},
		reason: Cache,
	}

	logScanner = pathScannerWithFilter{
		paths: []string{
			filepath.Join(homeDir, "Library", "Logs"),
		},
		filters: []filter{
			sizeFilter(size1M),
		},
		reason: Log,
	}

	tempScanner = pathScannerWithFilter{
		paths: []string{
			"/tmp",
			"/private/var/tmp/",
		},
		filters: []filter{
			sizeFilter(size1M),
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
			sizeFilter(size1K),
			prefixExclusionFilter("com.apple."), // Exclude apple built-in apps
			deletedAppDataFilter(),              // Only keep app data of deleted apps
		},
		reason: DeletedAppData,
	}

	leftOverUpdateDataScanner = pathScannerWithFilter{
		paths: []string{
			"/Library/Updates/",
			"/macOS Install Data/",
			filepath.Join(homeDir, "Library", "InstallerSandboxes"),
			filepath.Join(homeDir, "Library", "iTunes", "iPhone Software Updates"),
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
			sizeFilter(size1M),
		},
		reason: XcodeCache,
	}

	xcodeSimulatorScanner = pathScannerWithFilter{
		paths: []string{
			filepath.Join(homeDir, "Library", "Developer", "CoreSimulator", "Devices"),
		},
		filters: []filter{
			sizeFilter(size1M),
		},
		reason: XcodeSimulator,
	}
)

type filter func(*ScanResult) bool

type pathScannerWithFilter struct {
	paths   []string
	filters []filter
	reason  Reason
}

func sizeFilter(size uint64) filter {
	return func(s *ScanResult) bool {
		return s.Size >= size
	}
}

func prefixExclusionFilter(prefix string) filter {
	return func(s *ScanResult) bool {
		return !strings.HasPrefix(filepath.Base(s.Path), prefix)
	}
}

// Returns true if the ScannerResult belongs to a deleted app
func deletedAppDataFilter() filter {
	apps := getInstalledApps()
	return func(s *ScanResult) bool {
		for _, app := range apps {
			if strings.Contains(filepath.Base(s.Path), app) {
				return false
			}
		}
		return true
	}
}

func getInstalledApps() []string {
	var paths = []string{
		"/Applications",
		filepath.Join(homeDir, "Applications"),
	}
	var apps []string
	for _, root := range paths {
		filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() {
				return nil
			}
			if name, isApp := strings.CutSuffix(d.Name(), ".app"); isApp {
				apps = append(apps, name)
				// don't descend inside an .app bundle
				return filepath.SkipDir
			}
			return nil
		})
	}
	return apps
}

func (s *pathScannerWithFilter) scan() []*ScanResult {
	var results []*ScanResult
	var wg sync.WaitGroup
	ch := make(chan *ScanResult)

	for _, p := range s.paths {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			entries, err := os.ReadDir(p)
			if err != nil {
				log.Printf("failed to read dir %s: %v", p, err)
				return
			}

			for _, entry := range entries {
				if r := s.processEntry(filepath.Join(p, entry.Name())); r != nil {
					ch <- r
				}
			}
		}(p)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for result := range ch {
		results = append(results, result)
	}

	return results
}

func (s *pathScannerWithFilter) processEntry(path string) *ScanResult {
	info, err := os.Stat(path)
	if err != nil {
		log.Printf("failed to stat path %s: %v", path, err)
		return nil
	}

	var size uint64
	if info.IsDir() {
		size = fetchDirSize(path)
	} else if info.Mode().IsRegular() {
		size = uint64(info.Size())
	} else {
		return nil
	}

	r := ScanResult{
		Path:         path,
		Size:         size,
		ModifiedDate: info.ModTime(),
		Reason:       s.reason,
	}

	for _, filter := range s.filters {
		if !filter(&r) {
			return nil
		}
	}
	return &r
}

// Fetch dir size using `du`, which is faster than filepath.WalkDir()
func fetchDirSize(path string) uint64 {
	// -k: output in KB
	// -s: output the total size
	args := []string{"-k", "-s"}
	args = append(args, path)
	cmd := exec.Command("du", args...)
	output, err := cmd.Output()

	if err == nil {
		fields := strings.Fields(string(output))
		if len(fields) == 2 {
			size, _ := strconv.ParseUint(fields[0], 10, 64)
			return size * 1024
		}
	}
	return 0
}
