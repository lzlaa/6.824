package raft

// CondInstallSnapshot 如果lastIncludedIndex < rf.RWLog.SnapshotIndex，
// 则该快照已经过时立即返回false；否则返回true。
func (rf *Raft) CondInstallSnapshot(lastIncludedTerm int, lastIncludedIndex int, snapshot []byte) bool {
	rf.RWLog.mu.RLock()
	defer rf.RWLog.mu.RUnlock()

	// 不安装老旧的快照
	if lastIncludedIndex < rf.RWLog.SnapshotIndex {
		return false
	}
	return true
}

// Snapshot 更新rf.RWLog.SnapshotIndex、删除过时的日志并生成快照、持久化日志和快照
func (rf *Raft) Snapshot(index int, snapshot []byte) {

	// 防止rf.RWLog.SnapshotIndex的读写冲突
	rf.RWLog.mu.Lock()
	defer rf.RWLog.mu.Unlock()

	offset := index - rf.RWLog.SnapshotIndex

	// 丢弃过时的快照
	if offset < 0 || offset >= len(rf.Log) {
		return
	}
	rf.RWLog.SnapshotIndex = index

	// 删除原来的日志条目
	entries := make([]Entry, len(rf.Log[offset:]))
	copy(entries, rf.Log[offset:])
	rf.Log = entries

	// 首个日志条目用来存储snapshot
	rf.Log[0] = Entry{
		ApplyMsg: ApplyMsg{
			SnapshotValid: true,
			Snapshot:      snapshot,
			SnapshotTerm:  entries[0].Term,
			SnapshotIndex: entries[0].Index,
		},
		Term:  entries[0].Term,
		Index: entries[0].Index,
	}
	// 持久化日志和快照
	rf.persist()
}
