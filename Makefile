
silent:
	docker-compose -f docker-compose.yml up --detatch

up:
	docker-compose -f docker-compose.yml up

down:
	docker-compose -f docker-compose.yml down