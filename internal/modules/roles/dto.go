package roles

type CreateRequest struct {
	Name string `json:"name" binding:"required"`
}
type UpdateRequest struct {
	Name string `json:"name"`
}
