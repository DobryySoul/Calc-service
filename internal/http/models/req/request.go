package req

type Result struct {
	ID    int     `json:"id"`
	Value float64 `json:"result"`
}

type ExpressionRequest struct {
	Expression string `json:"expression"`
}
