/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 */

package mantis

type DOResource string

const (
	Droplet      DOResource = "Droplet"
	Volume       DOResource = "Volume"
	FloatingIP   DOResource = "FloatingIP"
	Snapshot     DOResource = "Snapshot"
	Image        DOResource = "Image"
	LoadBalancer DOResource = "LoadBalancer"
	Firewall     DOResource = "Firewall"
	Database     DOResource = "Database"
	Domain       DOResource = "Domain"
	SSH          DOResource = "SSH"
	VPC          DOResource = "VPC"
	Project      DOResource = "Project"
	Kubernetes   DOResource = "Kubernetes"
	App          DOResource = "App"
	CDN          DOResource = "CDN"
	Certificate  DOResource = "Certificate"
	Registry     DOResource = "Registry"
	Monitoring   DOResource = "Monitoring"
	Billing      DOResource = "Billing"
	Balance      DOResource = "Balance"
	Invoice      DOResource = "Invoice"
	Action       DOResource = "Action"
	Tag          DOResource = "Tag"
	Size         DOResource = "Size"
	Region       DOResource = "Region"
)

func (r DOResource) String() string {
	return string(r)
}
