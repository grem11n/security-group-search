package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/olekukonko/tablewriter"
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

func searchSg(sgList []secGroup) [][]string {
	var match = regexp.MustCompile(*ingressPtr)
	var re = regexp.MustCompile(`\r?\n|\t+|\s+`)
	var out [][]string
	for _, sgs := range sgList {
		for _, perms := range sgs.Permissions {
			if match.MatchString(perms.GoString()) {
				rules := re.ReplaceAllString(perms.GoString(), " ")
				info := []string{sgs.ID, sgs.Name, rules, sgs.Description}
				out = append(out, info)
				break
			}
		}
	}
	return out
}

func compileOutput(fn func([]secGroup) [][]string, sgList []secGroup) {
	var data = fn(sgList)
	switch *outputPtr {
	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Rule", "Description"})
		table.SetAutoMergeCells(true)
		table.SetRowLine(true)
		table.AppendBulk(data)
		table.Render()
	case "json":
		var jRaw []map[string]string
		for _, val := range data {
			element := make(map[string]string)
			element["ID"] = val[0]
			element["Name"] = val[1]
			element["Rule"] = val[2]
			element["Description"] = val[3]
			jRaw = append(jRaw, element)
		}
		jsonOut, err := json.Marshal(jRaw)
		if err != nil {
			log.Fatal("Unable to create JSON output.")
		}
		fmt.Println(string(jsonOut))
	case "text":
		for _, v := range data {
			fmt.Println("\t ID, \t", "\t Name, \t,", "\t Rule, \t", "\t Description \t")
			fmt.Println(strings.Join(v, ", "))
		}
	default:
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
	compileOutput(searchSg, sgList)
}
