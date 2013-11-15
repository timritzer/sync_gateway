package channels

import (
	"fmt"
	"sort"
)

type LogEntry struct {
	Sequence uint64 // Sequence number
	DocID    string // Document ID
	RevID    string // Revision ID
	Flags    uint8  // Deleted/Removed/Hidden flags
}

// Bits in LogEntry.Flags
const (
	Deleted = 1 << iota
	Removed
	Hidden

	kMaxFlag = (1 << iota) - 1
)

// A sequential log of document revisions added to a channel, used to generate _changes feeds.
// The log is sorted by order revisions were added; this is mostly but not always by sequence.
type ChangeLog struct {
	Since   uint64      // Sequence this log is valid _after_, i.e. max sequence not in the log
	Entries []*LogEntry // Entries in order they were added (not sequence order!)
}

func (entry *LogEntry) assertValid() {
	if entry.Sequence == 0 || entry.DocID == "" || entry.RevID == "" || entry.Flags > kMaxFlag {
		panic(fmt.Sprintf("Invalid entry: %+v", entry))
	}
}

// Adds a new entry, always at the end of the log.
func (cp *ChangeLog) Add(newEntry LogEntry) {
	newEntry.assertValid()
	if len(cp.Entries) == 0 || newEntry.Sequence == cp.Since {
		cp.Since = newEntry.Sequence - 1
	}
	cp.Entries = append(cp.Entries, &newEntry)
}

// Removes the oldest entries to limit the log's length to `maxLength`.
func (cp *ChangeLog) TruncateTo(maxLength int) int {
	if remove := len(cp.Entries) - maxLength; remove > 0 {
		// Set Since to the max of the sequences being removed:
		for _, entry := range cp.Entries[0:remove] {
			if entry.Sequence > cp.Since {
				cp.Since = entry.Sequence
			}
		}
		cp.Entries = cp.Entries[remove:]
		return remove
	}
	return 0
}

// Returns a slice of all entries added after the one with sequence number 'after'.
// (They're not guaranteed to have higher sequence numbers; sequences may be added out of order.)
func (cp *ChangeLog) EntriesAfter(after uint64) []*LogEntry {
	entries := cp.Entries
	for i, entry := range entries {
		if entry.Sequence == after {
			return entries[i+1:]
		}
	}
	return entries
}

// Filters the log to only the entries added after the one with sequence number 'after.
func (cp *ChangeLog) FilterAfter(after uint64) {
	cp.Entries = cp.EntriesAfter(after)
	if after > cp.Since {
		cp.Since = after
	}
}

// Sorts the entries by increasing sequence.
func (c *ChangeLog) Sort() {
	sort.Sort(c)
}

func (c *ChangeLog) Len() int { // part of sort.Interface
	return len(c.Entries)
}

func (c *ChangeLog) Less(i, j int) bool { // part of sort.Interface
	return c.Entries[i].Sequence < c.Entries[j].Sequence
}

func (c *ChangeLog) Swap(i, j int) { // part of sort.Interface
	c.Entries[i], c.Entries[j] = c.Entries[j], c.Entries[i]
}
