
silent:
	docker-compose --detatch -f docker-compose.yml up

up:
	docker-compose -f docker-compose.yml up

down:
	docker-compose -f docker-compose.yml down