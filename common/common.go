package common

import "sync"

const (
	OK             = "OK"
	ErrNoKey       = "ErrNoKey"
	ErrWrongLeader = "ErrWrongLeader"
)

type Err string

type signal struct{}

type OpReply struct {
	op    Op
	reply interface{} // op的执行结果
}

// 记录每个Op，及其执行结果
// 但server执行完Op之后，关闭s以通知等待当前Op的Client
type SignalWithOpReply struct {
	s    chan signal
	wait int // 等待s的clerk协程数目
	OpReply
}

type Index int

// OpReplays记录server提交给raft的op及其reply，所有的clerk协程在向server发送op后
// 阻塞在与该op绑定的channel上，server在执行完op之后通过关闭channel唤醒阻塞的clerk协程。
type OpReplys struct {
	// 如果table[index] != nil，说明存在clerk在等待index处的op
	table map[Index]*SignalWithOpReply
	mu    *sync.Mutex
	done  chan signal
}

func NewOpReplays() OpReplys {
	return OpReplys{
		table: make(map[Index]*SignalWithOpReply),
		mu:    new(sync.Mutex),
		done:  make(chan signal),
	}
}

// 阻塞在i处的Op对应的signal channel上，并在被唤醒后获取到Op以及对应的Result
// 当Op被执行之后applier协程通过关闭该signal channel通知阻塞的工作协程
// 如果调用协程是第一个调用者，首先初始化该signal
func (or OpReplys) Wait(i Index) *SignalWithOpReply {
	or.mu.Lock()
	ret, ok := or.table[i]
	if ok == false {
		or.table[i] = &SignalWithOpReply{
			s: make(chan signal),
		}
		ret = or.table[i]
	}
	or.mu.Unlock()
	select {
	case <-ret.s: // Op处理完毕
	case <-or.done: // Server宕机
	}
	return ret
}

// clerk协程等待reqOp的完成。
// 因为存在raft leadership的转变，多个clerk可能会等待同一个index的op；当server执行完
// index对应的op后，会通知所有等待的clerk协程；clerk协程会根据result.op == reqOp判断
// 完成的op是不是自己提交的op；如果不是就表明发生了leadership的转变
//
func (or OpReplys) WaitAndMatch(index int, reqOp Op) (interface{}, Err) {

	result := or.Wait(Index(index))
	resOp := result.op // raft所提交日志中的ApplyMsg

	// leadership 发生变更，或者原先提交的请求被覆盖时index处的 resOp != reqOp
	if resOp != reqOp {
		return nil, ErrWrongLeader
	}
	or.Delete(Index(index)) // 为了节约内存及时删除缓存
	return result.reply, ""
}

// server 调用SetAndBroadcast唤醒所有等待i对应Op的工作协程(clerk)。
// 如果目前还没有clerk等待该Op，且wake为false则直接返回；
// 如果wake为true，说明存在clerk会等待该op的完成，因此需要创建并插入对应的OpResult，
// 并且closer(ret.s)告知等待的clerk，该op已经完成
func (or OpReplys) SetAndBroadcast(i Index, op Op, re interface{}, wake bool) {
	or.mu.Lock()
	defer or.mu.Unlock()

	ret, ok := or.table[i]
	if !ok { // 没有等待该Op的工作协程，直接返回
		if !wake {
			return
		}
		or.table[i] = new(SignalWithOpReply)
		or.table[i].s = make(chan signal)
		ret = or.table[i]
	}

	ret.op = op
	ret.reply = re // 设置op的执行结果
	close(ret.s)   // 唤醒等待协程
}

func (or OpReplys) Delete(i Index) {
	or.mu.Lock()
	defer or.mu.Unlock()

	delete(or.table, i)
}

func (or OpReplys) Destory() {
	or.mu.Lock()
	defer or.mu.Unlock()
	close(or.done) // 通知所有等待的clerk返回
}

// 每一个clerk的Op都有一个唯一标识符Identifier，相同的Op其Identifier也相同
// ClerkID	标识Client ID
// Seq		标识Client下一个Op的序号
type Identifier struct {
	ClerkID int
	Seq     int
}

// ITable是Server记录目前各个Client待提交Op的Seq的哈希表：
// 关键字是Client的ClerkID
// 值是该Client目前待提交的Seq
// 如果遇到OpSeq较小的Op就可以判定该Op是被重复提交的，因此不会被执行。
type ITable struct {
	SeqTable   map[int]int         // 记录每一个clerk的待提交的Op Sequence number
	ReplyTable map[int]interface{} // 记录每一个clerk的上一个Op的执行结果，以便等待同一个op的clerk协程能够立即返回
}

func NewITable() ITable {
	return ITable{
		SeqTable:   make(map[int]int),
		ReplyTable: make(map[int]interface{}),
	}
}

// 返回clerkID对应clerk的可用的Op标识符
func (itable ITable) GetIdentifier(clerkID int) Identifier {

	return Identifier{
		ClerkID: clerkID,
		Seq:     itable.SeqTable[clerkID],
	}
}

// 返回clerkID对应clerk的上一个Op的执行结果
func (itable *ITable) GetCacheReply(clerkID int) (reply interface{}) {
	return itable.ReplyTable[clerkID]
}

// 更新clerkID对应clerk的下一个Op标识符
func (itable *ITable) UpdateIdentifier(clerkID int, seq int, reply interface{}) {
	itable.SeqTable[clerkID] = seq
	itable.ReplyTable[clerkID] = reply
}

// 如果i标识的Op已经被执行过了，Executed返回true
func (itable *ITable) Executed(i Identifier) bool {
	return i.Seq < itable.SeqTable[i.ClerkID]
}

type Op struct {
	// Your definitions here.
	// Field names must start with capital letters,
	// otherwise RPC will break.
	ServerID int // 打包该Op的Server
	Kind     string
	Key      string
	Value    string
	ID       Identifier
}
