.PHONY: run kill-backend kill-frontend down

run:
	@cd backend && make db && make run &
	@cd frontend && pnpm run dev --open

kill-backend:
	-lsof -ti:8080 | xargs kill || true


down:
	@cd backend && make down
	@$(MAKE) kill-backend
