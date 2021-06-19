package raft

import (
	"fmt"
	"log"
	"sync"
)

//
// as each Raft peer becomes aware that successive RWLog entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed RWLog entry.
//
// in part 2D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
//

type ApplyMsg struct {
	CommandValid bool // true，该条目是日志条目
	Command      interface{}
	CommandIndex int

	// For 2D:
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
	SnapshotValid bool // true，该条目是快照

	// Replay为true，标识该日志条目是被重放的；否则是首次提交
	// 当raft宕机之后，持久化的快照和日志会被重放，此时的快照/日志的
	// Replay标记为true
	Replay bool
}

type Entry struct {
	ApplyMsg
	Term  int
	Index int
}

type RWLog struct {
	mu            sync.RWMutex
	Log           []Entry // 如果当前的server有快照，那么快照一定是第一个日志条目
	SnapshotIndex int     // 当前快照的LastIncludeIndex，所有日志条目索引的基准
}

func (l *RWLog) String() string {
	str := "["
	for i := 0; i < len(l.Log); i++ {
		str += fmt.Sprintf("{%d, %d},", l.Log[i].Index, l.Log[i].Term)
	}
	str += "]"
	return str
}

func (rf *Raft) applier() {

	for commitIndex := range rf.commitCh {
		if rf.killed() {
			close(rf.applyCh)
			return
		}

		// rf.lastApplied 是递增的所以不会重复执行同一个日志条目
		for commitIndex >= rf.lastApplied {

			// 防止产生对rf.RWLog的读写冲突
			rf.RWLog.mu.RLock()
			snapshotIndex := rf.RWLog.SnapshotIndex

			if rf.lastApplied < snapshotIndex {
				rf.lastApplied = snapshotIndex
				rf.RWLog.mu.RUnlock()
				continue
			}

			if rf.lastApplied-snapshotIndex >= len(rf.Log) {
				rf.RWLog.mu.RUnlock()
				break
			}

			entry := rf.Log[rf.lastApplied-snapshotIndex]
			rf.lastApplied++
			rf.RWLog.mu.RUnlock()

			if rf.gid != 0 {
				log.Printf("====== [%d] R%d#%d APPLY LA:%d, SI:%d-TE:%d, IN:%d ======\n", rf.CurrentTerm, rf.me, rf.gid, rf.lastApplied-1, snapshotIndex, entry.Term, entry.Index)
			}

			// 过滤掉占位符日志条目
			if !entry.CommandValid && !entry.SnapshotValid {
				continue
			}

			Debug(DCommit, "[%d] R%d APPLY LA:%d, SI:%d", rf.CurrentTerm, rf.me, rf.lastApplied-1, snapshotIndex)
			rf.applyCh <- entry.ApplyMsg
		}
	}
}
