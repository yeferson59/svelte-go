package routes

import "github.com/gofiber/fiber/v3/middleware/paginate"

func (r *Routes) Users() {
	users := r.app.Group("/users")

	users.Get("", paginate.New(), r.handlers.GetListUsers)
	users.Get("/:id", r.handlers.GetUserByID)
}
