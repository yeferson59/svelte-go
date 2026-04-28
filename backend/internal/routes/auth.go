package routes

func (r *Routes) Auth() {
	r.app.Post("/login", r.handlers.Login)
	r.app.Use(r.middlewares.JWT())
}
