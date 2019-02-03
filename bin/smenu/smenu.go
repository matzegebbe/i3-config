// smenu
//
// (c) 2013, 2014 by Sascha L. Teichmann <teichmann@intevation.de>
// This is Free Software under the terms of the GPLv3+.
// See LICENSE for details.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"
)

const (
	execMask    = 0111
	notFileMask = os.ModeDir | os.ModeNamedPipe | os.ModeDevice // Accept sym links.

	cacheName = ".smenu_cache"
)

var (
	errNothingSelected = errors.New("Nothing selected")
)

// entry is a data structure to store executable name and the number of calls.
type entry struct {
	name  string
	count int
}

// entries are slices of entries.
type entries []entry

// readCache reads entries from cache file.
func readCache(cacheFile string) (map[string]int, error) {
	file, err := os.Open(cacheFile)
	if err != nil {
		log.Printf("Cannot open cache file '%s'.\n", cacheFile)
		return nil, err
	}
	defer file.Close()
	execs := make(map[string]int)
	lineNo := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			log.Printf("not enough parts in line %d\n", lineNo)
			continue
		}
		var count int64
		if count, err = strconv.ParseInt(parts[0], 10, 32); err != nil {
			log.Printf("Cannot parse '%s' in line %d as an integer.\n", parts[0], lineNo)
		} else {
			execs[parts[1]] = int(count)
		}
	}
	return execs, scanner.Err()
}

// Len() to fulfill sort.Interface.
func (ntrs entries) Len() int {
	return len(ntrs)
}

// Swap() to fulfill sort.Interface.
func (ntrs entries) Swap(i, j int) {
	ntrs[i], ntrs[j] = ntrs[j], ntrs[i]
}

// Less() to fulfill sort.Interface.
func (ntrs entries) Less(i, j int) (result bool) {
	ic, jc := ntrs[i].count, ntrs[j].count
	switch {
	case ic > jc:
		return true
	case ic < jc:
		return false
	}
	return ntrs[i].name < ntrs[j].name
}

// writeToCache writes a slice of entries to cache file. To do this
// in a 'atomic' way a temporary file with the new content
// is created at first and renamed afterwards.
func (ntrs entries) writeToCache(cacheFile string) error {
	i := 0
	pid := os.Getpid()
	var tmpFile string
	for {
		tmpFile = fmt.Sprintf("%s-%d.%d", cacheFile, pid, i)
		if _, err := os.Stat(tmpFile); err != nil {
			if os.IsNotExist(err) {
				break
			}
			return err
		}
		i++
	}
	defer os.Remove(tmpFile)
	err := ntrs.createCacheFile(tmpFile)
	if err == nil {
		return os.Rename(tmpFile, cacheFile)
	}
	return err

}

// createCacheFile writes a fresh cache file.
func (ntrs entries) createCacheFile(cacheFile string) (err error) {

	var file *os.File

	file, err = os.OpenFile(cacheFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("Cannot open cache file '%s' for writing.\n", cacheFile)
		return err
	}

	out := bufio.NewWriter(file)

	for _, entry := range ntrs {
		if _, err = fmt.Fprintf(out, "%d %s\n", entry.count, entry.name); err != nil {
			log.Printf("Error while writing to cache file '%s'.\n", cacheFile)
			file.Close()
			return
		}
	}
	err = out.Flush()
	file.Close()
	return
}

// sortEntries sorts the entries of the cache. Higher count numbers
// lead to positions at the beginning of the slice.
// entries with same counter values are ordered by their names.
func sortEntries(execs map[string]int) entries {
	ntrs := make(entries, len(execs))
	i := 0
	for name, count := range execs {
		ntrs[i].name = name
		ntrs[i].count = count
		i++
	}
	sort.Sort(ntrs)
	return ntrs
}

// takeOverCounters merges old counters to new generated entries. This is called
// if have to refresh the cache.
func takeOverCounters(oldExecs, newExecs map[string]int) {
	for name := range newExecs {
		if count, found := oldExecs[name]; found {
			newExecs[name] = count
		}
	}
	for name, count := range oldExecs {
		if _, found := newExecs[name]; !found { // Has not survived, yet.
			// Check if its a composition like "firefox -P -no-remote"
			parts := strings.SplitN(name, " ", 2)
			if len(parts) > 1 { // Its a composition.
				if _, found := newExecs[parts[0]]; found { // "firefox" found
					newExecs[name] = count // take over composition.
				}
			}
		}
	}
}

// callDmenu calls dmenu by piping the executable names in a retrieve the selected one.
// Returns the selected command, the exit code of dmenu and an error
// (nil if no error occured).
// If a command was successfully selected the counter of the program
// is increased by one.
func callDmenu(dmenu string, args []string, execs map[string]int) (command string, exit int, err error) {
	ntrs := sortEntries(execs)

	cmd := exec.Command(dmenu, args...)

	var stdin io.WriteCloser
	var stdout io.ReadCloser

	if stdin, err = cmd.StdinPipe(); err != nil {
		return
	}

	if stdout, err = cmd.StdoutPipe(); err != nil {
		return
	}

	if err = cmd.Start(); err != nil {
		return
	}

	done := make(chan struct{})

	go func() {
		for i := range ntrs {
			if _, err2 := fmt.Fprintf(
				stdin, "%s\n", ntrs[i].name); err2 != nil {
				log.Printf("pipe to dmenu failed: %s\n", err2)
				break
			}
		}
		stdin.Close()
		done <- struct{}{}
	}()

	lines := make([]string, 0, 1)

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	<-done
	if err = cmd.Wait(); err != nil {
		if err2, ok := err.(*exec.ExitError); ok {
			if ws, ok := err2.ProcessState.Sys().(syscall.WaitStatus); ok {
				exit = ws.ExitStatus()
			}
		}
		return
	}

	if len(lines) == 0 {
		err = errNothingSelected
	} else {
		command = lines[0]
		execs[command]++
	}
	return
}

