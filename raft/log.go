// Copyright 2015 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package raft


import pb "github.com/pingcap-incubator/tinykv/proto/pkg/eraftpb"

// RaftLog manage the log entries, its struct look like:
// RaftLog管理日志条目，其结构如下所示：
//
//	 快照/第一个…........已应用….......已提交….....已稳定….....最后一个
//  snapshot/first.....applied....committed....stabled.....last
//  --------|------------------------------------------------|
//                            log entries(日志条目)
//
// for simplify the RaftLog implement should manage all log entries
// that not truncated
//为了简化，RaftLog工具应该管理所有日志条目
//那不是被截断的
type RaftLog struct {
	// storage contains all stable entries since the last snapshot.（储存包含上次快照以来的所有稳定条目）
	storage Storage

	// committed is the highest log position that is known to be in（committed是已知处于的最高日志位置）
	// stable storage on a quorum of nodes.（节点仲裁上的稳定存储。）
	committed uint64

	// applied is the highest log position that the application has（applied是应用程序具有的最高日志位置）
	// been instructed to apply to its state machine.（已指示应用于其状态机。）
	// Invariant: applied <= committed（不变量：已应用<=已提交）
	applied uint64

	// log entries with index <= stabled are persisted to storage.
	// It is used to record the logs that are not persisted by storage yet.
	// Everytime handling `Ready`, the unstabled logs will be included.
	//索引为<=稳定的日志项将持久化到存储器中。
	//它用于记录尚未由存储持久化的日志
	//每次处理'Ready'，都会包含不稳定的日志。 
	stabled uint64

	// all entries that have not yet compact.（所有尚未压缩的条目。）
	entries []pb.Entry

	// the incoming unstable snapshot, if any.（传入的不稳定快照（如果有））
	// (Used in 2C)
	pendingSnapshot *pb.Snapshot

	// Your Data Here (2A).


}

// newLog returns log using the given storage. It recovers the log
// to the state that it just commits and applies the latest snapshot.
//newLog使用给定的存储返回日志。它将日志恢复到它刚刚提交的状态，并应用最新的快照。 
func newLog(storage Storage) *RaftLog {
	// Your Code Here (2A).
	return nil
}

// We need to compact the log entries in some point of time like
// storage compact stabled log entries prevent the log entries
// grow unlimitedly in memory
//我们需要在某个时间点压缩日志项，如存储压缩稳定的日志项防止日志项在内存中无限增长
func (l *RaftLog) maybeCompact() {
	// Your Code Here (2C).
}

// unstableEntries return all the unstable entries（返回所有不稳定条目）
func (l *RaftLog) unstableEntries() []pb.Entry {
	// Your Code Here (2A).
	return nil
}

// nextEnts returns all the committed but not applied entries（返回已提交但为应用的条目）
func (l *RaftLog) nextEnts() (ents []pb.Entry) {
	// Your Code Here (2A).
	return nil
}

// LastIndex return the last index of the log entries（返回日志项的最后一个索引）
func (l *RaftLog) LastIndex() uint64 {
	// Your Code Here (2A).
	return 0
}

// Term return the term of the entry in the given index（返回给定索引中条目的期限）
func (l *RaftLog) Term(i uint64) (uint64, error) {
	// Your Code Here (2A).
	return 0, nil
}
