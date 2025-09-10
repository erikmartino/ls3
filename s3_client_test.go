package main

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// mockS3Client is a mock implementation of the S3Client interface for testing.
type mockS3Client struct {
	ListBucketsFunc       func(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
	ListObjectsV2Func     func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	GetObjectFunc         func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	GetBucketLocationFunc func(ctx context.Context, params *s3.GetBucketLocationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error)
}

func (m *mockS3Client) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	return m.ListBucketsFunc(ctx, params, optFns...)
}

func (m *mockS3Client) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	return m.ListObjectsV2Func(ctx, params, optFns...)
}

func (m *mockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return m.GetObjectFunc(ctx, params, optFns...)
}

func (m *mockS3Client) GetBucketLocation(ctx context.Context, params *s3.GetBucketLocationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error) {
	return m.GetBucketLocationFunc(ctx, params, optFns...)
}

func TestGetBuckets(t *testing.T) {
	mockClient := &mockS3Client{
		ListBucketsFunc: func(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
			bucketName := "test-bucket"
			return &s3.ListBucketsOutput{
				Buckets: []types.Bucket{
					{Name: &bucketName},
				},
			}, nil
		},
	}

	buckets, err := getBuckets(context.TODO(), mockClient)
	if err != nil {
		t.Fatalf("getBuckets returned an error: %v", err)
	}

	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(buckets))
	}

	if *buckets[0].Name != "test-bucket" {
		t.Errorf("expected bucket name 'test-bucket', got '%s'", *buckets[0].Name)
	}
}

func TestListS3Objects(t *testing.T) {
	mockClient := &mockS3Client{
		ListObjectsV2Func: func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
			return &s3.ListObjectsV2Output{
				CommonPrefixes: []types.CommonPrefix{
					{Prefix: aws.String("folder/")},
				},
				Contents: []types.Object{
					{Key: aws.String("file.txt")},
				},
			}, nil
		},
	}

	objects, err := listS3Objects(context.TODO(), mockClient, "test-bucket", "")
	if err != nil {
		t.Fatalf("listS3Objects returned an error: %v", err)
	}

	if len(objects.CommonPrefixes) != 1 {
		t.Fatalf("expected 1 common prefix, got %d", len(objects.CommonPrefixes))
	}

	if *objects.CommonPrefixes[0].Prefix != "folder/" {
		t.Errorf("expected common prefix 'folder/', got '%s'", *objects.CommonPrefixes[0].Prefix)
	}

	if len(objects.Contents) != 1 {
		t.Fatalf("expected 1 object, got %d", len(objects.Contents))
	}

	if *objects.Contents[0].Key != "file.txt" {
		t.Errorf("expected object key 'file.txt', got '%s'", *objects.Contents[0].Key)
	}
}

func TestGetObjectContent(t *testing.T) {
	content := "hello world"
	mockClient := &mockS3Client{
		GetObjectFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
			return &s3.GetObjectOutput{
				Body: io.NopCloser(strings.NewReader(content)),
			}, nil
		},
	}

	body, err := getObjectContent(context.TODO(), mockClient, "test-bucket", "file.txt")
	if err != nil {
		t.Fatalf("getObjectContent returned an error: %v", err)
	}

	if string(body) != content {
		t.Errorf("expected content '%s', got '%s'", content, string(body))
	}
}

func TestGetBucketRegion(t *testing.T) {
	clearCache() // Clear cache before test
	mockClient := &mockS3Client{
		GetBucketLocationFunc: func(ctx context.Context, params *s3.GetBucketLocationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error) {
			return &s3.GetBucketLocationOutput{
				LocationConstraint: types.BucketLocationConstraint("us-west-2"),
			}, nil
		},
	}

	region, err := getBucketRegion(context.TODO(), mockClient, "test-bucket")
	if err != nil {
		t.Fatalf("getBucketRegion returned an error: %v", err)
	}

	if region != "us-west-2" {
		t.Errorf("expected region 'us-west-2', got '%s'", region)
	}
}

func TestGetBucketRegionUsEast1(t *testing.T) {
	clearCache() // Clear cache before test
	mockClient := &mockS3Client{
		GetBucketLocationFunc: func(ctx context.Context, params *s3.GetBucketLocationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error) {
			return &s3.GetBucketLocationOutput{
				LocationConstraint: "",
			}, nil
		},
	}

	region, err := getBucketRegion(context.TODO(), mockClient, "test-bucket")
	if err != nil {
		t.Fatalf("getBucketRegion returned an error: %v", err)
	}

	if region != "us-east-1" {
		t.Errorf("expected region 'us-east-1', got '%s'", region)
	}
}
