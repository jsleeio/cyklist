# cyklist

Gentle replacement of Kubernetes nodes hosted in AWS.

Much more info in [the tech proposal](PROPOSAL.md).

## AWS API permissions required

Phase       | AWS API actions required
----------- | ----------------------------------
All phases  | `ec2:DescribeInstances`
`detach`    | `autoscaling:DetachInstances`
`drain`     | no phase-specific permissions required
`terminate` | `ec2:TerminateInstances`
