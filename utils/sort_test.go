package utils

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSortContents(t *testing.T) {
	contents := []*s3.Object{
		&s3.Object{Key: aws.String("foo_2019-Oct-01"), LastModified: aws.Time(time.Date(2019, 10, 1, 0, 0, 0, 0, time.UTC))},
		&s3.Object{Key: aws.String("foo_2019-Oct-02"), LastModified: aws.Time(time.Date(2019, 10, 2, 0, 0, 0, 0, time.UTC))},
		&s3.Object{Key: aws.String("foo_2019-Oct-03"), LastModified: aws.Time(time.Date(2019, 10, 3, 0, 0, 0, 0, time.UTC))},
	}

	SortContents(contents)

	assert.Equal(t, "foo_2019-Oct-03", aws.StringValue(contents[0].Key))
	assert.Equal(t, "foo_2019-Oct-02", aws.StringValue(contents[1].Key))
	assert.Equal(t, "foo_2019-Oct-01", aws.StringValue(contents[2].Key))
}
