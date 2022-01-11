# 分布式

[toc]

## 1. raft

区别：

- Paxos
  - 角色多；Client ：系统外部角色、Proposer: 接受Client请求，想集群提出提议(propose)、Accpetor(Voter): 投票者、Learner：接受者
  - 分四个阶段；Phase 1a： Prepare阶段、hase 1b: Promise阶段、Phase 2a: Accept阶段、Phase2b: Accepted阶段
  - 相对复杂

- raft
  - 定义新角色；`Leader` 一个集群只有一个leader、`Follower` 一个服从leader决定的角色、 `Cadidate` follower发现集群没有leader，

- ZAB
  - 基本和raft相同，只是在一些名词的叫法上有一些区别，比如ZAB 将某一个leader的周期称为epoch,而raft称为 term。
  - 实现上的话 raft为了保证日志连续性，心跳方向是从leader到follower，ZAB则是相反的。
