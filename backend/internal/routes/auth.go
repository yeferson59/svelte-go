package routes

func (r *Routes) Auth() {
	auth := r.app.Group("/auth")

	auth.Post("/register", r.handlers.Register)
	auth.Post("/login", r.handlers.Login)
	auth.Use(r.middlewares.JWT())
	auth.Get("/session", r.handlers.GetSession)
}
