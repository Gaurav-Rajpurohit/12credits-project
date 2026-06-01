.PHONY: help deploy rollback benchmark health-check clean

help:
	@echo "=========================================================="
	@echo " Zero-Downtime Blue-Green Framework Command Line Utility  "
	@echo "=========================================================="
	@echo "make deploy       - Execute automated blue-green rollout"
	@echo "make rollback     - Revert traffic and stop new container stack"
	@echo "make benchmark    - Run concurrent load testing & downtime audit"
	@echo "make health-check - Run diagnostic queries on running environments"
	@echo "make clean        - Stop and prune all Docker containers"
	@echo "=========================================================="

deploy:
	ansible-playbook ansible/deploy.yml

rollback:
	ansible-playbook ansible/rollback.yml

benchmark:
	cd tests && ./benchmark.sh

health-check:
	ansible-playbook ansible/health_check.yml

clean:
	cd docker && docker compose -f docker-compose.blue.yml down || true
	cd docker && docker compose -f docker-compose.green.yml down || true
