package routes

import "github.com/gofiber/fiber/v3/middleware/paginate"

func (r *Routes) Users() {
	users := r.router.Group("/users")

	users.Get("", paginate.New(), r.handlers.GetListUsers)
	users.Get("/:id", r.handlers.GetUserByID)
	users.Post("", r.handlers.CreateUser)
	users.Patch("/:id", r.handlers.UpdateUser)
	users.Delete("/:id", r.handlers.DeleteUser)
}
