package req

type Result struct {
	ID    int    `json:"id"`
	Value string `json:"result"`
}

type ExpressionRequest struct {
	Expression string `json:"expression"`
}