// shellSplit splits a given string into substrings which are separated
// by Unicode whitespace. strings.Fields does not have
// a notion of quoted strings like "Hello World" so its not
// used.
func shellSplit(s string) []string {

	const (
		ORDINARY = iota
		QUOTED
		WHITESPACE
	)

	result := []string{}

	tmp := make([]rune, 0, len(s))

	state := ORDINARY
	var sep rune

	for _, c := range s {
		switch state {
		case ORDINARY:
			if unicode.IsSpace(c) {
				if len(tmp) > 0 {
					result = append(result, string(tmp))
					tmp = tmp[0:0]
				}
				state = WHITESPACE
			} else if c == '\'' || c == '"' {
				sep = c
				state = QUOTED
			} else {
				tmp = append(tmp, c)
			}
		case QUOTED:
			if c != sep {
				tmp = append(tmp, c)
			} else {
				state = ORDINARY
			}
		case WHITESPACE:
			if !unicode.IsSpace(c) {
				if c == '\'' || c == '"' {
					sep = c
					state = QUOTED
				} else {
					tmp = append(tmp, c)
					state = ORDINARY
				}
			}
		}
	}

	if len(tmp) > 0 {
		result = append(result, string(tmp))
	}

	return result
}

// modifiedAfter checks if the modification time of one of the PATH
// entries was after the reference time. In this case we
// have to refresh the cache.
func modifiedAfter(reference time.Time, dirs []string) bool {
	for _, dir := range dirs {
		if info, err := os.Stat(dir); err != nil {
			log.Printf("Cannot stat dir '%s'.\n", dir)
		} else if info.ModTime().After(reference) {
			return true
		}
	}
	return false
}

// findExecutables scans all PATH dirs for executables and stores their
// name in a map with zero calls.
func findExecutables(dirs []string) map[string]int {
	execs := make(map[string]int)

	for _, dir := range dirs {
		infos, err := ioutil.ReadDir(dir)
		if err != nil {
			log.Printf("Failed reading dir '%s'.\n", dir)
			continue
		}
		for i := range infos {
			name, mode := infos[i].Name(), infos[i].Mode()
			if (mode&notFileMask == 0) &&
				(mode&execMask != 0) &&
				!strings.HasPrefix(name, ".") {
				execs[name] = 0
			}
		}
	}
	return execs
}

// execute replaces the current program image with a given
// new one. Should only return in case of an error.
func execute(cmd string) (err error) {
	parts := shellSplit(cmd)

	if len(parts) < 1 {
		return errNothingSelected
	}

	var path string
	if path, err = exec.LookPath(parts[0]); err != nil {
		return
	}
	return syscall.Exec(path, parts, os.Environ())
}

func main() {

	path := os.Getenv("PATH")
	if path == "" {
		log.Fatal("No PATH variable")
	}

	home := os.Getenv("HOME")
	if home == "" {
		log.Fatal("No HOME variable")
	}

	var dmenu string

	flag.StringVar(&dmenu, "dmenu", "dmenu", "dmenu executable")
	flag.StringVar(&dmenu, "d", "dmenu", "dmenu executable (shorthand)")

	flag.Parse()

	callD := func(execs map[string]int) (string, int, error) {
		return callDmenu(dmenu, flag.Args(), execs)
	}

	dirs := strings.Split(path, string(os.PathListSeparator))

	cacheFile := filepath.Join(home, cacheName)

	var selection string
	var derr error
	var exit int

	cacheInfo, err := os.Stat(cacheFile)
	if err != nil { // Cannot stat -> create new cache file.
		if !os.IsNotExist(err) {
			log.Fatalf("Stat failed for '%s'.", cacheFile)
		}
		execs := findExecutables(dirs)
		selection, exit, derr = callD(execs)
		ntrs := sortEntries(execs)
		err = ntrs.createCacheFile(cacheFile)
	} else if modifiedAfter(cacheInfo.ModTime(), dirs) {
		// Re-generate the cache file.
		var newExecs, oldExecs map[string]int
		newExecs = findExecutables(dirs)
		oldExecs, err = readCache(cacheFile)
		if err == nil {
			takeOverCounters(oldExecs, newExecs)
		}
		selection, exit, derr = callD(newExecs)
		ntrs := sortEntries(newExecs)
		err = ntrs.writeToCache(cacheFile)
	} else {
		// Cache is new enough.
		var execs map[string]int
		if execs, err = readCache(cacheFile); err == nil {
			if selection, exit, derr = callD(execs); derr == nil {
				ntrs := sortEntries(execs)
				err = ntrs.writeToCache(cacheFile)
			}
		}
	}
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	if derr != nil {
		if exit == 0 { // Log only if its some internal error.
			exit = 1
			log.Printf("error: %s\n", derr)
		}
		os.Exit(exit)
	}

	if err = execute(selection); err != nil {
		log.Fatalf("error: %s", err)
	}
}
