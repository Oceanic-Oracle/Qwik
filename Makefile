.PHONY: run

run:
	docker build -t migrator ./Backend/infra/DbNomad/DbNomad
	docker build -t patroni ./Backend/infra/Patroni
	docker compose -f ./Backend/compose.yml up --build