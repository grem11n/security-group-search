# Simple script to search in AWS security groups
## Description
Sometimes you may want to change security groups rules and it's great to know what exactly would be affected with such change.

This script is designed to simply output a list of security groups, which contain provided pattern.
It could be an IP or another security group ID. Script only do string search in IpPermissions.

## Installation
Script is written in Go 1.6. No special dependencies needed. You can simply get this script and compile it locally

## Usage

* ```-config```  Allow changing path to the file with AWS credentials
* ```-region```  Defines region (default "us-east-1")
* ```-section``` Which section of AWS credentials to use (default "default")
* ```-pattern``` Set what are you looking for. Could be regexp
* ```-egress```  Search in egress rules