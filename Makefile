.PHONY: services
services:
	docker-compose -f docker-compose.services.yml up

.PHONY: services-down
services-down:
	docker-compose -f docker-compose.services.yml down