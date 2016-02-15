package main

import (
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	term "github.com/buger/goterm"
)

var profile = flag.String("p", "", "AWS config profile")

func main() {
	flag.Parse()

	var awsSession *session.Session
	var err error
	var clusters []*string

	// Create session from profile or other...
	if *profile != "" {
		awsSession, err = GetSessionWithProfile(*profile)
	} else {
		awsSession, err = GetSession()
	}

	// Error check: session
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	svc := ecs.New(awsSession)
	//svc := ec2.New(session.New(), &aws.Config{Region: aws.String("us-west-2")})

	// List clusters
	lcParams := &ecs.ListClustersInput{
		MaxResults: aws.Int64(100),
		NextToken:  aws.String(""),
	}

	for {
		resp, err := svc.ListClusters(lcParams)

		// Error check
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		//for i := range resp.ClusterArns {
		clusters = append(clusters, resp.ClusterArns...)
		//}

		// Cycle token
		if resp.NextToken != nil {
			lcParams.NextToken = resp.NextToken
		} else {
			// Kill loop ... out of clusters
			break
		}

	}

	// Get cluster details for all the things...
	dcParams := &ecs.DescribeClustersInput{
		Clusters: clusters,
	}
	resp, err := svc.DescribeClusters(dcParams)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	// Pretty-print the response data.
	term.Println()
	topTable := term.NewTable(0, 10, 5, ' ', 0)
	fmt.Fprintf(topTable, "Cluster\tStatus\tPendingTasks\tRunningTasks\tContainerInstances\n")
	for i := range resp.Clusters {
		fmt.Fprintf(topTable, "%s\t%s\t%d\t%d\t%d\n", *resp.Clusters[i].ClusterName, *resp.Clusters[i].Status, *resp.Clusters[i].PendingTasksCount, *resp.Clusters[i].RunningTasksCount, *resp.Clusters[i].RegisteredContainerInstancesCount)
	}
	term.Println(topTable)
	term.Flush()

	// Get task details
}
