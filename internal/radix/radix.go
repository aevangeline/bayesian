// Package radix implements a radix tree for storing a collection of strings processing
package radix

import (
	"errors"
	"sort"
	"strings"
)

// Tree represents how we can interface with our specialized radix tree
type Tree interface {
	Insert(needle string, category int) error
	Find(needle string) ([]int, bool)
	GetTotals() []int
	CategoryCount() int
	UniqueWords() int
}

type radixRoot struct {
	numCategories  int
	categoryTotals []int
	uniqueWords    int
	root           *radixNode
}

type radixChild struct {
	prefix string
	node   *radixNode
}

type radixNode struct {
	values   []int
	isLeaf   bool
	children []radixChild
}

var ErrOutOfBoundsCategory = errors.New("radix: out of bounds category")
var ErrInvalidCategoryCount = errors.New("radix: invalid category count")
var ErrNoSuchNode = errors.New("radix: no such node")

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

	return &radixRoot{
		numCategories:  numCategories,
		categoryTotals: make([]int, numCategories, numCategories),
		root:           &radixNode{isLeaf: false, values: make([]int, numCategories, numCategories)},
	}, nil
}

// Insert creates or finds a node representing this string in this radix tree and increments the category
func (r *radixRoot) Insert(needle string, category int) error {
	if category >= r.numCategories {
		return ErrOutOfBoundsCategory
	}

	node, isNew := r.findOrCreate(needle)
	if node != nil {
		if node.values == nil {
			node.values = make([]int, r.numCategories, r.numCategories)
		}

		node.values[category] += 1

		if isNew {
			r.uniqueWords += 1
		}

		r.categoryTotals[category] += 1
	}

	return nil
}

// Find gets the category values associated with a given string
func (r *radixRoot) Find(needle string) ([]int, bool) {
	node := r.find(needle)
	if node == nil {
		return nil, false
	}

	return node.values, true
}

// GetTotals fetches the totals associated with each category
func (r *radixRoot) GetTotals() []int {
	return r.categoryTotals
}

// CategoryCount returns the number of categories we are tracking in this tree
func (r *radixRoot) CategoryCount() int {
	return r.numCategories
}

// UniqueWords returns the number of words represented in this trie
func (r *radixRoot) UniqueWords() int {
	return r.uniqueWords
}

func longestCommonPrefix(left, right string) string {
	shorter := []rune(left)
	longer := []rune(right)
	if len(longer) < len(shorter) {
		temp := shorter
		shorter = longer
		longer = temp
	}
	shared := ""
	for i := range shorter {
		if shorter[i] != longer[i] {
			break
		}
		shared += string(shorter[i])
	}

	return shared
}

func searchChildren(children []radixChild, needle string) (int, matchType, string) {
	// here we handle the degenerate case of no children to make the rest of the function simpler
	numLeaves := len(children)
	if numLeaves == 0 {
		return 0, nomatch, ""
	}

	idx := sort.Search(len(children), func(i int) bool {
		return children[i].prefix >= needle
	})

	if idx < numLeaves {
		// here we handle getting an exact match
		if children[idx].prefix == needle {
			return idx, exact, needle
		}

		lcp := longestCommonPrefix(children[idx].prefix, needle)
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

	lcp := longestCommonPrefix(children[idx-1].prefix, needle)

	// if it is a substring, then report that to the user
	if lcp == children[idx-1].prefix {
		return idx - 1, substring, lcp
	} else if lcp != "" {
		return idx - 1, shared_prefix, lcp
	}
	// otherwise we have no match
	return idx, nomatch, ""

}

// findNode searches through the tree and finds the node that represents this string, if it exists
func (r *radixRoot) find(needle string) *radixNode {
	current := r.root
	remainder := needle

	// we loop until we either find the correct node, or we definitively cannot find it
	for {
		if remainder == "" {
			if current.isLeaf {
				return current
			}
			return nil
		}

		idx, match, lcp := searchChildren(current.children, remainder)
		if match == exact || match == substring {
			current = current.children[idx].node
			remainder = strings.TrimPrefix(remainder, lcp)
		} else {
			return nil
		}
	}
}

// inserts a new leaf at the specified index
func insertChild(children []radixChild, newLeaf radixChild, idx int) []radixChild {
	children = append(children, radixChild{})
	copy(children[idx+1:], children[idx:])
	children[idx] = newLeaf
	return children
}

// findOrCreate returns either an existing node representing the string, or creates a new one, the bool reports whether the node is new
func (r *radixRoot) findOrCreate(needle string) (*radixNode, bool) {
	current := r.root
	remainder := needle
	// we loop until we find either a node where we need to insert our string, or a node that already represents it
	for {
		if remainder == "" {
			current.isLeaf = true
			return current, false
		}
		idx, match, lcp := searchChildren(current.children, remainder)
		// if we find an exact match for the key, or just a substring prefix, we just keep looping
		if match == exact || match == substring {
			current = current.children[idx].node
			remainder = strings.TrimPrefix(remainder, lcp)

		} else if match == shared_prefix {
			// if there's a shared prefix, we replace the prefix on the child with the lcp and then add children for those
			previousKey := current.children[idx].prefix
			// compute the suffixes for the new nodes
			oldNodeKey := strings.TrimPrefix(previousKey, lcp)
			needleKey := strings.TrimPrefix(needle, lcp)

			// pull out the nodes we will have for our new radix nodes
			oldNode := current.children[idx].node
			newNode := &radixNode{isLeaf: true}

			// sort the children of the new super node
			newChildren := []radixChild{{prefix: oldNodeKey, node: oldNode}, {prefix: needleKey, node: newNode}}
			sort.Slice(newChildren, func(i int, j int) bool {
				return newChildren[i].prefix < newChildren[j].prefix
			})
			current.children[idx] = radixChild{prefix: lcp, node: &radixNode{children: newChildren}}
			return newNode, true
		} else if match == super {
			suffix := strings.TrimPrefix(current.children[idx].prefix, lcp)
			newNode := &radixNode{isLeaf: true, children: []radixChild{{prefix: suffix, node: current.children[idx].node}}}
			current.children[idx] = radixChild{prefix: lcp, node: newNode}
			return newNode, true
		} else {
			newNode := &radixNode{isLeaf: true}
			// if there's no match, we just insert the child in sorted order
			current.children = insertChild(current.children, radixChild{prefix: remainder, node: newNode}, idx)
			return newNode, true
		}
	}
}
