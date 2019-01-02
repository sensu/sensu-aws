package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/sreejita-biswas/aws-plugins/awsclient"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
)

/*
#
# check-elb-certs
#
# DESCRIPTION:
#   This plugin looks up all ELBs in the region and checks https
#   endpoints for expiring certificates
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#  ./check-elb-certs -aws_region=${your_region} -warning=${days_to_warn} -critical=${days_to_critical}
#
# NOTES:
#
# LICENSE:
#   TODO
#
*/

var (
	awsRegion string
	warning   int
	critical  int
	verbose   bool
	elbClient *elb.ELB
)

func checkExpiry() {
	awsSession := aws_session.CreateAwsSession()

	success, elbClient := awsclient.GetElbClient(awsSession)
	if !success {
		return
	}

	describeLoadBalancerInput := &elb.DescribeLoadBalancersInput{}
	describeLoadBalancerOutput, err := elbClient.DescribeLoadBalancers(describeLoadBalancerInput)
	if err != nil {
		fmt.Println("Error :", err)
		return
	}
	if !(describeLoadBalancerOutput != nil && describeLoadBalancerOutput.LoadBalancerDescriptions != nil && len(describeLoadBalancerOutput.LoadBalancerDescriptions) > 0) {
		return
	}
	for _, loadBalancer := range describeLoadBalancerOutput.LoadBalancerDescriptions {
		for _, listener := range loadBalancer.ListenerDescriptions {
			elbListener := listener.Listener
			if strings.ToUpper(*elbListener.Protocol) == "HTTPS" {
				dnsName := *loadBalancer.DNSName
				fmt.Println(dnsName)
				port := *elbListener.LoadBalancerPort
				ips, err := net.LookupIP(dnsName)
				if err != nil {
					fmt.Println("Error :", err)
					return
				}
				dialer := net.Dialer{}
				connection, err := tls.DialWithDialer(&dialer, "tcp", fmt.Sprintf("[%s]:%d", ips[0], port), &tls.Config{ServerName: dnsName})
				if err != nil {
					fmt.Println("Error :", err)
					return
				}
				for _, chain := range connection.ConnectionState().VerifiedChains {
					for _, cert := range chain {
						if cert.IsCA {
							continue
						}
						expiryDate := cert.NotAfter
						if critical > 0 && float64(critical) > (time.Since(expiryDate).Minutes()/float64(60))/float64(24.0) {
							fmt.Println(fmt.Sprintf("CRITICAL:Load Balancer Name:'%s' , Expiry Date:%s", *loadBalancer.LoadBalancerName, expiryDate.Format(time.RFC3339)))
							break
						}
						if warning > 0 && float64(warning) > (time.Since(expiryDate).Minutes()/float64(60))/float64(24.0) {
							fmt.Println(fmt.Sprintf("WARNING:Load Balancer Name:'%s' , Expiry Date:%s", *loadBalancer.LoadBalancerName, expiryDate.Format(time.RFC3339)))
							break
						}
					}
				}
			}
		}
	}
}

func main() {
	rootCmd := configureRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		_ = cmd.Help()
		return fmt.Errorf("invalid argument(s) received")
	}
	checkExpiry()
	return nil
}

func configureRootCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "check-elb-certs",
		Short: "The Sensu Go Aws Load Balancer handler for certificate expiry management",
		RunE:  run,
	}

	cmd.Flags().StringVar(&awsRegion, "aws_region", "us-west-1", "AWS Region (defaults to us-east-1).")
	cmd.Flags().IntVar(&warning, "warning", 30, "Warn on minimum number of days to SSL/TLS certificate expiration")
	cmd.Flags().IntVar(&critical, "critical", 5, "Minimum number of days to SSL/TLS certificate expiration")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Provide SSL/TLS certificate expiration details even when OK")

	return cmd
}
