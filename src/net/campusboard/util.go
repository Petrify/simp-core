package campusboard

import (
	"golang.org/x/net/html"
)

func condTypeText() func(*html.Node) bool {
	return func(n *html.Node) bool {
		return n.Type == html.TextNode
	}
}

func condTypeElement() func(*html.Node) bool {
	return func(n *html.Node) bool {
		return n.Type == html.ElementNode
	}
}

func condAnd(A func(*html.Node) bool, B func(*html.Node) bool) func(*html.Node) bool {
	return func(n *html.Node) bool {
		return A(n) && B(n)
	}
}

func condOr(A func(*html.Node) bool, B func(*html.Node) bool) func(*html.Node) bool {
	return func(n *html.Node) bool {
		return A(n) || B(n)
	}
}

func condHasData(data string) func(*html.Node) bool {
	return func(n *html.Node) bool {
		return n.Data == data
	}
}

func condHasAttr(key string, val string) func(*html.Node) bool {
	return func(n *html.Node) bool {
		for _, a := range n.Attr {
			if a.Key == key {
				if a.Val == val {
					return true
				}
			}
		}
		return false
	}
}

func toUtf8(iso8859_1Buf []byte) string {
	buf := make([]rune, len(iso8859_1Buf))
	for i, b := range iso8859_1Buf {
		buf[i] = rune(b)
	}
	return string(buf)
}

func FindNode(root *html.Node, condition func(*html.Node) bool) *html.Node {
	if condition(root) {
		return root
	}
	for c := root.FirstChild; c != nil; c = c.NextSibling {
		node := FindNode(c, condition)
		if node != nil {
			return node
		}
	}
	return nil
}

//Finds all nodes matching the condition. Does not search children of nodes that match the condition
func FindAll(root *html.Node, condition func(*html.Node) bool) []*html.Node {

	lst := make([]*html.Node, 0, 10)
	lst = _findAll(root, condition, lst)
	return lst
}

func _findAll(n *html.Node, condition func(*html.Node) bool, lst []*html.Node) []*html.Node {

	if condition(n) {
		lst = append(lst, n)
		return lst
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		lst = _findAll(c, condition, lst)
	}

	return lst
}

func SelectFieldToMap(selNode *html.Node) map[string]string {
	//get each option node
	optionNodes := FindAll(selNode, condAnd(condTypeElement(), condHasData("option")))

	//parse from nodes to map
	m := make(map[string]string)
	for _, n := range optionNodes {
		for _, a := range n.Attr {
			if a.Key == "value" {
				m[n.FirstChild.Data] = a.Val
			}
		}
	}
	return m
}
