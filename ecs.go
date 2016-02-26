package main

import (
	"flag"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/buger/goterm"
)

var profile = flag.String("p", "", "AWS config profile")

// type ClusterServices struct {
// 	cluster       *string
// 	clusterDetail []*ecs.Cluster
// 	services      []*string
// 	serviceDetail []*ecs.Service
// }

func main() {
	flag.Parse()

	var awsSession *session.Session
	var err error
	var clusters []*string

	var clusterStatus *ecs.Cluster
	var clusterServices []*ecs.Service
	var clusterTasks []*ecs.Task
	var clusterContainerInstances []*ecs.ContainerInstance
	var ec2Instances []*ec2.Reservation

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
	ec2svc := ec2.New(awsSession)

	//svc := ec2.New(session.New(), &aws.Config{Region: aws.String("us-west-2")})

	// Get clusters
	clusters, err = ListClusters(svc)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	goterm.Println()

	// Get lots of extra data by cluster
	for _, thisCluster := range clusters {

		// Get cluster statuses
		clusterStatus, err = DescribeCluster(svc, thisCluster)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		// Get cluster services
		clusterServices, err = GetServices(svc, thisCluster)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		// Get cluster tasks
		clusterTasks, err = GetTasks(svc, thisCluster)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		// Get cluster container instances (tuple)
		clusterContainerInstances, ec2Instances, err = GetContainerInstances(svc, ec2svc, thisCluster)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		clusterServices = clusterServices
		clusterStatus = clusterStatus
		clusterTasks = clusterTasks
		clusterContainerInstances = clusterContainerInstances
		ec2Instances = ec2Instances
		//fmt.Println(clusterServices)
		//fmt.Println(clusterStatus)
		//fmt.Println(ec2Instances)

		// Pretty-print the response data.
		clusterTable := goterm.NewTable(0, 10, 5, ' ', 0)

		fmt.Fprintf(clusterTable, "Cluster\tStatus\tPending Tasks\tRunning Tasks\tContainer Instances\n")
		fmt.Fprintf(clusterTable, "%s\t%s\t%d\t%d\t%d\n", *clusterStatus.ClusterName,
			*clusterStatus.Status, *clusterStatus.PendingTasksCount, *clusterStatus.RunningTasksCount,
			*clusterStatus.RegisteredContainerInstancesCount)

		serviceTable := goterm.NewTable(0, 10, 5, ' ', 0)
		fmt.Fprintf(serviceTable, "Service\tStatus\tTask Definition\tPending Tasks\tRunning Tasks\n")
		taskDefRex := regexp.MustCompile("task-definition/(.+)$")
		for _, service := range clusterServices {
			taskDef := taskDefRex.FindString(*service.TaskDefinition)
			fmt.Fprintf(serviceTable, "%s\t%s\t%s\t%d\t%d\n", *service.ServiceName, *service.Status,
				taskDef, *service.PendingCount, *service.RunningCount)
		}

		taskTable := goterm.NewTable(0, 10, 5, ' ', 0)
		fmt.Fprintf(taskTable, "Task Service\tTask Arn\tStatus\tDesired Status\n")
		//taskDefRex := regexp.MustCompile("task-definition/(.+)$")
		for _, task := range clusterTasks {
			taskDefName := ""
			// Look up the service this task belongs to
			for _, service := range clusterServices {
				if *service.TaskDefinition == *task.TaskDefinitionArn {
					taskDefName = *service.ServiceName
					break
				}
			}
			//taskDef := taskDefRex.FindString(*service.TaskDefinition)
			fmt.Fprintf(taskTable, "%s\t%s\t%s\t%s\n", taskDefName, *task.TaskArn, *task.LastStatus, *task.DesiredStatus)
		}

		goterm.Println(clusterTable)
		goterm.Println(serviceTable)
		goterm.Println(taskTable)
		goterm.Println()
	}

	goterm.Flush()

	// Describe cluster services

	// Get task details
}

// ListClusters will return a slice of ECS Clusters
func ListClusters(svc *ecs.ECS) ([]*string, error) {
	var clusters []*string

	// List clusters
	reqParams := &ecs.ListClustersInput{
		MaxResults: aws.Int64(100),
		NextToken:  aws.String(""),
	}

	for {
		resp, err := svc.ListClusters(reqParams)

		// Error check
		if err != nil {
			return nil, fmt.Errorf("ecs.ListClusters: %s", err.Error())
		}

		// Expand slice of clusters and append to our comprehensive list
		clusters = append(clusters, resp.ClusterArns...)

		// Cycle token
		if resp.NextToken != nil {
			reqParams.NextToken = resp.NextToken
		} else {
			// Kill loop ... out of clusters
			break
		}

	}

	return clusters, nil
}

