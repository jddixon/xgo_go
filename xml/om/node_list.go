package om

// xgo/xml/om/node_list.go

import ()

// A container for Nodes.  Each Holder (Document or Element) has a
// NodeList, but the reverse is not necessarily true.
//
type NodeList struct {
	// list of child nodes
	nodes  []NodeI
	holder ElementI  // immediate parent, might be nil
	doc    DocumentI // ultimate parent, might be nil
}

// Create an empty node list.
//
func NewNewNodeList() *NodeList {
	var nodes []NodeI
	return &NodeList{
		nodes: nodes,
	}
}

// Create a node list with only one member.
//
func NewNodeList(node NodeI) *NodeList {
	nodes := []NodeI{node}
	return &NodeList{
		nodes: nodes,
	}
}

// Add a Node to the NodeList.
//
// XXX Should check for cycles; if the Holder is a document,
// XXX there may be only one Element node.
//
// Returns NilNode if the Node argument is nil.
//
func (nl *NodeList) Append(node NodeI) (err error) {

	if node == nil {
		err = NilNode
	} else {
		node.SetHolder(nl.holder)
		nl.nodes = append(nl.nodes, node)
	}
	return
}

// Pointless synonym for Append() ?
//
func (nl *NodeList) AddChild(node NodeI) error {
	return nl.Append(node)
}

// Copy the nodes from another NodeList into this one, then
// delete them from the source, to ease GC.
//
func (nl *NodeList) MoveFrom(otherList *NodeList) (this *NodeList, err error) {
	if otherList == nil {
		err = EmptyOtherList
	} else {
		for i := uint(0); i < otherList.Size(); i++ {
			var node NodeI
			node, err = otherList.Get(i)
			if err != nil {
				break
			}
			node.SetHolder(nl.holder)
			nl.nodes = append(nl.nodes, node)
		}
	}
	if err == nil {
		otherList.Clear()
	}
	return
}

// Make a NodeList empty.
//
func (nl *NodeList) Clear() {
	var nodes []NodeI
	nl.nodes = nodes
}

// Insert a node into a list at a specific zero-based point.  Will
// return an error if the node argument is nil or the index n is
// out of bounds.
//
func (nl *NodeList) Insert(n uint, node *Node) (err error) {
	if n > nl.Size() {
		err = IndexOutOfBounds
	}
	if err == nil && node == nil {
		err = NilNode
	}
	if err == nil {
		node.SetHolder(nl.holder)
		if n == nl.Size() {
			nl.nodes = append(nl.nodes, node)
		} else {
			head := nl.nodes[0:n]
			tail := nl.nodes[n:]
			nl.nodes = append(head, node)
			nl.nodes = append(nl.nodes, tail...)
		}
	}
	return
}

// Return whether there are no nodes in the list.
//
func (nl *NodeList) IsEmpty() bool {
	return len(nl.nodes) == 0
}

// Return the Nth node in the list.  Will return an error if the
// index n is out of bounds.
//
func (nl *NodeList) Get(n uint) (node NodeI, err error) {
	if n >= nl.Size() {
		err = IndexOutOfBounds
	} else {
		node = nl.nodes[n]
	}
	return
}

// Return number of nodes in the list.
//
func (nl *NodeList) Size() uint {
	return uint(len(nl.nodes))
}

// PROPERTIES ///////////////////////////////////////////////////

// Return the immediate parent of this list.
//
func (nl *NodeList) GetHolder() ElementI {
	return nl.holder
}

// Change the immediate parent of this list, here and in
// descendent nodes.
//
// XXX SHOULD CHECK FOR GRAPH CYCLES
//
// @param h the new parent; may be nil
//
func (nl *NodeList) SetHolder(h ElementI) {
	var doc DocumentI
	if h != nil {
		doc = h.GetDocument()
	}
	for i := uint(0); i < nl.Size(); i++ {
		node, _ := nl.Get(i)
		node.SetHolder(h)
		node.SetDocument(doc)
	}
}

// VISITOR-RELATED///////////////////////////////////////////////

// Take the visitor through every node in the list, recursing.
//
func (nl *NodeList) WalkAll(v VisitorI) (err error) {
	for i := uint(0); i < nl.Size(); i++ {
		node, _ := nl.Get(i)
		err = node.WalkAll(v)
		if err != nil {
			break
		}
	}
	return
}

// Take the Visitor through the list, visiting any node which is
// a Holder, recursively.  Used when you don't want to visit, for
// example, attributes.
//
func (nl *NodeList) WalkHolders(v VisitorI) (err error) {
	for i := uint(0); err == nil && i < nl.Size(); i++ {
		var n NodeI
		n, err = nl.Get(i)
		if err == nil {
			switch w := n.(type) {
			case *Element:
				err = w.WalkHolders(v)
			}
		}
	}
	return
}

// SERIALIZATION METHODS ////////////////////////////////////////

// A String containing each of the Nodes in XML form, recursively,
// without indenting.
//
func (nl *NodeList) ToXml() (s string) {
	for i := uint(0); i < nl.Size(); i++ {
		var node NodeI
		node, _ = nl.Get(i)
		s += node.ToXml()
	}
	return
}
