package routes

func (r *Routes) Health() {
	r.app.Get("/health", r.handlers.HealthStatus)
}
