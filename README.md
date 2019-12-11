
[![Bonsai Asset Badge](https://img.shields.io/badge/Sensu%20AWS-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/sensu/sensu-aws) TravisCI: [![Build Status](https://travis-ci.org/sensu/sensu-aws.svg?branch=master)](https://travis-ci.org/sensu/sensu-aws)

# Sensu Go AWS Plugin Collection

- [Overview](#overview)
- [Usage examples](#usage-examples)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Check definition](#check-manifest)
  - [On-disk configuration](#on-disk-configuration)
- [Installation from source](#installation-from-source)
- [Contributing](#contributing)

## Overview

This Sensu Go plugin collection provides metric and status checks for monitoring AWS services with Sensu. 

## Usage examples

**check-alb-target-group-health**

```
  ./check-alb-target-group-health --aws_region=us-east-1

  ./check-alb-target-group-health --aws_region=us-east-1 --target_groups=target-group-1
  
  ./check-alb-target-group-health --aws_region=us-east-1 --target_groups=target-group-a,target-group-b
  
```
**check-cloudwatch-alarm**

```
  ./check-cloudwatch-alarm --aws_region=eu-west-1
   
  ./check-cloudwatch-alarm --state=ALEARM

```
**check-cloudwatch-alarms**

```
  ./check-cloudwatch-alarms --exclude_alarms=CPUAlarmLow
  
  ./check-cloudwatch-alarms --aws_region=eu-west-1 --exclude_alarms=CPUAlarmLow
  
  ./check-cloudwatch-alarms --state=ALEARM
      
```
**check-cloudwatch-composite-metric**

```
  ./check-cloudwatch-composite-metric --namespace AWS/ELB --dimensions="LoadBalancerName=test-elb" 
                                      --period=60 --statistics=Maximum --operator=equal --critical=0 
```

**check-ebs-burst-limit**

```
  ./check-ebs-burst-limit --aws_region=eu-west-1
  
```

**check-ebs-snapshots**

```
  ./check-ebs-snapshots --check_ignored=false
  
```

**check-ec2-cpu_balance**

```
  ./check-ec2-cpu_balance --critical=3
  
  ./check-ec2-cpu_balance --critical=1 --warning=5
  
  ./check-ec2-cpu_balance --critical=1 --warning=5 --tag=TESTING
  
```

**check-ec2-filter**

```
  ./check-ec2-filter --filters="{\"filters\" : [{\"name\" : \"instance-state-name\", \"values\": [\"running\"]}]}"
  
  ./check-ec2-filter --exclude_tags="{\"TAG_NAME\" : \"TAG_VALUE\"}" --compare=not
  
```

**check-ec2-network**

```
  ./check-ec2-network --instance_id=i-0f1626fsbfvbafa2 --direction=NetworkOut
  
```

**metrics-ec2-count**

```
  ./metrics-ec2-count --metric_type=status
  
  ./metrics-ec2-count --metric_type=instance
  
```

**metrics-ec2-filter**

```
  ./metric-ec2-filter --filters="{\"filters\" : [{\"name\" : \"instance-state-name\", \"values\": [\"running\"]}]}"
  
```

**check-elb-certs**

```
  ./check-elb-certs --aws_region=${your_region} --warning=${days_to_warn} -critical=${days_to_critical}
  
```

**check-elb-health-fog**

```
  ./check-elb-health-fog --aws_region=${you_region} --instances=${your_instance_ids} 
                         --elb_name=${your_elb_name} --verbose=true  
```

**check-elb-health-sdk**

```
  ./check-elb-health-sdk --aws_region=region
  
  ./check-elb-health-sdk --aws_region=region --elb_name=my-elb
  
  ./check-elb-health-sdk --aws_region=region --elb_name=my-elb --instances=instance1,instance2
  
```

**check-elb-instance-inservice**

```
  ./check-elb-instance-inservice --aws_region=${your_region}
  
  ./check-elb-instance-inservice --aws_region=${your_region} --elb_name=${LoadBalancerName}
  
```

**check-elb-latency**

```
  ./check-elb-latency --warning_over=1 --critical_over=3
  
  ./check-elb-latency --elb_names=app --critical_over=5 --statistics=maximum --period=3600
  
```

**check-elb-nodes**

```
  ./check-elb-nodes --warning=3 --critical=2 --load_balancer=#{your-load-balancer}

  ./check-elb-nodes --warning_percentage=50 --critical_percentage=25 --load_balancer=#{your-load-balancer}
  
```

**elb-metrics**

```
  ./elb-metrics --aws_region=${your_region}
  
```

**check-rds**

```
  ./check-rds --db_instance_id=sensu-admin-db --available_zone_severity=critical --available_zone=ap-northeast-1a
  
  ./check-rds --db_instance_id=sensu-admin-db --cpu_warning_over=80 --cpu_critical_over=90
  
  ./check-rds --db_instance_id=sensu-admin-db --cpu_critical_over=90 --statistics=maximum --period=3600
  
  ./check-rds --db_instance_id=sensu-admin-db --connections_critical_over=120 --connections_warning_over=100 
              --statistics=maximum --period=3600
              
  ./check-rds --db_instance_id=sensu-admin-db --iops_critical_over=200 --iops_warning_over=100 --period=300
  
  ./check-rds --db_instance_id=sensu-admin-db --memory_warning_over=80 --statistics=minimum --period=7200
  
  ./check-rds --db_instance_id=sensu-admin-db --disk_warning_over=80 --period=7200
  
  ./check-rds --db_instance_id=sensu-admin-db --cpu_warning_over=80 --cpu_critical_over=90 
              --memory_warning_over=60 --memory_critical_over=80
              
 ```
 
 **check-rds-events**
 
 ```
  ./check-rds-events --aws_region=${your_region}  --db_instance_id=${your_rds_instance_id_name}
  
  ./check-rds-events.rb --aws_region=${your_region}
  
  ```
  
 **check-rds-pending**
  
 ```
   ./check-rds-pending --aws_region=${you_region}

 ```

 **rds-metrics**
 
 ```
  ./rds-metrics --aws_region=eu-west-1
  
  ./rds-metrics --aws_region=eu-west-1 --db_instance_id=sr2x8pbti0eon1
  
 ```
 
 **check-s3-bucket**
 
 ```
  ./check-s3-bucket --bucket_name=mybucket
  
 ```
 
 **check-s3-bucket-visibility**
 
 ```
  ./check-s3-bucket-visibility.go --exclude_buckets_regx=sensu --bucket_names=ssensu-ec2,sensu-ec3 
                                  --exclude_cuckets=sensu-ec3
 ```
 
 **check-s3-object**
 
 ```
  ./check-s3-object --bucket_name=aws-testing --key_prefix=s3
  
 ```
  
 **check-s3-tag**
 
 ```
  ./check-s3-tag --tag_keys=sensu
  
 ```
 
 **s3-metrics**
 
 ```
  ./metrics-s3
  
 ```

## Configuration

### Asset registration

Assets are the best way to make use of this handler. If you're not using an asset, please consider doing so! If you're using Sensu 5.13 or later, you can use the following command to add the asset: 

`sensuctl asset add sensu/sensu-aws`

If you're using an earlier version of Sensu, you can download the asset definition from [this project's Bonsai Asset Index page](https://bonsai.sensu.io/assets/sensu/sensu-aws).

### Check definition

```yaml
---
type: CheckConfig
api_version: core/v2
metadata:
  name: check-ec2-network
  namespace: default
spec:
  check_hooks: null
  command: check-ec2-network -instance_id i-1234567890
  env_vars: null
  handlers: []
  high_flap_threshold: 0
  interval: 10
  low_flap_threshold: 0
  output_metric_format: ""
  output_metric_handlers: null
  proxy_entity_name: ""
  publish: true
  round_robin: false
  runtime_assets:
  - sensu/sensu-aws
  stdin: false
  subdue: null
  subscriptions:
  - aws
  timeout: 0
  ttl: 0
```

### On-disk configuration

This plugin requires on-disk configuration for any host that will run checks from this collection. See the Sample Credential Configuration below to create the file you need to operate this plugin.

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

## Installation from source

The preferred way to install and deploy this plugin is to use it as an [asset][2]. To compile and install the plugin from source or contribute to the plugin, download the latest version of the sensu-CHANGEME from [releases][1] or create an executable script from this source.

From the local path of the sensu-CHANGEME repository:

```
go build -o /usr/local/bin/sensu-CHANGEME main.go
```
## Contributing
For more information about contributing to this plugin, see https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md

[1]: https://github.com/sensu/sensu-aws/releases
[2]: #asset-registration
