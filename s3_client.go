package main

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Client defines the interface for S3 client operations.
type S3Client interface {
	ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	GetBucketLocation(ctx context.Context, params *s3.GetBucketLocationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error)
}

func getBuckets(ctx context.Context, client S3Client) ([]types.Bucket, error) {
	result, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}
	return result.Buckets, nil
}

func listS3Objects(ctx context.Context, client S3Client, bucketName, prefix string) (*s3.ListObjectsV2Output, error) {
	delimiter := "/"
	input := &s3.ListObjectsV2Input{
		Bucket:    &bucketName,
		Delimiter: &delimiter,
	}
	if prefix != "" {
		input.Prefix = &prefix
	}
	return client.ListObjectsV2(ctx, input)
}

func getObjectContent(ctx context.Context, client S3Client, bucketName, objectKey string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	}
	result, err := client.GetObject(ctx, input)
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

func getBucketRegion(ctx context.Context, client S3Client, bucketName string) (string, error) {
	result, err := client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
		Bucket: &bucketName,
	})
	if err != nil {
		return "", err
	}

	// AWS returns empty string for us-east-1 region
	if result.LocationConstraint == "" {
		return "us-east-1", nil
	}
	return string(result.LocationConstraint), nil
}
