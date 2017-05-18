package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type secGroup struct {
	ID          string
	Name        string
	Description string
	Permissions []*ec2.IpPermission
}

type connParams struct {
	Region  string
	Config  string
	Section string
	Egress  bool
}

var (
	configPtr  = flag.String("config", "", "Allow changing path to the file with AWS credentials")
	sectionPtr = flag.String("section", "default", "Which part of AWS credentials to use")
	regionPtr  = flag.String("region", "us-east-1", "Defines region")
	patternPtr = flag.String("pattern", "", "Specify pattern to find")
	egressPtr  = flag.Bool("egress", false, "Search Egress rules. Search in Ingress by default")
)

func getSecurityGroups(connParams connParams) []secGroup {
	var sgList []secGroup
	creds := credentials.NewSharedCredentials(connParams.Config, connParams.Section)
	_, err := creds.Get()
	if err != nil {
		log.Fatal(err)
	}

	svc := ec2.New(session.New(), &aws.Config{
		Region:      aws.String(*regionPtr),
		Credentials: creds,
	})

	sgIn := &ec2.DescribeSecurityGroupsInput{}
	securityGroupsResult, err := svc.DescribeSecurityGroups(sgIn)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	for _, sgs := range securityGroupsResult.SecurityGroups {
		var perm []*ec2.IpPermission
		id := *sgs.GroupId
		name := *sgs.GroupName
		desc := *sgs.Description
		if connParams.Egress {
			perm = sgs.IpPermissionsEgress
		} else {
			perm = sgs.IpPermissions
		}
		sg := secGroup{
			ID:          id,
			Name:        name,
			Description: desc,
			Permissions: perm,
		}
		sgList = append(sgList, sg)
	}
	return sgList
}

func main() {
	flag.Parse()
	var match = regexp.MustCompile(*patternPtr)
	if *patternPtr == "" {
		log.Fatal("No search pattern specified")
	}
	connParams := connParams{
		Region:  *regionPtr,
		Config:  *configPtr,
		Section: *sectionPtr,
		Egress:  *egressPtr,
	}
	sgList := getSecurityGroups(connParams)
	write := new(tabwriter.Writer)
	write.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintln(write, "   GroupId\t|\t     Name\t|\t       Description\t")

	for _, sgs := range sgList {
		for _, perms := range sgs.Permissions {
			if match.MatchString(perms.GoString()) {
				fmt.Fprintln(write, sgs.ID, "\t|\t", sgs.Name, "\t|\t", sgs.Description, "\t")
				break
			}
		}
	}
	write.Flush()
}
