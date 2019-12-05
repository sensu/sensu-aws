# aws-plugins

TravisCI: [![Build Status](https://travis-ci.org/sensu/sensu-aws.svg?branch=master)](https://travis-ci.org/sensu/sensu-aws)

## Functionality
- check-alb-target-group-health
- check-cloudwatch-alarm
- check-cloudwatch-alarms
- check-cloudwatch-composite-metric
- check-ebs-burst-limit
- check-ebs-snapshots
- check-ec2-cpu_balance
- check-ec2-filter
- check-ec2-network
- metrics-ec2-count
- metrics-ec2-filter
- check-elb-certs
- check-elb-health-fog
- check-elb-health-sdk
- check-elb-instance-inservice
- check-elb-latency
- check-elb-nodes
- check-elb-sum-requests
- elb-metrics
- check-rds
- check-rds-events
- check-rds-pending
- rds-metrics
- check-s3-bucket
- check-s3-bucket-visibility
- check-s3-object
- check-s3-tag
- s3-metrics

## Files

* /plugins/alb/check-alb-target-group-health/main.go
* /plugins/cloudwatch/check-cloudwatch-alarm/main.go
* /plugins/cloudwatch/check-cloudwatch-alarms/main.go
* /plugins/cloudwatch/check-cloudwatch-composite-metric/main.go
* /plugins/ebs/check-ebs-burst-limit/main.go
* /plugins/ebs/check-ebs-snapshots/main.go
* /plugins/ec2/check-ec2-cpu_balance/main.go
* /plugins/ec2/check-ec2-filter/main.go
* /plugins/ec2/check-ec2-network/main.go
* /plugins/ec2/metrics-ec2-count/main.go
* /plugins/ec2/metrics-ec2-filter/main.go
* /plugins/elb/check-elb-certs/main.go
* /plugins/elb/check-elb-health-fog/main.go
* /plugins/elb/check-elb-health-sdk/main.go
* /plugins/elb/check-elb-instances-inservice/main.go
* /plugins/elb/check-elb-latency/main.go
* /plugins/elb/check-elb-nodes/main.go
* /plugins/elb/check-elb-sum-requests/main.go
* /plugins/elb/metrics-elb/main.go
* /plugins/rds/check-rds/main.go
* /plugins/rds/check-rds-events/main.go
* /plugins/rds/check-rds-pending/main.go
* /plugins/rds/metrics-rds/main.go
* /plugins/s3/check-s3-bucket/main.go
* /plugins/s3/check-s3-bucket-visibility/main.go
* /plugins/s3/check-s3-object/main.go
* /plugins/s3/check-s3-tag/main.go
* plugins/s3/metrics-s3/main.go

## Binaries

* /bin/check-alb-target-group-health
* /bin/check-cloudwatch-alarm
* /bin/check-cloudwatch-alarms
* /bin/check-cloudwatch-composite-metric
* /bin/check-ebs-burst-limit
* /bin/check-ebs-snapshots
* /bin/check-ec2-cpu_balance
* /bin/check-ec2-filter
* /bin/check-ec2-network
* /bin/metrics-ec2-count
* /bin/metrics-ec2-filter
* /bin/check-elb-certs
* /bin/check-elb-health-fog
* /bin/check-elb-health-sdk
* /bin/check-elb-instances-inservice
* /bin/check-elb-latency
* /bin/check-elb-nodes
* /bin/check-elb-sum-requests
* /bin/metrics-elb
* /bin/check-rds
* /bin/check-rds-events
* /bin/check-rds-pending
* /bin/metrics-rds
* /bin/check-s3-bucket
* /bin/check-s3-bucket-visibility
* /bin/check-s3-object
* /bin/check-s3-tag
* /bin/metrics-s3

## Usage

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
