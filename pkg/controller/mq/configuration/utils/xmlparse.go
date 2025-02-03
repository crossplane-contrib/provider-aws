package utils

import (
	"bytes"
	"encoding/xml"
	"sort"

	"github.com/google/go-cmp/cmp"
)

type Node struct {
	XMLName xml.Name
	Attr    []xml.Attr `xml:",any,attr"`
	Content []byte     `xml:",innerxml"`
	Nodes   []Node     `xml:",any"`
}

func (n *Node) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type node Node
	return d.DecodeElement((*node)(n), &start) //nolint:musttag
}

func DiffXMLConfigs(c1, c2 string) (string, error) {
	var configNode1 Node
	if err := decodeString(c1, &configNode1); err != nil {
		return "", err
	}
	var configNode2 Node
	if err := decodeString(c2, &configNode2); err != nil {
		return "", err
	}
	return diffNodes(configNode1, configNode2), nil
}

func decodeString(xmlString string, node *Node) error {
	buf := bytes.NewBuffer([]byte(xmlString))
	dec := xml.NewDecoder(buf)
	return dec.Decode(node)
}

func diffNodes(forProvider, atProvider Node) string {
	diff := cmp.Diff(forProvider, atProvider, nodeComparerOpt())
	return diff
}

// recursively check equality of nodes
// based on attributes, subnodes and tag name equality
func nodeComparerOpt() cmp.Option {
	return cmp.Comparer(func(x, y Node) bool {
		if x.XMLName.Local != y.XMLName.Local {
			return false
		}
		if len(x.Nodes) != len(y.Nodes) {
			return false
		}

		sort.Slice(x.Attr, func(i, j int) bool {
			return x.Attr[i].Name.Local < x.Attr[j].Name.Local
		})
		sort.Slice(y.Attr, func(i, j int) bool {
			return y.Attr[i].Name.Local < y.Attr[j].Name.Local
		})

		if !cmp.Equal(x.Attr, y.Attr) {
			return false
		}

		sort.Slice(x.Nodes, func(i, j int) bool {
			return x.Nodes[i].XMLName.Local < x.Nodes[j].XMLName.Local
		})
		sort.Slice(y.Nodes, func(i, j int) bool {
			return y.Nodes[i].XMLName.Local < y.Nodes[j].XMLName.Local
		})

		for i := range x.Nodes {
			if !cmp.Equal(x.Nodes[i], y.Nodes[i], nodeComparerOpt()) {
				return false
			}
		}
		return true
	})
}
