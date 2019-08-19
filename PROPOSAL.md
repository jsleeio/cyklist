# node lifecycle management v2 proposal

## overview

Nodes move through a multi-phase lifecycle. At any time, a node's current phase
is indicated via an AWS resource tag `cyklist.io/phase`.  The
lifecycle phases of an instance are, in order from launch to termination:

phase     | tag value   | description
--------- | ----------- | -------------------------------------------------
normal    | _none_      | The starting state. Normal operations, hosting workloads
detach    | `detach`    | The node is detaching from its parent autoscaling group
drain     | `drain`     | The [Kubernetes] node is being cordoned and drained
terminate | `terminate` | The node may be terminated at any time

## phases

### normal

This is where nodes start their life. In this state, applications are running
on the node as normal.

After some time, if newer EC2 AMIs become available, they will have their first
phase tag affixed by the `amiupdate` application, which scans AWS autoscaling
groups and updates their launch configurations to use a newer EC2 AMIs. The
phase tag will be set to `detach`.

### detach

In the `detach` phase, workloads are still running on the node, but it is ready
to be detached from its parent autoscaling group prior to being drained and
terminated.

The benefit here is that the autoscaling group has an opportunity to launch a
replacement instance _before_ (in the next phase) the detached node is drained
and terminated; this increases the likelihood that workloads evicted by the
draining process can be immediately recreated elsewhere, rather than having to
wait for the cluster autoscaler to catch up and provision a new node.

### drain

In the `drain` phase, non-`Daemonset` workloads are drained from the node using
`kubectl drain`, and as part of that process, the node is also cordoned such
that no new workloads can be scheduled on it. The explicit drain operation
allows pods to be evicted with graceful termination.

As `Daemonset` workloads are ignored by the drain operation, a useful side
effect of this phase is that container log collection agents on the node will
have time to catch up on any backlog.

### terminate

In the `terminate` phase, the node is considered ready for termination. At this
point, all possible attempts to "nicely" shut down workloads on the node have
been attempted, so a simple EC2 `TerminateInstances` API call is sufficient.
