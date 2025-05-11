package req

type Result struct {
	ID     int    `json:"id"`
	Value  any    `json:"result"`
	UserID uint64 `json:"user_id"`
}

type ExpressionRequest struct {
	Expression string `json:"expression"`
}
