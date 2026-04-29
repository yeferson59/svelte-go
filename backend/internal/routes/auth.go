package routes

func (r *Routes) Auth() {
	auth := r.app.Group("/auth")

	auth.Post("/login", r.handlers.Login)
	auth.Use(r.middlewares.JWT())
}
