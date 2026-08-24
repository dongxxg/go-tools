package httpx

import (
	"testing"
)

func TestRequest(t *testing.T) {
	Request("Get", "https://gorest.co.in/public-api/users", nil)
	// assert.Nil(t, err)
	// assert.NotNil(t, ret)

	// Request("Get", "https://jsonplaceholder.typicode.com/todos/1", nil)

	Request("Get", "https://gorest.co.in/public-api/todos", nil)
	// Request("Get", "https://jsonplaceholder.typicode.com/todos/2", nil)
}
