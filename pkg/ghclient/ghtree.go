package ghclient

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Interim object for holding tree data from api calls in expected format
type childRef struct {
	Path string `json:"path"`
	Type string `json:"type"`
	Sha  string `json:"sha"`
}
type treeMarshal struct {
	Sha  string     `json:"sha"`
	Tree []childRef `json:"tree"`
}

type Node interface {
	GetParent() Node
	GetChildren() []Node
	SetChild(Node)
	Write(io.Writer) error

	setParent(Node)
}

type Tree struct {
	Sha  string `json:"sha"`
	Path string `json:"path"`

	parent   Node
	children []Node
}

// NewTreeFromJSON tree factory
func NewTreeFromJSON(tJSON []byte) (*Tree, error) {
	tree := &Tree{}
	err := json.Unmarshal(tJSON, tree)
	if err != nil {
		return nil, fmt.Errorf("error building tree from json: %s", err)
	}
	return tree, nil
}

func (t *Tree) GetParent() Node {
	return t.parent
}

func (t *Tree) GetChildren() []Node {
	return t.children
}

func (t *Tree) setParent(parent Node) {
	t.parent = parent
}

func (t *Tree) SetChild(child Node) {
	child.setParent(t)
	t.children = append(t.children, child)
}

func (t *Tree) Write(w io.Writer) error {
	return nil
}

func recursivePrint(t Node) (string, error) {
	var sb strings.Builder
	sb.WriteString("\n")
	tJ, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	sb.Write(tJ)

	for _, child := range t.GetChildren() {
		res, err := recursivePrint(child)
		if err != nil {
			return "", err
		}
		sb.WriteString(" ")
		sb.WriteString(res)
	}

	return sb.String(), nil
}

//Blob encoded contents of a file
type Blob struct {
	Sha      string `json:"sha"`
	Encoding string `json:"encoding"`
	Content  string `json:"content"`

	Path   string `json:"Path"`
	parent Node
}

// NewBlobFromJSON blob factory
func NewBlobFromJSON(bJSON []byte) (*Blob, error) {
	blob := &Blob{}
	err := json.Unmarshal(bJSON, blob)
	if err != nil {
		return nil, fmt.Errorf("error building blob from json: %s", err)
	}
	return blob, nil
}

func (b *Blob) GetParent() Node {
	return b.parent
}

func (b *Blob) GetChildren() []Node {
	return nil
}

func (b *Blob) setParent(parent Node) {
	b.parent = parent
}

func (b *Blob) SetChild(child Node) {
	return
}

func (b *Blob) Write(w io.Writer) error {
	return nil
}

//Cache one way cache for tracking github trees
type Cache struct {
	trees map[string]*Tree
	blobs map[string]*Blob
}

//NewCache create cache
func NewCache() Cache {
	return Cache{
		trees: make(map[string]*Tree),
		blobs: make(map[string]*Blob),
	}
}

func (c *Cache) GetBlob(sha string) *Blob {
	return c.blobs[sha]
}

func (c *Cache) GetTree(sha string) *Tree {
	return c.trees[sha]
}

func (c *Client) buildTree(parent Node, treeMarsh treeMarshal, repo Repository) error {
	if parent == nil {
		return nil
	}

	for _, cRef := range treeMarsh.Tree {
		switch cRef.Type {
		case "blob":
			child := c.cache.GetBlob(cRef.Sha)
			if child != nil {
				parent.SetChild(child)
				continue
			}

			blobJSON, err := c.api.GetBlob(repo.Owner.Login, repo.Name, cRef.Sha)
			if err != nil {
				return err
			}

			child, err = NewBlobFromJSON(blobJSON)
			if err != nil {
				return err
			}

			child.Path = cRef.Path

			parent.SetChild(child)

		case "tree":
			child := c.cache.GetTree(cRef.Sha)
			if child != nil {
				parent.SetChild(child)
				continue
			}

			treeJSON, err := c.api.GetTree(repo.Owner.Login, repo.Name, cRef.Sha)
			if err != nil {
				return err
			}

			child, err = NewTreeFromJSON(treeJSON)
			if err != nil {
				return err
			}

			child.Path = cRef.Path

			parent.SetChild(child)

			err = json.Unmarshal(treeJSON, &treeMarsh)
			if err != nil {
				return fmt.Errorf("error while parsing tree json: %s", err)
			}

			err = c.buildTree(child, treeMarsh, repo)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetTree checks cache for tree, else pulls tree from github
func (c *Client) GetTree(sha string, repo Repository) (*Tree, error) {
	t := c.cache.GetTree(sha)

	if t != nil {
		return t, nil
	}

	treeJSON, err := c.api.GetTree(repo.Owner.Login, repo.Name, sha)
	if err != nil {
		return nil, err
	}

	treeMarsh := treeMarshal{}
	err = json.Unmarshal(treeJSON, &treeMarsh)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling tree json: %s", err)
	}

	top := &Tree{
		Sha:      treeMarsh.Sha,
		Path:     treeMarsh.Sha,
		parent:   nil,
		children: []Node{},
	}

	err = c.buildTree(top, treeMarsh, repo)
	if err != nil {
		return nil, err
	}
	return top, nil
}