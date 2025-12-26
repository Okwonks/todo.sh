package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Okwonks/go-todo/internal/model"
)

type Client struct {
	baseURL string
	http *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http: &http.Client{},
	}
}

func (c *Client) CreateTodo(t *model.Todo) (*model.Todo, error) {
	body, _ := json.Marshal(t)

	resp, err := c.http.Post(c.baseURL + "/api/v1/todos", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var created model.Todo
	err = json.NewDecoder(resp.Body).Decode(&created)
	return &created, err
}

func (c *Client) List() ([]model.Todo, error) {
	resp, err := c.http.Get(c.baseURL + "/api/v1/todos")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var todos []model.Todo
	err = json.NewDecoder(resp.Body).Decode(&todos)
	return todos, err
}

func (c *Client) GetById(id int64) (*model.Todo, error) {
	resp, err := c.http.Get(fmt.Sprintf("%s/api/v1/todos/%d", c.baseURL, id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var todo model.Todo
	err = json.NewDecoder(resp.Body).Decode(&todo)
	return &todo, err
}

func (c *Client) Update(data map[string]any) (*model.Todo, error) {
	payload, _ := json.Marshal(data)
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/todos", c.baseURL), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var todo model.Todo
	err = json.NewDecoder(resp.Body).Decode(&todo)
	return &todo, err
}

func (c *Client) Delete(id int64) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/todos/%d", c.baseURL, id), nil)
	if err != nil {
		return err
	}
	
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
