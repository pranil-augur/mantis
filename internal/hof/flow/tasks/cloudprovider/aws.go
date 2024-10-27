/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 */

package cloudprovider

import (
	"context"
	"fmt"
	"io"
	"path"

	"cuelang.org/go/cue"
	"gocloud.dev/docstore"
	"gocloud.dev/pubsub"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
	"github.com/opentofu/opentofu/internal/hof/lib/mantis"
)

type AWSTask struct{}

func NewAWSTask(val cue.Value) (hofcontext.Runner, error) {
	return &AWSTask{}, nil
}

func (t *AWSTask) Run(ctx *hofcontext.Context) (any, error) {
	v := ctx.Value
	configValue := v.LookupPath(cue.ParsePath("config"))

	// Query resources based on the configuration
	resources, err := queryAWSResources(context.Background(), configValue)
	if err != nil {
		return nil, fmt.Errorf("failed to query AWS resources: %v", err)
	}

	newV := v.FillPath(cue.ParsePath(mantis.MantisTaskOuts), resources)
	return newV, nil
}

type AWSResourceFilters struct {
	Name   string `json:"name"`
	Region string `json:"region"`
	Tag    string `json:"tag"`
	Status string `json:"status"`
	Size   string `json:"size"`
}

func queryAWSResources(ctx context.Context, config cue.Value) (map[string]interface{}, error) {
	resourceTypeStr, err := config.LookupPath(cue.ParsePath("resourceType")).String()
	if err != nil {
		return nil, fmt.Errorf("failed to get resource type: %v", err)
	}

	resourceType := mantis.AWSResource(resourceTypeStr)

	filters := AWSResourceFilters{}
	if filtersValue := config.LookupPath(cue.ParsePath("filters")); filtersValue.Exists() {
		if err := filtersValue.Decode(&filters); err != nil {
			return nil, fmt.Errorf("failed to decode filters: %v", err)
		}
	}

	switch resourceType {
	case mantis.S3Bucket:
		return queryS3Buckets(ctx, config, filters)
	case mantis.DynamoDB:
		return queryDynamoDB(ctx, config, filters)
	case mantis.SNS:
		return querySNS(ctx, config)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

func queryS3Buckets(ctx context.Context, config cue.Value, filters AWSResourceFilters) (map[string]interface{}, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(filters.Region), // Use the region from filters if specified
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %v", err)
	}

	svc := s3.New(sess)
	result, err := svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list S3 buckets: %v", err)
	}

	var bucketNames []string
	for _, bucket := range result.Buckets {
		if matchesS3BucketFilters(*bucket.Name, filters) {
			bucketNames = append(bucketNames, *bucket.Name)
		}
	}

	return map[string]interface{}{"buckets": bucketNames}, nil
}

func matchesS3BucketFilters(bucketName string, filters AWSResourceFilters) bool {
	// Implement filtering logic for bucket names based on filters
	if filters.Name != "" {
		matched, _ := path.Match(filters.Name, bucketName)
		if !matched {
			return false
		}
	}
	return true
}

func queryDynamoDB(ctx context.Context, config cue.Value, filters AWSResourceFilters) (map[string]interface{}, error) {
	tableURL, err := config.LookupPath(cue.ParsePath("table")).String()
	if err != nil {
		return nil, fmt.Errorf("failed to get table URL: %v", err)
	}

	coll, err := docstore.OpenCollection(ctx, tableURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open collection: %v", err)
	}
	defer coll.Close()

	var items []map[string]interface{}
	iter := coll.Query().Get(ctx)
	defer iter.Stop()

	for {
		var item map[string]interface{}
		err := iter.Next(ctx, &item)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get next item: %v", err)
		}
		if matchesDynamoDBFilters(item, filters) {
			items = append(items, item)
		}
	}

	return map[string]interface{}{"items": items}, nil
}

func matchesDynamoDBFilters(item map[string]interface{}, filters AWSResourceFilters) bool {
	// Implement filtering logic for DynamoDB items based on filters
	return true // Placeholder, implement actual logic
}

func querySNS(ctx context.Context, config cue.Value) (map[string]interface{}, error) {
	topicURL, err := config.LookupPath(cue.ParsePath("topic")).String()
	if err != nil {
		return nil, fmt.Errorf("failed to get topic URL: %v", err)
	}

	// Open topic using the Go CDK portable interface
	topic, err := pubsub.OpenTopic(ctx, topicURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open topic: %v", err)
	}
	defer topic.Shutdown(ctx)

	return map[string]interface{}{"topic": topicURL}, nil
}
