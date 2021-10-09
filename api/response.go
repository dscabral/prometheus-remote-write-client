package api

import "net/http"

type promRes struct {
	created bool
}

func (p promRes) Code() int {
	if p.created {
		return http.StatusCreated
	}
	return http.StatusOK
}

func (p promRes) Headers() map[string]string {
	return map[string]string{}
}

func (p promRes) Empty() bool {
	return false
}
