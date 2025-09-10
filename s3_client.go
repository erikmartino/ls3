package main

import (
	"context"
	"io"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Global cache for bucket regions and region-specific clients
var (
	bucketRegionCache = make(map[string]string)
	regionClientCache = make(map[string]*s3.Client)
	cacheMutex        sync.RWMutex
)

// ClientManager manages region-specific S3 clients
type ClientManager struct {
	defaultClient S3Client
}

// NewClientManager creates a new client manager
func NewClientManager(defaultClient S3Client) *ClientManager {
	return &ClientManager{
		defaultClient: defaultClient,
	}
}

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
	// Check cache first
	cacheMutex.RLock()
	if region, exists := bucketRegionCache[bucketName]; exists {
		cacheMutex.RUnlock()
		return region, nil
	}
	cacheMutex.RUnlock()

	// Not in cache, fetch from AWS
	result, err := client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
		Bucket: &bucketName,
	})
	if err != nil {
		return "", err
	}

	// AWS returns empty string for us-east-1 region
	region := string(result.LocationConstraint)
	if region == "" {
		region = "us-east-1"
	}

	// Cache the result
	cacheMutex.Lock()
	bucketRegionCache[bucketName] = region
	cacheMutex.Unlock()

	return region, nil
}

// GetClientForBucket returns a region-specific client for the bucket
func (cm *ClientManager) GetClientForBucket(ctx context.Context, bucketName string) (S3Client, error) {
	// Get the bucket's region (uses cache)
	region, err := getBucketRegion(ctx, cm.defaultClient, bucketName)
	if err != nil {
		// If we can't get the region, fall back to default client
		return cm.defaultClient, nil
	}

	// Check if we already have a client for this region
	cacheMutex.RLock()
	if client, exists := regionClientCache[region]; exists {
		cacheMutex.RUnlock()
		return client, nil
	}
	cacheMutex.RUnlock()

	// Create a new region-specific client
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		// If we can't create a region-specific client, fall back to default
		return cm.defaultClient, nil
	}

	regionClient := s3.NewFromConfig(cfg)

	// Cache the client
	cacheMutex.Lock()
	regionClientCache[region] = regionClient
	cacheMutex.Unlock()

	return regionClient, nil
}

// clearCache clears the internal caches (for testing)
func clearCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	bucketRegionCache = make(map[string]string)
	regionClientCache = make(map[string]*s3.Client)
}
