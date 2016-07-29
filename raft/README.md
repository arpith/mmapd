# Raft
Raft is a distributed consensus algorithm/protocol. This is my implementation in-progress.

## Issues
### Stable initial election
For a cluster of three nodes, stable election happens only when you kill a node a couple of times. My guess is that this is because a leader steps down if it gets a heartbeat with the same term.

### Bullying
When a leader dies, and comes up after sometime, the desired behaviour is that it respects the state of the cluster and accepts missing entries from the current leader. This is not what happens here! Old leaders regain leadership when they come up! This could be related to the previous issue.

## How it works
### Timeouts
There are two randomized timeouts, typically between 150 and 300ms. The heartbeat timeout is used by the leader to inform the followers that it is still up. When a follower gets a request from the leader it resets its election timeout. 

### Leadership
If a follower's election timeout goes off before receiving a request from the leader, it starts an election, votes for itself, and requests votes from all other nodes. The other nodes vote for the first request they receive, and reject subsequent requests. This is also the protocol to elect the leader when the nodes first start up.

### RPCs
#### Request for votes
#### Append Entry

