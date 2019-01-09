# Simple script to search in AWS security groups
## Description
Sometimes you may want to change security groups rules and it's great to know what exactly would be affected with such change.

This script is designed to simply output a list of security groups, which contain provided pattern.
It could be an IP or another security group ID. Script only do string search in IpPermissions.

## Installation
I use [Go Modules](https://github.com/golang/go/wiki/Modules) to manage dependencies. Hence, to build this, you need to have Go >= 1.11

## Usage

* ```-config```  Allow changing path to the file with AWS credentials
* ```-region```  Defines region (default "us-east-1")
* ```-section``` Which section of AWS credentials to use (default "default")
* ```-ingress``` Set what are you looking for in ingress rules. Could be regexp
* ```-egress```  Search in egress rules
* ```-output```  Set output format. `table`, `json`, and `text` formats are supported for now (default "`table`")
