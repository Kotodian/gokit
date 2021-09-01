package api

type Response struct {
	Status    int    `json:"status"`
	Rows      int    `json:"rows"`
	Code      string `json:"code"`
	Msg       string `json:"msg"`
	Timestamp int64  `json:"timestamp"`
}
