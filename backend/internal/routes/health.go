package routes

func (r *Routes) Health() {
	health := r.app.Group("/health")

	health.Get("", r.handlers.HealthStatus)
}
