package main

import (
	"testing"

	"gopkg.in/yaml.v2"
)

const metadataSample = `
name: "Guestbook"
maintainer: "Flexiant Ltd."
email: "contact@flexiant.com"
description: "Installs/Configures Guestbook Example"
version: "0.10.4"

rc:
  # redis-master: "redis-master-controller.yaml"
  # redis-slave: "redis-slave-controller.yaml"
  frontend: "frontend-service.yaml"

svc:
  redis-master: "redis-master-service.yaml"
  redis-slave: "redis-slave-service.yaml"
  frontend: "frontend-controller.yaml"


attributes:
  svc:
    frontend:
      balancer:
        description: "Defines how we want to expose the Frontend Service"
        default: LoadBalancer
        required: true
      port:
        description: "Defines expose port for the Frontend Service"
        default: 80
        required: true
  rc:
    frontend:
      number:
        default: 6
        required: true
      name:
        default: "php-redis"
        required: true
      image:
        default: "kubernetes/example-guestbook-php-redis"
        required: true
      version:
        default: "v2"
        required: true
`

func TestMetadataParsingModel(t *testing.T) {
	var metadata Metadata
	err := yaml.Unmarshal([]byte(metadataSample), &metadata)
	if err != nil {
		t.Errorf("error parsing metadata: %v", err)
	}
}

func TestAttributeDefaults(t *testing.T) {
	var metadata Metadata
	err := yaml.Unmarshal([]byte(metadataSample), &metadata)
	if err != nil {
		t.Fatalf("error parsing metadata: %v", err)
	}
	defaults, err := metadata.AttributeDefaults()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, err := defaults.GetString("rc/frontend/name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if val != "php-redis" {
		t.Errorf("unexpected value: %v", val)
	}
}
