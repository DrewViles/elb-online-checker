package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"log"
	"os"
)

var (
	checkForVPC bool
	awsProfile, region, vpcID, internetFacingFlag *string
	loadBalancers []types.LoadBalancer
	conf aws.Config
	err error
)

func main() {
	ParseFlags()
	fmt.Printf("Using AWS Profile: %s\n", *awsProfile)

	conf = InitConfig()
	svc := elasticloadbalancingv2.NewFromConfig(conf)

	loadBalancers = FetchAllELBs(svc)

	ParseLoadBalancers()
}

func ParseLoadBalancers() {
	if loadBalancers != nil {
		if *internetFacingFlag == "internet" {
			fmt.Println("Internet Facing Load Balancers")
			loadBalancers = FetchInternetFacingELBs(loadBalancers)
		} else if *internetFacingFlag == "internal" {
			fmt.Println("Internal Facing Load Balancers")
			loadBalancers = FetchInternalFacingELBs(loadBalancers)
		} else {
			fmt.Println("All Load Balancers")
		}

		if checkForVPC {
			loadBalancers = FilterByVPC(loadBalancers)
		}

		returnString := ""
		for _, elb := range loadBalancers {
			returnString = fmt.Sprintf("%s%s\n", returnString, PrintLBInfo(elb))
		}
		fmt.Printf("%v\n", returnString)
	}
}



func InitConfig() aws.Config {
	c, er := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(*region),
		config.WithSharedConfigProfile(*awsProfile))
	if er != nil {
		log.Fatal(er)
	}
	return c
}

func FetchAllELBs(svc *elasticloadbalancingv2.Client) []types.LoadBalancer {
	els, e := svc.DescribeLoadBalancers(context.TODO(), &elasticloadbalancingv2.DescribeLoadBalancersInput{})
	if e != nil {
		log.Fatal(e.Error())
	}
	return els.LoadBalancers
}

func ParseFlags() {
	awsProfile = flag.String("profile", "", "aws profile name")
	region = flag.String("region", "eu-west-1", "The AWS Region to search in.")
	vpcID = flag.String("vpc", "", "Set a VPC ID to run this against.")
	internetFacingFlag = flag.String("internet", "all", "Option: all, internet, internal. 'all' returns everything. 'internal' fetches internal facing Load Balancers and 'internet' will face internet-facing Load Balancers")

	flag.Parse()

	if *awsProfile == ""{
		prof := os.Getenv("AWS_PROFILE")
		awsProfile = &prof
	}

	if *awsProfile == "" {
		//TODO: Search for kube2iam/internal locations
	}

	if *vpcID != "" {
		checkForVPC = true
	}
}

func FilterByVPC(lbs []types.LoadBalancer) []types.LoadBalancer {
	var elbs []types.LoadBalancer

	fmt.Printf("VPC-ID: %s\n\n", *vpcID)
	for _, elb := range lbs{
		if *elb.VpcId == *vpcID{
			elbs = append(elbs, elb)
		}
	}

	return elbs
}

func FetchInternetFacingELBs(lbs []types.LoadBalancer) []types.LoadBalancer {
	var elbs []types.LoadBalancer

	for _, elb := range lbs{
		if elb.Scheme == types.LoadBalancerSchemeEnumInternetFacing{
			elbs = append(elbs, elb)
		}
	}

	return elbs
}

func FetchInternalFacingELBs(lbs []types.LoadBalancer) []types.LoadBalancer {
	var elbs []types.LoadBalancer

	for _, elb := range lbs{
		if elb.Scheme == types.LoadBalancerSchemeEnumInternal{
			elbs = append(elbs, elb)
		}
	}
	return elbs
}

func PrintLBInfo(elb types.LoadBalancer) string {
	var printInfo string
	printItems := map[string]string{
		"Name": *elb.LoadBalancerName,
		"ARN": *elb.LoadBalancerArn,
		"VPC-ID": *elb.VpcId,
		"Scheme": fmt.Sprintf("%v", elb.Scheme),
	}

	for k, v := range printItems{
		printInfo = fmt.Sprintf("%s%s:\t\t%s\n", printInfo, k, v)
	}

	return printInfo
}