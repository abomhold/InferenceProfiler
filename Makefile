BIN_DIR         := bin
GO_BINARY       := $(BIN_DIR)/infpro
GO_MAIN         := main.go
GO_SOURCES      := main.go $(shell find pkg -name '*.go' 2>/dev/null)
TOFU_LOCK       := .terraform.lock.hcl
SSH_KEY         := ../infpro.pem
SSH             := ssh -i $(SSH_KEY) \
					   -o StrictHostKeyChecking=no \
					   -o UserKnownHostsFile=/dev/null \
					   -o LogLevel=ERROR


include configs/default.env
-include configs/cloud.env
-include ../secrets.env
export

.DELETE_ON_ERROR:
MAKEFLAGS += --no-print-directory
.PHONY: all help clean build build-local refresh \
        infra-init infra-plan infra-up infra-up-client infra-up-server infra-down infra-down-client infra-down-server \
        infra-clean \
        deploy deploy-env deploy-client deploy-server ssh-server ssh-client \
        restart-services restart-vllm restart-infpro restart-nodes \
        pull-snapshot pull-results

all: help

help:
	@echo 'Usage: make [target]'
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

refresh: ## Tidy, format and vet
	go mod tidy && go fmt ./... && go vet ./...

build-docker: $(GO_SOURCES) ## Compile Go binary for Linux amd64 locally (might need a C cross-compilation setup)
	@mkdir -p $(BIN_DIR)
	docker run --rm --user $(shell id -u):$(shell id -g) \
		-e GOPATH=/tmp/go -e GOCACHE=/tmp/go-cache \
		-v $(CURDIR):/app -w /app golang:latest \
		sh -c "CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o $@ $(GO_MAIN)"
	chmod +x $@

build: $(GO_BINARY) ## Compile Go binary for Linux amd64 using an ephemeral Docker container
$(GO_BINARY): $(GO_SOURCES)
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o $(GO_BINARY) $(GO_MAIN)
	chmod +x $(GO_BINARY)

infra-init: $(TOFU_LOCK) ## Initialize Terraform (only on first run or if lockfile is missing)
$(TOFU_LOCK):
	tofu init

infra-plan: infra-init ## Show execution plan
	tofu plan

infra-up: infra-init build ## Provision infrastructure
	tofu apply -auto-approve

infra-down: ## Destroy cluster
	tofu destroy -auto-approve

infra-clean: infra-down ## Destroy + remove state
	rm -rf .terraform *.tfstate *.tfstate.backup $(TOFU_LOCK)

deploy: ## Push binaries, scripts, env, and service files to all nodes
	$(MAKE) deploy-env
	$(MAKE) deploy-server
	$(MAKE) deploy-client
	$(MAKE) restart-services

deploy-env: ## Deploy experiment env to both nodes
	@cat configs/default.env configs/cloud.env ../secrets.env > /tmp/experiment.env
	@for ip in $(shell tofu output -raw all_ips | tr ',' ' '); do \
  		echo "-> Deploying combined env file to $$ip"; \
  		rsync -az --rsync-path="sudo rsync" -e "$(SSH)" /tmp/experiment.env ubuntu@$$ip:/opt/infpro/; \
	done
	@rm -f /tmp/experiment.env

deploy-client: ## Deploy scripts to client node
	rsync -az --rsync-path="sudo rsync" -e "$(SSH)" scripts/ ubuntu@$(shell tofu output -raw client_ip):/opt/infpro/scripts/
	$(SSH) ubuntu@$(shell tofu output -raw client_ip) "sudo ln -sf /opt/infpro/scripts/start_bench /usr/local/bin/start_bench"

deploy-server: build ## Deploy binaries, env, and systemd files to server nodes
	@echo "-> Deploying binary to server"
	@rsync -az --mkpath --rsync-path="sudo rsync" -e "$(SSH)" bin/ ubuntu@$(shell tofu output -raw server_ip):/usr/local/bin
	@echo "-> Deploying services files to server"
	@rsync -az --mkpath --rsync-path="sudo rsync" -e "$(SSH)" configs/systemd/ ubuntu@$(shell tofu output -raw server_ip):/etc/systemd/system/

restart-services: ## Restart infpro + vllm on server nodes
	@echo "-> Starting services on server"
	@$(SSH) ubuntu@$(shell tofu output -raw server_ip) "sudo systemctl daemon-reload"
	$(MAKE) restart-vllm
	$(MAKE) restart-infpro

restart-vllm: 
	@echo "-> Starting vllm on server"
	@$(SSH) ubuntu@$(shell tofu output -raw server_ip) "sudo systemctl enable vllm && sudo systemctl restart vllm"

restart-infpro: 
	@echo "-> Starting infpro on server"
	@$(SSH) ubuntu@$(shell tofu output -raw server_ip) "sudo systemctl enable infpro && sudo systemctl restart infpro"

restart-nodes: ## Reboot all server nodes
	@for ip in $(shell tofu output -raw all_ips | tr ',' ' '); do \
	  echo "-> Rebooting $$ip"; \
	  $(SSH) ubuntu@$$ip "sudo reboot" & \
	done; wait; echo "Waiting 30s..."; sleep 30

ssh-server: ## SSH into server node
	$(SSH) ubuntu@$(shell tofu output -raw server_ip)

ssh-client: ## SSH into client node
	$(SSH) ubuntu@$(shell tofu output -raw client_ip)

pull-snapshot: ## Pull latest snapshot from server to local machine
	curl $(shell tofu output -raw server_ip):$(INFPRO_PORT)/snapshot | jq 'walk(if type == "object" and has("V") and has("T") then "V: \(.V) | T: \(.T)" else . end)' | less

pull-results: ## Pull benchmark results from client to local machine
	@mkdir -p results
	rsync -avz --progress -e "$(SSH)" ubuntu@$(shell tofu output -raw client_ip):/opt/infpro/results/ results/

clean: infra-clean ## Remove build artifacts
	rm -rf $(BIN_DIR) coverage.out *.parquet *.jsonl *.prof
