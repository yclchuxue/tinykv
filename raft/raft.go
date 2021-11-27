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

import (
	"errors"
	"math/rand"
	"time"

	pb "github.com/pingcap-incubator/tinykv/proto/pkg/eraftpb"
)

// None is a placeholder node ID used when there is no leader.（没有leader时的占位符）
const None uint64 = 0

// StateType represents the role of a node in a cluster.（）
type StateType uint64

const (
	StateFollower StateType = iota
	StateCandidate
	StateLeader
)

var stmap = [...]string{
	"StateFollower",
	"StateCandidate",
	"StateLeader",
}

func (st StateType) String() string {
	return stmap[uint64(st)]
}

// ErrProposalDropped is returned when the proposal is ignored by some cases,
// so that the proposer can be notified and fail fast.
var ErrProposalDropped = errors.New("raft proposal dropped")

// Config contains the parameters to start a raft.
// Config 结构体包含启动raft的参数
type Config struct {
	// ID is the identity of the local raft. ID cannot be 0.（ID是本地raft的标识。ID不能为0。）
	ID uint64

	// peers contains the IDs of all nodes (including self) in the raft cluster. It
	// should only be set when starting a new raft cluster.(只有在启动时才设置该值) Restarting raft from
	// previous configuration will panic(死机) if peers is set. peer is private and only
	// used for testing right now.
	peers []uint64

	// ElectionTick is the number of Node.Tick invocations that must pass between
	// elections. That is, if a follower does not receive any message from the
	// leader of current term before ElectionTick has elapsed, it will become
	// candidate and start an election. ElectionTick must be greater than
	// HeartbeatTick. We suggest ElectionTick = 10 * HeartbeatTick to avoid
	// unnecessary leader switching.
	//ElectionTick是在两次选择之间必须传递的Node.Tick调用数。也就是说，如果一名追随者在
	//选举结束之前没有收到本任期领导人的任何信息，他将成为候选人并开始选举。ElectionTick必须
	//大于HeartbeatTick。我们建议ElectionTick=10*HeartbeatTick，以避免不必要的领导人转换。
	ElectionTick int

	// HeartbeatTick is the number of Node.Tick invocations that must pass between
	// heartbeats. That is, a leader sends heartbeat messages to maintain its
	// leadership every HeartbeatTick ticks.
	HeartbeatTick int

	// Storage is the storage for raft. raft generates entries and states to be
	// stored in storage. raft reads the persisted entries and states out of
	// Storage when it needs. raft reads out the previous state and configuration
	// out of storage when restarting.
	Storage Storage
	// Applied is the last applied index. It should only be set when restarting
	// raft. raft will not return entries to the application smaller or equal to
	// Applied. If Applied is unset when restarting, raft might return previous
	// applied entries. This is a very application dependent configuration.
	Applied uint64
}

func (c *Config) validate() error {
	if c.ID == None {
		return errors.New("cannot use none as id")
	}

	if c.HeartbeatTick <= 0 {
		return errors.New("heartbeat tick must be greater than 0")
	}

	if c.ElectionTick <= c.HeartbeatTick {
		return errors.New("election tick must be greater than heartbeat tick")
	}

	if c.Storage == nil {
		return errors.New("storage cannot be nil")
	}

	return nil
}

// Progress represents a follower’s progress in the view of the leader. Leader maintains
// progresses of all followers, and sends entries to the follower based on its progress.
//进度代表领导者眼中跟随者的进度。Leader维护所有追随者的进度，并根据其进度向追随者发送条目。
type Progress struct {
	Match, Next uint64
}

type Raft struct {
	id uint64

	Term uint64
	Vote uint64

	// the log
	RaftLog *RaftLog

	// log replication progress of each peers（记录每个peer的复制进度)
	Prs map[uint64]*Progress

	mid uint64

	electionRandomTimeout int

	// this peer's role (角色)
	State StateType

	// votes records（投票记录）
	votes map[uint64]bool

	// msgs need to send
	msgs []pb.Message

	// the leader id
	Lead uint64

	// heartbeat interval, should send（心跳间隔）
	heartbeatTimeout int
	// baseline of election interval（选举间隔标准）
	electionTimeout int
	// number of ticks since it reached last heartbeatTimeout.
	// only leader keeps heartbeatElapsed.
	heartbeatElapsed int
	// Ticks since it reached last electionTimeout when it is leader or candidate.
	// Number of ticks since it reached last electionTimeout or received a
	// valid message from current leader when it is a follower.
	electionElapsed int

	// leadTransferee is id of the leader transfer target when its value is not zero.
	// Follow the procedure defined in section 3.10 of Raft phd thesis.
	// (https://web.stanford.edu/~ouster/cgi-bin/papers/OngaroPhD.pdf)
	// (Used in 3A leader transfer)
	leadTransferee uint64

	// Only one conf change may be pending (in the log, but not yet
	// applied) at a time. This is enforced via PendingConfIndex, which
	// is set to a value >= the log index of the latest pending
	// configuration change (if any). Config changes are only allowed to
	// be proposed if the leader's applied index is greater than this
	// value.
	// (Used in 3A conf change)
	PendingConfIndex uint64
}

