SHELL:=/bin/bash
UNAME_S := $(shell uname -s)

container_registry_host = asia.gcr.io
gcp_project = chechaichang
operator_name = cattle-operator

git_branch_name = $(shell git rev-parse --abbrev-ref HEAD)
git_tag_name = $(shell git describe --tags --abbrev=0)
ifdef CIRCLE_TAG
  git_branch_name = ${CIRCLE_TAG}
  git_tag_name = ${CIRCLE_TAG}
endif
git_commit_sha = $(shell git rev-parse --short HEAD)
image_name = $(container_registry_host)/$(gcp_project)/$(operator_name)
image_version = $(git_branch_name)-$(git_commit_sha)

.PHONY: build tag generate

mod:
	go mod download
	go mod vendor

generate:
	operator-sdk generate k8s

local: mod generate
	OPERATOR_NAME=$(operator_name) operator-sdk up local --namespace=default

build: generate
	operator-sdk build $(image_name):$(image_tag)

minikube:
	minikube start

push: build
	docker push $(image_name):$(image_tag)

ifeq ($(UNAME_S),Linux)
tag:
	sed -i 's|image:.*|image: $(image_name):$(image_tag)|g' deploy/operator.yaml
endif
ifeq ($(UNAME_S),Darwin)
tag:
	sed -i "" 's|image:.*|image: $(image_name):$(image_tag)|g' deploy/operator.yaml
endif

apply-operator: tag
	kubectl apply -f deploy/service_account.yaml
	kubectl apply -f deploy/role.yaml
	kubectl apply -f deploy/role_binding.yaml
	kubectl apply -f deploy/operator.yaml

apply-crd: tag
	kubectl apply -f deploy/crds

delete:
	kubectl delete -f deploy/crds
	kubectl delete -f deploy/service_account.yaml
	kubectl delete -f deploy/role.yaml
	kubectl delete -f deploy/role_binding.yaml
	kubectl delete -f deploy/operator.yaml
