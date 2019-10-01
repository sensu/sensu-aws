package utils

import "github.com/aws/aws-sdk-go/service/s3"

func SortContents(contents []*s3.Object) {
	var (
		n      = len(contents)
		sorted = false
	)
	for !sorted {
		swapped := false
		for i := 0; i < n-1; i++ {
			if (*contents[i].LastModified).Before(*contents[i+1].LastModified) {
				contents[i+1], contents[i] = contents[i], contents[i+1]
				swapped = true
			}
		}
		if !swapped {
			sorted = true
		}
		n = n - 1
	}
}
