/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 */

package mantis

type AWSResource string

const (
	S3Bucket AWSResource = "S3Bucket"
	DynamoDB AWSResource = "DynamoDB"
	SNS      AWSResource = "SNS"
)
