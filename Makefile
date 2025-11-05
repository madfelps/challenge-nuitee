.PHONY: run down cluster app destroy apply expose

run:
	docker compose up --build

down:
	docker compose down

cluster:
	kind create cluster --name cluster-nuitee --config infra/kind-config.yaml

apply:
	kubectl apply -f infra/manifests/namespace.yaml
	@sleep 1
	kubectl apply -f infra/manifests

expose:
	@kubectl port-forward -n nuitee-challenge svc/api-service 4000:80 &
	@kubectl port-forward -n nuitee-challenge svc/postgres-service 5432:5432 &

app:
	cd infra && terraform apply --auto-approve

destroy:
	kind delete cluster --name cluster-nuitee
	cd infra && rm -f terraform.tfstate*