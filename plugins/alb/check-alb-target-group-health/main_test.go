package main

import (
	"errors"
	"io/ioutil"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/stretchr/testify/mock"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

type elbClient struct {
	mock.Mock
}

func (e *elbClient) DescribeTargetGroups(input *elbv2.DescribeTargetGroupsInput) (*elbv2.DescribeTargetGroupsOutput, error) {
	args := e.Called(input)
	out, _ := args.Get(0).(*elbv2.DescribeTargetGroupsOutput)
	return out, args.Error(1)
}

func (e *elbClient) DescribeTargetHealth(input *elbv2.DescribeTargetHealthInput) (*elbv2.DescribeTargetHealthOutput, error) {
	args := e.Called(input)
	out, _ := args.Get(0).(*elbv2.DescribeTargetHealthOutput)
	return out, args.Error(1)
}

func TestCheckHealth(t *testing.T) {
	tests := []struct {
		Name       string
		Groups     []string
		ClientFunc func(testing.TB) *elbClient
		MockExpect func(testing.TB, *mock.Mock)
		Critical   bool
		ExpStatus  int
		ExpError   bool
	}{
		{
			Name: "no groups specified",
			ClientFunc: func(t testing.TB) *elbClient {
				t.Helper()
				return new(elbClient)
			},
			MockExpect: func(testing.TB, *mock.Mock) {},
			ExpStatus:  2,
			ExpError:   true,
		},
		{
			Name: "DescribeTargetGroups error",
			Groups: []string{
				"group A",
				"group B",
			},
			ClientFunc: func(t testing.TB) *elbClient {
				t.Helper()
				client := new(elbClient)
				client.On("DescribeTargetGroups", mock.Anything).Return(nil, errors.New("error"))
				return client
			},
			MockExpect: func(t testing.TB, client *mock.Mock) {
				t.Helper()
				client.AssertCalled(t, "DescribeTargetGroups", mock.Anything)
				client.AssertNotCalled(t, "DescribeTargetHealth", mock.Anything)
			},
			ExpStatus: 2,
			ExpError:  true,
		},
		{
			Name: "DescribeTargetHealth error",
			Groups: []string{
				"group A",
				"group B",
			},
			ClientFunc: func(t testing.TB) *elbClient {
				t.Helper()
				groupNameA := "group A"
				groupNameB := "group B"
				client := new(elbClient)
				client.On("DescribeTargetGroups", mock.Anything).Return(&elbv2.DescribeTargetGroupsOutput{
					TargetGroups: []*elbv2.TargetGroup{
						{
							TargetGroupName: &groupNameA,
						},
						{
							TargetGroupName: &groupNameB,
						},
					},
				}, nil)
				client.On("DescribeTargetHealth", mock.Anything).Return(nil, errors.New("error"))
				return client
			},
			MockExpect: func(t testing.TB, client *mock.Mock) {
				t.Helper()
				client.AssertCalled(t, "DescribeTargetGroups", mock.Anything)
				client.AssertNumberOfCalls(t, "DescribeTargetHealth", 1)
			},
			ExpStatus: 2,
			ExpError:  true,
		},
		{
			Name: "Some target groups are unhealthy, non-critical",
			Groups: []string{
				"group A",
				"group B",
			},
			ClientFunc: func(t testing.TB) *elbClient {
				t.Helper()
				client := new(elbClient)
				groupNameA := "group A"
				groupNameB := "group B"
				client.On("DescribeTargetGroups", mock.Anything).Return(&elbv2.DescribeTargetGroupsOutput{
					TargetGroups: []*elbv2.TargetGroup{
						{
							TargetGroupName: &groupNameA,
							TargetGroupArn:  &groupNameA,
						},
						{
							TargetGroupName: &groupNameB,
							TargetGroupArn:  &groupNameB,
						},
					},
				}, nil)
				healthy := "healthy"
				unhealthy := "unhealthy"
				inputA := &elbv2.DescribeTargetHealthInput{
					TargetGroupArn: &groupNameA,
				}
				inputB := &elbv2.DescribeTargetHealthInput{
					TargetGroupArn: &groupNameB,
				}
				client.On("DescribeTargetHealth", inputA).Return(&elbv2.DescribeTargetHealthOutput{
					TargetHealthDescriptions: []*elbv2.TargetHealthDescription{
						{
							TargetHealth: &elbv2.TargetHealth{
								State: &healthy,
							},
							Target: &elbv2.TargetDescription{
								Id: &groupNameA,
							},
						},
					},
				}, nil)
				client.On("DescribeTargetHealth", inputB).Return(&elbv2.DescribeTargetHealthOutput{
					TargetHealthDescriptions: []*elbv2.TargetHealthDescription{
						{
							TargetHealth: &elbv2.TargetHealth{
								State: &unhealthy,
							},
							Target: &elbv2.TargetDescription{
								Id: &groupNameB,
							},
						},
					},
				}, nil)
				return client
			},
			MockExpect: func(t testing.TB, client *mock.Mock) {
				t.Helper()
				client.AssertCalled(t, "DescribeTargetGroups", mock.Anything)
				client.AssertNumberOfCalls(t, "DescribeTargetHealth", 2)
			},
			ExpStatus: 1,
			ExpError:  true,
		},
		{
			Name: "Some target groups are unhealthy, critical",
			Groups: []string{
				"group A",
				"group B",
			},
			ClientFunc: func(t testing.TB) *elbClient {
				t.Helper()
				client := new(elbClient)
				groupNameA := "group A"
				groupNameB := "group B"
				client.On("DescribeTargetGroups", mock.Anything).Return(&elbv2.DescribeTargetGroupsOutput{
					TargetGroups: []*elbv2.TargetGroup{
						{
							TargetGroupName: &groupNameA,
							TargetGroupArn:  &groupNameA,
						},
						{
							TargetGroupName: &groupNameB,
							TargetGroupArn:  &groupNameB,
						},
					},
				}, nil)
				healthy := "healthy"
				unhealthy := "unhealthy"
				inputA := &elbv2.DescribeTargetHealthInput{
					TargetGroupArn: &groupNameA,
				}
				inputB := &elbv2.DescribeTargetHealthInput{
					TargetGroupArn: &groupNameB,
				}
				client.On("DescribeTargetHealth", inputA).Return(&elbv2.DescribeTargetHealthOutput{
					TargetHealthDescriptions: []*elbv2.TargetHealthDescription{
						{
							TargetHealth: &elbv2.TargetHealth{
								State: &healthy,
							},
							Target: &elbv2.TargetDescription{
								Id: &groupNameA,
							},
						},
					},
				}, nil)
				client.On("DescribeTargetHealth", inputB).Return(&elbv2.DescribeTargetHealthOutput{
					TargetHealthDescriptions: []*elbv2.TargetHealthDescription{
						{
							TargetHealth: &elbv2.TargetHealth{
								State: &unhealthy,
							},
							Target: &elbv2.TargetDescription{
								Id: &groupNameB,
							},
						},
					},
				}, nil)
				return client
			},
			MockExpect: func(t testing.TB, client *mock.Mock) {
				t.Helper()
				client.AssertCalled(t, "DescribeTargetGroups", mock.Anything)
				client.AssertNumberOfCalls(t, "DescribeTargetHealth", 2)
			},
			Critical:  true,
			ExpStatus: 2,
			ExpError:  true,
		},
		{
			Name: "target groups are healthy",
			Groups: []string{
				"group A",
				"group B",
			},
			ClientFunc: func(t testing.TB) *elbClient {
				t.Helper()
				client := new(elbClient)
				groupNameA := "group A"
				groupNameB := "group B"
				client.On("DescribeTargetGroups", mock.Anything).Return(&elbv2.DescribeTargetGroupsOutput{
					TargetGroups: []*elbv2.TargetGroup{
						{
							TargetGroupName: &groupNameA,
						},
						{
							TargetGroupName: &groupNameB,
						},
					},
				}, nil)
				client.On("DescribeTargetHealth", mock.Anything).Return(&elbv2.DescribeTargetHealthOutput{}, nil)
				return client
			},
			MockExpect: func(t testing.TB, client *mock.Mock) {
				t.Helper()
				client.AssertCalled(t, "DescribeTargetGroups", mock.Anything)
				client.AssertNumberOfCalls(t, "DescribeTargetHealth", 2)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := test.ClientFunc(t)
			status, err := checkHealth(client, test.Groups, test.Critical)
			if got, want := status, test.ExpStatus; got != want {
				t.Errorf("bad status: got %d, want %d", got, want)
			}
			if got, want := (err != nil), test.ExpError; got != want {
				t.Errorf("conflicting error expectations: got (err != nil) == %v, want %v", got, want)
			}
			test.MockExpect(t, &client.Mock)
		})
	}
}
