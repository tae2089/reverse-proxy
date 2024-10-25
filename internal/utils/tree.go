package utils

import "strings"

type node struct {
	path           string
	isPathVariable bool
	children       map[string]*node
	variableChild  *node
}

// Tree struct represents the entire tree with a root node
type Tree struct {
	root *node
}

// NewTree creates a new tree and initializes the root node
func NewTree() *Tree {
	return &Tree{
		root: &node{
			path:     "/",
			children: make(map[string]*node),
		},
	}
}

// Insert function inserts a path into the tree
func (t *Tree) Insert(path string) {
	segments := strings.Split(path, "/")[1:]
	current := t.root

	for _, segment := range segments {
		isPathVariable := strings.HasPrefix(segment, "{")
		var next *node
		if isPathVariable {
			if current.variableChild == nil {
				current.variableChild = &node{
					path:           segment,
					isPathVariable: true,
					children:       make(map[string]*node),
				}
			}
			next = current.variableChild
		} else {
			if _, exists := current.children[segment]; !exists {
				current.children[segment] = &node{
					path:           segment,
					isPathVariable: false,
					children:       make(map[string]*node),
				}
			}
			next = current.children[segment]
		}
		current = next
	}
}

// Search function searches for a pattern and returns the pattern path
func (t *Tree) Search(pattern string) string {
	segments := strings.Split(pattern, "/")[1:]
	current := t.root
	var result []string

	for _, segment := range segments {
		if child, exists := current.children[segment]; exists {
			result = append(result, child.path)
			current = child
		} else if current.variableChild != nil {
			result = append(result, current.variableChild.path)
			current = current.variableChild
		} else {
			break
		}
	}
	if len(result) == 0 {
		return "no match"
	}
	return "/" + strings.Join(result, "/")
}

// ReplaceWithPattern function replaces matched pattern in string with the pattern
func (t *Tree) ReplaceWithPattern(input, pattern string) string {
	if pattern == "no match" {
		return input
	}
	inputSegments := strings.Split(input, "/")[1:]
	patternSegments := strings.Split(pattern, "/")[1:]
	var result []string = append(patternSegments, inputSegments[len(patternSegments):]...)
	return "/" + strings.Join(result, "/")
}
