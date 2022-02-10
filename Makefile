down:
	docker compose down

rebuild:
	docker compose down; docker compose build; docker compose up -d

rebuild/debug:
	docker compose -f docker-compose-debug.yml down; docker compose -f docker-compose-debug.yml build; docker compose -f docker-compose-debug.yml up -d

rebuild/prod:
	docker compose -f docker-compose-prod.yml down && docker compose -f docker-compose-prod.yml build && docker compose -f docker-compose-prod.yml up -d