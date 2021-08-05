package beego

type DataTablesRequest struct {
	Order  string `json:"order"`
	Sort   string `json:"sort"`
	Limit  int    `json:"limit"`
	Search string `json:"search"`
	Offset int    `json:"offset"`
}
