# go-ecs

go-ecs is a simple command line utility to view [AWS ECS](https://aws.amazon.com/ecs/) cluster status quickly.

## Installation

On OS X or Linux:

```
curl https://raw.githubusercontent.com/colinmutter/go-ecs/master/install.sh | sh
```

On Windows download [binary](https://github.com/colinmutter/go-ecs/releases).

## Usage

```
go-ecs [options]
  -p profile      aws profile name
```

## Configuration
An IAM Policy will need to be created, or the following permissions granted to allow this utility to function.

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ecs:DescribeClusters",
                "ecs:DescribeContainerInstances",
                "ecs:DescribeServices",
                "ecs:DescribeTaskDefinition",
                "ecs:DescribeTasks",
                "ecs:DiscoverPollEndpoint",
                "ecs:ListClusters",
                "ecs:ListContainerInstances",
                "ecs:ListServices",
                "ecs:ListTaskDefinitions",
                "ecs:ListTasks",
                "ecs:Poll",
                "ec2:DescribeInstances"
            ],
            "Resource": [
                "*"
            ]
        }
    ]
}
```
