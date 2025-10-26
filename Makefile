.PHONY: run down cluster app destroy

run:
	docker compose up --build

down:
	docker compose down

cluster:
	kind create cluster --name cluster-nuitee --config infra/kind-config.yaml

app:
	cd infra && terraform apply --auto-approve

destroy:
	kind delete cluster --name cluster-nuitee
	cd infra && rm -f terraform.tfstate*