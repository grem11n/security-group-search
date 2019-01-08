package main

import (
	"encoding/json"
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

type sgInfo struct {
	ID          string `json:"groupId"`
	Name        string `json:"groupName"`
	Rule        string `json:"rule"`
	Description string `json:"Description"`
}

type output []sgInfo

var (
	configPtr  = flag.String("config", "", "Allow changing path to the file with AWS credentials")
	sectionPtr = flag.String("section", "default", "Which part of AWS credentials to use")
	regionPtr  = flag.String("region", "us-east-1", "Defines region")
	ingressPtr = flag.String("ingress", "", "Specify ingress to find")
	egressPtr  = flag.Bool("egress", false, "Search Egress rules. Search in Ingress by default")
	outputPtr  = flag.String("output", "table", "Set output format")
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

func compileOutput(sgList []secGroup) {
	var match = regexp.MustCompile(*ingressPtr)
	var re = regexp.MustCompile(`\r?\n|\t+|\s+`)
	if *outputPtr == "table" {
		write := new(tabwriter.Writer)
		write.Init(os.Stdout, 0, 8, 0, '\t', 0)
		fmt.Fprintln(write, "   GroupId\t|\t     Name\t|\t       Rule\t|\t       Description\t")

		for _, sgs := range sgList {
			for _, perms := range sgs.Permissions {
				if match.MatchString(perms.GoString()) {
					rules := re.ReplaceAllString(perms.GoString(), " ")
					fmt.Fprintln(write, sgs.ID, "\t|\t", sgs.Name, "\t|\t", rules, "\t|\t", sgs.Description, "\t")
					break
				}
			}
		}
		write.Flush()
	} else if *outputPtr == "json" {
		var out output
		for _, sgs := range sgList {
			for _, perms := range sgs.Permissions {
				if match.MatchString(perms.GoString()) {
					rules := re.ReplaceAllString(perms.GoString(), " ")
					info := sgInfo{sgs.ID, sgs.Name, rules, sgs.Description}
					out = append(out, info)
					break
				}
			}
		}
		jsonOut, err := json.Marshal(out)
		if err != nil {
			log.Fatal("Unable to create JSON output.")
		}
		fmt.Println(string(jsonOut))
	} else if *outputPtr == "text" {
		for _, sgs := range sgList {
			for _, perms := range sgs.Permissions {
				if match.MatchString(perms.GoString()) {
					rules := re.ReplaceAllString(perms.GoString(), " ")
					fmt.Println("   GroupId\t|\t     Name\t|\t       Rule\t|\t       Description\t")
					fmt.Println(sgs.ID, sgs.Name, rules, sgs.Description)
					break
				}
			}
		}

	} else {
		log.Fatal("Only table, json, and text outputs are supported for now.")
	}
}

func main() {
	flag.Parse()
	if *ingressPtr == "" {
		log.Fatal("No search ingress specified")
	}
	connParams := connParams{
		Region:  *regionPtr,
		Config:  *configPtr,
		Section: *sectionPtr,
		Egress:  *egressPtr,
	}
	sgList := getSecurityGroups(connParams)
	compileOutput(sgList)
}
