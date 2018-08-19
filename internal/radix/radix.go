// Package radix implements a radix tree for storing a collection of strings processing
package radix

import (
	"encoding/gob"
	"errors"
	"sort"
	"strings"
	"unicode/utf8"
)

// Tree represents how we can interface with our specialized radix tree
type Tree interface {
	Insert(needle string, category int) error
	Find(needle string) ([]int, bool)
	GetTotals() []int
	CategoryCount() int
	UniqueWords() int
}

type root struct {
	NumCategories    int
	CategoryTotals   []int
	UniqueWordsCount int
	Root             *node
}

type child struct {
	Prefix string
	Node   *node
}

type node struct {
	Values   []int
	IsLeaf   bool
	Children []child
}

var ErrOutOfBoundsCategory = errors.New("radix: out of bounds category")
var ErrInvalidCategoryCount = errors.New("radix: invalid category count")
var ErrNoSuchNode = errors.New("radix: no such node")
var ErrCannotCreateNode = errors.New("radix: no node created")

type matchType string

const (
	nomatch       matchType = "nomatch"       // nomatch means that there is no match
	shared_prefix matchType = "shared_prefix" // means that the there is a non-empty shared prefix between these numbers
	substring     matchType = "substring"     // substring means that this match is substring of the needle
	exact         matchType = "exact"         // exact means that this search result is an exact match for the needle
	super         matchType = "super"         // super means that this search result is a super string of the needle
)

// New creates a new instance of a radix tree
func New(numCategories int) (Tree, error) {
	if numCategories <= 0 {
		return nil, ErrInvalidCategoryCount
	}

	return &root{
		NumCategories:  numCategories,
		CategoryTotals: make([]int, numCategories, numCategories),
		Root:           &node{IsLeaf: false, Values: make([]int, numCategories, numCategories)},
	}, nil
}

// Insert creates or finds a node representing this string in this radix tree and increments the category
func (r *root) Insert(needle string, category int) error {
	if category >= r.NumCategories {
		return ErrOutOfBoundsCategory
	}

	node, isNew := r.findOrCreate(needle)
	if node != nil {
		if node.Values == nil {
			node.Values = make([]int, r.NumCategories, r.NumCategories)
		}

		node.Values[category] += 1

		if isNew {
			r.UniqueWordsCount += 1
		}

		r.CategoryTotals[category] += 1

		return nil
	}

	return ErrCannotCreateNode
}

// Find gets the category values associated with a given string
func (r *root) Find(needle string) ([]int, bool) {
	node := r.find(needle)
	if node == nil {
		return nil, false
	}

	return node.Values, true
}

// GetTotals fetches the totals associated with each category
func (r *root) GetTotals() []int {
	return r.CategoryTotals
}

// CategoryCount returns the number of categories we are tracking in this tree
func (r *root) CategoryCount() int {
	return r.NumCategories
}

// UniqueWords returns the number of words represented in this trie
func (r *root) UniqueWords() int {
	return r.UniqueWordsCount
}

func longestCommonPrefix(left, right string) string {
	if utf8.RuneCountInString(left) > utf8.RuneCountInString(right) {
		temp := left
		left = right
		right = temp
	}

	end := 0
	for i, r := range left {
		other, width := utf8.DecodeRuneInString(right[i:])
		if other == r {
			end = i + width
		} else {
			break
		}

	}

	return left[:end]
}

func init() {
	gob.Register(&root{})
}

func searchChildren(children []child, needle string) (int, matchType, string) {
	// here we handle the degenerate case of no children to make the rest of the function simpler
	numLeaves := len(children)
	if numLeaves == 0 {
		return 0, nomatch, ""
	}

	idx := sort.Search(len(children), func(i int) bool {
		return children[i].Prefix >= needle
	})

	if idx < numLeaves {
		// here we handle getting an exact match
		if children[idx].Prefix == needle {
			return idx, exact, needle
		}

		lcp := longestCommonPrefix(children[idx].Prefix, needle)
		// if it's not an exact match, it might be a strict super string
		if lcp == needle {
			return idx, super, lcp
		} else if lcp != "" {
			return idx, shared_prefix, lcp
		}
	}

	//if we are at the beginning of the children, we can't check before us
	if idx == 0 {
		return 0, nomatch, ""
	}

	lcp := longestCommonPrefix(children[idx-1].Prefix, needle)

	// if it is a substring, then report that to the user
	if lcp == children[idx-1].Prefix {
		return idx - 1, substring, lcp
	} else if lcp != "" {
		return idx - 1, shared_prefix, lcp
	}
	// otherwise we have no match
	return idx, nomatch, ""

}

// findNode searches through the tree and finds the node that represents this string, if it exists
func (r *root) find(needle string) *node {
	current := r.Root
	remainder := needle

	// we loop until we either find the correct node, or we definitively cannot find it
	for {
		if remainder == "" {
			if current.IsLeaf {
				return current
			}
			return nil
		}

		idx, match, lcp := searchChildren(current.Children, remainder)
		if match == exact || match == substring {
			current = current.Children[idx].Node
			remainder = strings.TrimPrefix(remainder, lcp)
		} else {
			return nil
		}
	}
}

// inserts a new leaf at the specified index
func insertChild(children []child, newLeaf child, idx int) []child {
	children = append(children, child{})
	copy(children[idx+1:], children[idx:])
	children[idx] = newLeaf
	return children
}

// findOrCreate returns either an existing node representing the string, or creates a new one, the bool reports whether the node is new
func (r *root) findOrCreate(needle string) (*node, bool) {
	current := r.Root
	remainder := needle
	// we loop until we find either a node where we need to insert our string, or a node that already represents it
	for {
		if remainder == "" {
			current.IsLeaf = true
			return current, false
		}
		idx, match, lcp := searchChildren(current.Children, remainder)
		// if we find an exact match for the key, or just a substring prefix, we just keep looping
		if match == exact || match == substring {
			current = current.Children[idx].Node
			remainder = strings.TrimPrefix(remainder, lcp)

		} else if match == shared_prefix {
			// if there's a shared prefix, we replace the prefix on the child with the lcp and then add children for those
			previousKey := current.Children[idx].Prefix
			// compute the suffixes for the new nodes
			oldNodeKey := strings.TrimPrefix(previousKey, lcp)
			remainderKey := strings.TrimPrefix(remainder, lcp)

			// pull out the nodes we will have for our new radix nodes
			oldNode := current.Children[idx].Node
			newNode := &node{IsLeaf: true}

			// sort the children of the new super node
			newChildren := []child{{Prefix: oldNodeKey, Node: oldNode}, {Prefix: remainderKey, Node: newNode}}
			sort.Slice(newChildren, func(i int, j int) bool {
				return newChildren[i].Prefix < newChildren[j].Prefix
			})
			current.Children[idx] = child{Prefix: lcp, Node: &node{Children: newChildren}}
			return newNode, true
		} else if match == super {
			suffix := strings.TrimPrefix(current.Children[idx].Prefix, lcp)
			newNode := &node{IsLeaf: true, Children: []child{{Prefix: suffix, Node: current.Children[idx].Node}}}
			current.Children[idx] = child{Prefix: lcp, Node: newNode}
			return newNode, true
		} else {
			newNode := &node{IsLeaf: true}
			// if there's no match, we just insert the child in sorted order
			current.Children = insertChild(current.Children, child{Prefix: remainder, Node: newNode}, idx)
			return newNode, true
		}
	}
}
