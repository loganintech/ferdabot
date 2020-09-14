
silent:
	docker-compose -f docker-compose.yml up --detach

up:
	docker-compose -f docker-compose.yml up

down:
	docker-compose -f docker-compose.yml down