// newRaft return a raft peer with the given config
func newRaft(c *Config) *Raft {
	if err := c.validate(); err != nil {
		panic(err.Error())
	}
	// Your Code Here (2A).

	pes := make(map[uint64]*Progress)
	vet := make(map[uint64]bool)
	//count := uint64(1)
	for _,i := range c.peers{

		pes[i] = new(Progress)
		vet[i] = false	
			//count++

	}
	mid := uint64(0)
	if len(c.peers)&1 == 1{
		mid = uint64((len(c.peers)+1)/2)
	}else{
		mid = uint64(len(c.peers)/2)
	}

	return &Raft{
		id : c.ID,
		State: StateFollower,
		heartbeatTimeout: c.HeartbeatTick,
		electionTimeout: c.ElectionTick,
		electionRandomTimeout: c.ElectionTick,
		RaftLog: &RaftLog{
			storage: c.Storage,
		},
		heartbeatElapsed: 0,
		electionElapsed: 0,
		Prs: pes,
		votes: vet,
		Vote: None,
		mid: mid,
	}
}


// sendAppend sends an append RPC with new entries (if any) and the
// current commit index to the given peer. Returns true if a message was sent.
func (r *Raft) sendAppend(to uint64) bool {
	// Your Code Here (2A).
	return false
}

// sendHeartbeat sends a heartbeat RPC to the given peer.
func (r *Raft) sendHeartbeat(to uint64) {
	// Your Code Here (2A).
	r.msgs = append(r.msgs, pb.Message{
			MsgType: pb.MessageType_MsgHeartbeat, 
			To: to,
			From: r.id,
			Term: r.Term,
		})
}

// tick advances the internal logical clock by a single tick.
func (r *Raft) tick() {
	// Your Code Here (2A).
	switch r.State{
	case StateFollower:
		r.electionElapsed++
		if r.electionElapsed >= r.electionRandomTimeout{
			r.becomeCandidate()
			r.electionElapsed = 0
			rand.Seed(time.Now().UnixNano())
			r.electionRandomTimeout = rand.Intn(r.electionTimeout)+r.electionTimeout
			r.Step(pb.Message{MsgType: pb.MessageType_MsgHup,})
		}
	case StateCandidate:
		r.electionElapsed++
		if r.electionElapsed >= r.electionRandomTimeout{
			r.Term++
			r.electionElapsed = 0
			rand.Seed(time.Now().UnixNano())
			r.electionRandomTimeout = rand.Intn(r.electionTimeout)+r.electionTimeout
			r.Step(pb.Message{MsgType: pb.MessageType_MsgHup,})
		}
	case StateLeader:
		r.heartbeatElapsed++
		if r.heartbeatElapsed >= r.heartbeatTimeout{
			r.Step(pb.Message{MsgType: pb.MessageType_MsgBeat,})
			r.heartbeatElapsed = 0
		}
	}

}

// becomeFollower transform this peer's state to Follower(转换为follower)
func (r *Raft) becomeFollower(term uint64, lead uint64) {
	// Your Code Here (2A).
	r.Term = term
	r.Lead = lead
	r.State = StateFollower
	//r.electionElapsed = 0
	//fmt.Print("AAA")
}

// becomeCandidate transform this peer's state to candidate（转换为candidate）
func (r *Raft) becomeCandidate() {
	// Your Code Here (2A).
	r.Term++
	r.State = StateCandidate
	r.electionElapsed = 0
	//fmt.Print("BBB")
}

// becomeLeader transform this peer's state to leader(转换为leader)
func (r *Raft) becomeLeader() {
	// Your Code Here (2A).
	// NOTE: Leader should propose a noop entry on its term
	//fmt.Print("ID " ,r.id,  r.mid, len(r.Prs),"\n")
	r.State = StateLeader
	r.electionElapsed = 0
	r.heartbeatElapsed = 0
}

