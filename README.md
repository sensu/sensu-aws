# aws-plugins
![Alt text](https://travis-ci.org/sreejita-biswas/aws-handler.svg?branch=develop "TravisCI")
## Functionality

**check-ec2-cpu_balance**

**check-ec2-filter**

**check-ec2-network**

## Files

* /plugins/ec2/check-ec2-cpu_balance/check-ec2-cpu_balance.go
* /plugins/ec2/check-ec2-filter/check-ec2-filter.go
* /plugins/ec2/check-ec2-network/check-ec2-network.go

## Binaries

* /plugins/ec2/check-ec2-cpu_balance/bin/check-ec2-cpu_balance
* /plugins/ec2/check-ec2-filter/bin/check-ec2-filter
* /plugins/ec2/check-ec2-filter/bin/check-ec2-network

## Usage

**check-ec2-cpu_balance**

```
$ ./check-ec2-cpu_balance -h
Usage of ./check-ec2-cpu_balance:
  -critical float
    	Trigger a critical when value is below the criticalThreshold. (default 1.2)
  -tag string
    	Add instance TAG value to warn/critical message. (default "NAME")
  -warning float
    	Trigger a warning when value is below warningThreshold (default 2.3)
```
**check-ec2-cpu_balance**

```
$ ./check-ec2-filter -h
Usage of ./check-ec2-filter:
  -compare string
    	Comparision operator for threshold: equal, not, greater, less (default "equal")
  -critical int
    	Critical threshold for filter (default 1)
  -detailed_message
    	Detailed description is required or not
  -exclude_tags string
    	JSON String Representation of tag values (default "{}")
  -filters string
    	JSON String representation of Filters (default "{\"filters\" : [{\"name\" : \"instance-state-name\", \"values\": [\"running\"]}]}")
  -min_running_secs float
    	Minimum running seconds (default 60)
  -warning int
    	Warning threshold for filter (default 2)

```
**check-ec2-network**

```
$ ./check-ec2-network -h

Usage of ./check-ec2-network:
  -critical float
    	Trigger a critical if network traffice is over specified Bytes (default 1e+06)
  -direction string
    	Select NetworkIn or NetworkOut (default "NetworkIn")
  -instance_id string
    	EC2 Instance ID to check.
  -period int
    	CloudWatch metric statistics period in seconds (default 60)
  -start_time string
    	CloudWatch metric statistics end time, e.g. 2014-11-12T11:45:26.371Z (default "2018-12-05T13:34:18+05:30")
  -warning float
    	Trigger a warning if network traffice is over specified Bytes (default 1.5e+06)
      
```
   
## AWS Configuration

```
Sample Credential Configuration:
[default]
aws_access_key_id=AGFGFHGFHHJGHJG
aws_secret_access_key=cdfedbfdjdjsbsjdgbjdsgbjdskg

Sample Config:
[default]
region=us-east-2
output=json

$ mkdir ~/.aws
$ cd ~/.aws
$ vi credentials - copy and paste the above sample credential and change the aws_access_key_id and aws_secret_access_key to some valid values. Save the file.
$ vi config - copy and paste the above sample config and change the region to some valid value. Save the file.

```

## Example

```
Command : ./check-ec2-filter -filters="{\"filters\" : [{\"name\" : \"instance-state-name\", \"values\": [\"running\"]}]}"
Output : Critical threshold for filter ,  Current Count : 1  

```
## Binary Generation

```
Environment : MAC OS/Linux

1. Install Go - https://nats.io/documentation/tutorials/go-install/
2. Clone the code using command - "go get github.com/sreejita-biswas/aws-plugins"
3. $ cd ~/go/src/github.com/sreejita-biswas/aws-plugins/plugins/check-ec2-filter 
4. go build check-ec2-filter.go
5. You will find the binary in the current directory. If you want you can move the same to ~/go/src/github.com/sreejita-biswas/aws-plugins/plugins/ec2/check-ec2-filter/bin directory.

```
