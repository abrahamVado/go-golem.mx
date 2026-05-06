package permissions

type CreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}
type UpdateRequest struct {
	Name string `json:"name"`
}
