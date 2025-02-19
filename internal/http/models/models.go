package models


type Expression struct {
	Id         int    `json:"id"`
	Expression string `json:"expression"`
}

type Answer struct {
	Id int `json:"id"`
}