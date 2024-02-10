package watcher

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/jithin-kg/gomon/constants"
	"github.com/jithin-kg/gomon/internal/utils"
)

type Watcher struct {
	watcher       *fsnotify.Watcher
	Paths         []string
	Ignore        []string //patterns to ignore
	OnChange      func()   // call back function to be called on file change
	buildDelay    time.Duration
	lastEvent     time.Time
	eventMutex    sync.Mutex
	buildTimer    *time.Timer
	eventQueue    []fsnotify.Event
	fileChecksums map[string]string
	ConfigChange  chan struct{}
}

// new creates a new watcher
func New(paths []string, ignore []string, onChange func()) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()

	if err != nil {
		return nil, err
	}
	return &Watcher{
		watcher:       w,
		Paths:         paths,
		Ignore:        ignore,
		OnChange:      onChange,
		buildDelay:    1 * time.Second,
		eventQueue:    []fsnotify.Event{},
		fileChecksums: make(map[string]string),
		ConfigChange:  make(chan struct{}),
	}, nil
}
func (w *Watcher) handleEvent(event fsnotify.Event) {
	w.eventMutex.Lock()
	defer w.eventMutex.Unlock()

	//queue the event
	w.eventQueue = append(w.eventQueue, event)
	now := time.Now()

	if w.shouldIgnore(event.Name) {
		return
	}

	currentChecksum, err := fileCheckSum(event.Name)
	if err != nil {
		log.Printf("Error calculating fileChecksum %s %v\n", event.Name, err)
		return
	}

	// if checksum changed
	if lastChecksum, ok := w.fileChecksums[event.Name]; ok && lastChecksum == currentChecksum {
		// checksum hastn changed, no need to rebuild
		return
	} else {
		// checksum changed
		color.Blue("chesum changed lastChecksum: %s, currentChecksum: %s", lastChecksum, currentChecksum)
		w.fileChecksums[event.Name] = currentChecksum
	}
	if w.buildTimer != nil {
		w.buildTimer.Stop()
	}
	w.buildTimer = time.AfterFunc(w.buildDelay, func() {
		w.eventMutex.Lock()
		defer w.eventMutex.Unlock()

		// Check if enough time has passed since the last event to consider it "quiet".
		if time.Since(w.lastEvent) < w.buildDelay {
			return // Too soon to trigger a build, likely a new event arrived.
		}

		// Now, it's safe to assume we're in a quiet period.
		if len(w.eventQueue) > 0 {
			color.Magenta("Triggering build due to quiet period after last event.")
			w.triggerBuild(event.Name)
			w.flushEvents() // Now it makes sense to clear the queue after confirming a build.
		}
	})

	w.lastEvent = now

}

func fileCheckSum(filename string) (string, error) {
	contents, err := os.ReadFile(filename)

	if err != nil {
		return "", err
	}
	hasher := sha256.New()
	_, err = hasher.Write(contents)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
func (w *Watcher) flushEvents() {
	// clear the event queue
	w.eventQueue = []fsnotify.Event{}
}
func (w *Watcher) triggerBuild(eventName string) {

	// ensure this runs in a non blocking manner
	go func() {
		if filepath.Base(eventName) == constants.ConfigFileName {
			color.Yellow("gomon.json changed")
			w.ConfigChange <- struct{}{}
		} else {
			if w.OnChange != nil {
				w.OnChange()
			}
		}

	}()
}

// starts begins watching file system for changes
func (w *Watcher) Start() {

	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if w.shouldIgnore(event.Name) {
					continue
				}
				fmt.Printf("Event: %s \n", event)
				// Here you can add logic to handle the event, like rebuilding
				// if w.OnChange != nil {
				// 	w.OnChange() //call the callback function
				// }
				w.handleEvent(event)
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}

				log.Println("Error:", err)
			}

		}
	}()
	for _, path := range w.Paths {
		if err := w.addWatchPaths(path); err != nil {
			log.Fatalf("Failed to watch directory %s: %v", path, err)
		}
	}

}

func (w *Watcher) addWatchPaths(startPath string) error {
	// todo Walk is less efficient than WalkDir, introduced in Go 1.16, which avoids calling os.Lstat on every visited file or directory.
	filepath.Walk(startPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// before adding the directory to the watcher check if it should be ignored
			if w.shouldIgnore(path) {
				return filepath.SkipDir
			}
			err := w.watcher.Add(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return nil
}

func (w *Watcher) shouldIgnore(path string) bool {
	projectRoot, err := utils.GetProjectRoot()
	if err != nil {
		log.Fatalf("Unable to find project root: %v", err)
	}

	// Ensure path is absolute for reliable relative path calculation
	absPath := path
	if !filepath.IsAbs(path) {
		absPath = filepath.Join(projectRoot, path)
	}

	relPath, err := filepath.Rel(projectRoot, absPath)
	if err != nil {
		log.Printf("Error converting to relative path: %s, error: %v", absPath, err)
		return false
	}

	// Normalize the relative path
	relPath = utils.CleanPath(relPath)

	for _, pattern := range w.Ignore {
		normalizedPattern := utils.CleanPath(pattern)

		if match, _ := filepath.Match(normalizedPattern, relPath); match {
			color.Cyan("Ignoring path: %s based on pattern: %s", relPath, normalizedPattern)
			return true
		}
	}

	return false
}

/**
 *  Iterates through each pattern in the Watcher's Ignore list.
 * 	These patterns are strings that can include wildcards
 *	like *.log for ignoring all files with a .log extension, or temp/* for
 *	ignoring all files in a temp directory.
 **/

// func (w *Watcher) shouldIgnore(path string) bool {
// 	for _, pattern := range w.Ignore {
// 		match, err := filepath.Match(pattern, path)
// 		color.Yellow("path-based -> pattern: %s, path: %s, match: %s, err : %v", pattern, path, match, err)
// 		if err != nil {
// 			color.Red("Error matching pattern %s with path %s, error %v\n", pattern, path, err)
// 			continue
// 		}
// 		if match {
// 			color.Green("path-based -> match")
// 			return true
// 		}

// 		// if strings.Contains(path, "tmp/main") {
// 		// 	return true
// 		// }
// 		// if the entire path doesn't match, we take the base name of the path
// 		// (ie it will get file name with its extension minus any directory components)
// 		// This ensure that the file names can be matched regardless of their directory
// 		// eg: if you want to ignore all files named `config.json` regardless of where they are in your project
// 		base := filepath.Base(path)
// 		match, err = filepath.Match(pattern, base)
// 		color.Yellow("base-based -> pattern: %s, path: %s, match : %s, err : %v", pattern, base, match, err)
// 		if err != nil {
// 			log.Printf("Error matching pattern %s with base %s, error %v\n", pattern, base, err)
// 			continue
// 		}
// 		if match {
// 			color.Green("base-based -> match")
// 			return true
// 		}
// 	}
// 	// Explicitly ignore build artifacts or temporary files
// 	// if strings.Contains(path, "/tmp/") || strings.HasSuffix(path, "tmp/main") {
// 	// 	return true
// 	// }
// 	return false
// }

// close stops the watcher
func (w *Watcher) Close() error {
	return w.watcher.Close()
}
