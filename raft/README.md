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
If a follower's election timeout goes off before receiving a request from the leader, promotes itself to candidate, starts an election, votes for itself, and requests votes from all other nodes. Each election is in a new term - an integer that is incremented each time. 

The other nodes vote for the first request they receive, and reject subsequent requests. If the candidate receives votes from a majority of the nodes, it promotes itself to leader and begins replicating its log to the other nodes. 
This is also the protocol to elect the leader when the nodes first start up.

### RPCs
#### Request for votes
When a candidate requests votes from the other nodes, it sends its id, the index of the last entry in its log and the term it was appended in. It also sends the current term.

Nodes respond with the term they have seen last, and whether they are granting the vote or not.

#### Append Entry
When a leader gets a client request to set a key/value pair it appends it to its log and sends a request to all the followers to append the entry to their logs. The request also includes (apart from the entry itself) the current term, the leader's id, the index of the last log entry before this new one, and the term of that entry. Finally, the request also includes the index of the entry that was last committed to the leader's database.

When a follower receives an append entry request, it responds with its term and success or failure. If the receiver's term is greater than the leader's, it responds false. It also responds false if the entry it stored at the leader's previous log index has a term that is different from the previous log term sent in the request.
