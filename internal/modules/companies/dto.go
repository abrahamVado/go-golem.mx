package companies

type UpdateCompanyRequest struct {
	Name string `json:"name" binding:"required"`
}