// DescribeCluster will return a Cluster (detail struct)
func DescribeCluster(svc *ecs.ECS, cluster *string) (*ecs.Cluster, error) {
	// Get cluster details for all the things...
	reqParams := &ecs.DescribeClustersInput{
		Clusters: []*string{cluster},
	}

	resp, err := svc.DescribeClusters(reqParams)

	if err != nil {
		return nil, fmt.Errorf("ecs.DescribeClusters: %s", err.Error())
	}

	return resp.Clusters[0], err
}

// GetServices will return a slice of ECS Services for a given cluster
func GetServices(svc *ecs.ECS, cluster *string) ([]*ecs.Service, error) {

	var serviceArns []*string

	// List clusters
	reqParams := &ecs.ListServicesInput{
		Cluster:    cluster,
		MaxResults: aws.Int64(10),
		NextToken:  aws.String(""),
	}

	// Loop through tokens until no more results remain
	for {
		resp, err := svc.ListServices(reqParams)

		// Error check
		if err != nil {
			return nil, fmt.Errorf("ecs.ListServices: %s", err.Error())
		}

		// Expand slice of services and append to our comprehensive list
		serviceArns = append(serviceArns, resp.ServiceArns...)

		// Cycle token
		if resp.NextToken != nil {
			reqParams.NextToken = resp.NextToken
		} else {
			// Kill loop ... out of response pages
			break
		}

	}

	// Describe the services that we just got back
	resp, err := svc.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  cluster,
		Services: serviceArns,
	})
	if err != nil {
		return nil, fmt.Errorf("ecs.DescribeServices: %s", err.Error())
	}

	return resp.Services, nil
}

// GetTasks will return a slice of ECS Tasks within a cluster
func GetTasks(svc *ecs.ECS, cluster *string) ([]*ecs.Task, error) {

	var taskArns []*string

	// List clusters
	reqParams := &ecs.ListTasksInput{
		Cluster:    cluster,
		MaxResults: aws.Int64(100),
		NextToken:  aws.String(""),
	}

	// Loop through tokens until no more results remain
	for {
		resp, err := svc.ListTasks(reqParams)

		// Error check
		if err != nil {
			return nil, fmt.Errorf("ecs.ListTasks: %s", err.Error())
		}

		// Expand slice of tasks and append to our comprehensive list
		taskArns = append(taskArns, resp.TaskArns...)

		// Cycle token
		if resp.NextToken != nil {
			reqParams.NextToken = resp.NextToken
		} else {
			// Kill loop ... out of response pages
			break
		}

	}

	// Describe the tasks that we just got back
	resp, err := svc.DescribeTasks(&ecs.DescribeTasksInput{
		Cluster: cluster,
		Tasks:   taskArns,
	})

	if err != nil {
		return nil, fmt.Errorf("ecs.DescribeTasks: %s", err.Error())
	}

	return resp.Tasks, nil
}

// GetContainerInstances will return a slice of ECS Container Instances within a cluster
func GetContainerInstances(svc *ecs.ECS, ec2svc *ec2.EC2, cluster *string) (containerInstances []*ecs.ContainerInstance, instances []*ec2.Reservation, e error) {

	var ciArns []*string

	// List clusters
	reqParams := &ecs.ListContainerInstancesInput{
		Cluster:    cluster,
		MaxResults: aws.Int64(100),
		NextToken:  aws.String(""),
	}

	// Loop through tokens until no more results remain
	for {
		resp, err := svc.ListContainerInstances(reqParams)

		// Error check
		if err != nil {
			return nil, nil, fmt.Errorf("ecs.ListContainerInstances: %s", err.Error())
		}

		// Expand slice of container instances and append to our comprehensive list
		ciArns = append(ciArns, resp.ContainerInstanceArns...)

		// Cycle token
		if resp.NextToken != nil {
			reqParams.NextToken = resp.NextToken
		} else {
			// Kill loop ... out of response pages
			break
		}

	}

	// Describe the tasks that we just got back
	ciResponse, err := svc.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
		Cluster:            cluster,
		ContainerInstances: ciArns,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("ecs.DescribeContainerInstances: %s", err.Error())
	}

	var instanceIds []*string
	for _, k := range ciResponse.ContainerInstances {
		instanceIds = append(instanceIds, k.Ec2InstanceId)
	}

	// Create a map of container instances by ci arn...
	// Note: Will work for <= 1000 instances w/o having to use NextToken
	ec2Resp, err := ec2svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: instanceIds,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("ec2.DescribeInstances: %s", err.Error())
	}

	return ciResponse.ContainerInstances, ec2Resp.Reservations, nil
}