// Step the entrance of handle message, see `MessageType`
// on `eraftpb.proto` for what msgs should be handled
func (r *Raft) Step(m pb.Message) error {
	// Your Code Here (2A).

	switch r.State {
	case StateFollower:
		switch m.MsgType{
		case pb.MessageType_MsgHup:  //选择超时
			r.becomeCandidate()
			r.Step(pb.Message{MsgType: pb.MessageType_MsgHup,})

		case pb.MessageType_MsgAppend:  //复制日志项
			if m.Term < r.Term{
				return nil
			}
			r.Term = m.Term

			r.handleAppendEntries(m)
		case pb.MessageType_MsgRequestVote:  //投票
			if r.Term < m.Term{
				r.Term = m.Term
				r.Vote = None
			}
			p := false
			if r.Vote == None || r.Vote == m.From{
				r.Vote = m.From
			}else{
				p = true
			}
			r.msgs = append(r.msgs, pb.Message{
				From: r.id,
				To: m.From,
				Term: m.Term,
				MsgType: pb.MessageType_MsgRequestVoteResponse,
				Reject: p,
			})
		case pb.MessageType_MsgHeartbeat:   //接收到心跳包
			if m.Term >= r.Term{
				r.becomeFollower(m.Term, m.From)
			}
			r.Vote = m.From
			r.msgs = append(r.msgs, pb.Message{
				From: r.id,
				To: m.From,
				Term: r.Term,
				MsgType: pb.MessageType_MsgHeartbeatResponse,
			})
		}


	case StateCandidate:
		switch m.MsgType{
		case pb.MessageType_MsgAppend:  //复制日志项
			if m.Term >= r.Term{
				r.becomeFollower(m.Term, m.From)
			}
			if m.Term < r.Term{
				return nil
			}
			r.handleAppendEntries(m)
		case pb.MessageType_MsgHeartbeat:  //接收到心跳包，
			if m.Term >= r.Term{
				r.becomeFollower(m.Term, m.From)
				r.Vote = m.From
			}else{
				return nil
			}
			r.msgs = append(r.msgs, pb.Message{
				From: r.id,
				To: m.From,
				Term: r.Term,
				MsgType: pb.MessageType_MsgHeartbeatResponse,
			})
		case pb.MessageType_MsgRequestVote:    //收到选举请求
			T := true
			if m.Term > r.Term{
				T = false
				r.becomeFollower(m.Term, m.From)
				r.Vote = m.From
				r.msgs = append(r.msgs, pb.Message{
					From: r.id,
					To: m.From,
					Term: r.Term,
					MsgType: pb.MessageType_MsgRequestVoteResponse,
					Reject: false,
				})
			}else if r.id != m.From{
				r.msgs = append(r.msgs, pb.Message{
					From: r.id,
					To: r.id,
					Term: r.Term,
					MsgType: pb.MessageType_MsgRequestVoteResponse,
					Reject: T,
				})
			}
		case pb.MessageType_MsgHup:   //候选者发起选举
			if r.Vote == r.id{
				r.Term++
			}
			r.votes[r.id] = true
			r.Vote = r.id
			r.Step(pb.Message{
				From: r.id,
				To: r.id,
				Term: m.Term,
				MsgType: pb.MessageType_MsgRequestVoteResponse,
				Reject: false,
			})
			for p := range r.Prs{
				if p != r.id{
					r.msgs = append(r.msgs, pb.Message{
						MsgType: pb.MessageType_MsgRequestVote,
						From: r.id,
						To: p,
						Term: r.Term,
					})
				}
			}
			
		case pb.MessageType_MsgRequestVoteResponse: //处理收到的选票

			if !m.Reject && m.Term == r.Term{
				r.votes[m.From] = true
			}	
			
			sum := uint64(0)
			nsum := uint64(0)
			for p := range r.votes{
				if r.votes[p]{
					sum++
				}else{
					nsum++
				}
			}
			
			if len(r.Prs)&1 ==0{
				if sum > r.mid{
					r.becomeLeader()
					r.Vote = r.id
					r.Step(pb.Message{MsgType: pb.MessageType_MsgBeat,})
				}
			}else{
				if sum >= r.mid{
					//fmt.Print("vote" , sum, r.mid, len(r.Prs),"\n")
					r.becomeLeader()
					r.Vote = r.id
					r.Step(pb.Message{MsgType: pb.MessageType_MsgBeat,})
				}
			}
		}


	case StateLeader:
		switch m.MsgType{
		case pb.MessageType_MsgAppend:  //复制日志项
			if m.Term > r.Term{
				r.becomeFollower(m.Term, m.From)
			}
			if m.Term < r.Term{
				return nil
			}

		case pb.MessageType_MsgPropose:   //领导者接受一个新信息，本地信息
			for p := range r.Prs{
				if p != r.id{
					r.sendAppend(p)
				}
			}
		case pb.MessageType_MsgBeat:   //该向追随者发心跳包了
			for p := range r.Prs{
				if p != r.id{
					r.sendHeartbeat(p)
				}
			}
		case pb.MessageType_MsgRequestVote:
			p := false
			if r.Term < m.Term{
				r.becomeFollower(m.Term, m.From)
				r.Vote = m.From
			}else{
				p = true
			}
			r.msgs = append(r.msgs, pb.Message{
				From: r.id,
				To: m.From,
				Term: m.Term,
				MsgType: pb.MessageType_MsgRequestVoteResponse,
				Reject: p,
			})
		
			}
	
	}
	return nil
}

// handleAppendEntries handle(处理) AppendEntries RPC request
func (r *Raft) handleAppendEntries(m pb.Message) {
	// Your Code Here (2A).
}

// handleHeartbeat handle（处理） Heartbeat RPC request
func (r *Raft) handleHeartbeat(m pb.Message) {
	// Your Code Here (2A).
}

// handleSnapshot handle Snapshot RPC request
func (r *Raft) handleSnapshot(m pb.Message) {
	// Your Code Here (2C).
}

// addNode add a new node to raft group
func (r *Raft) addNode(id uint64) {
	// Your Code Here (3A).
}

// removeNode remove a node from raft group
func (r *Raft) removeNode(id uint64) {
	// Your Code Here (3A).
}
