package routes

func (r *Routes) Users() {
	users := r.app.Group("/users")

	users.Get("", r.handlers.GetListUsers)
	users.Get("/:id", r.handlers.GetUserByID)
}
