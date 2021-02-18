package ipldlegacy_test

import (
	"fmt"
	"testing"

	format "github.com/ipfs/go-ipld-format"
	ipldlegacy "github.com/ipfs/go-ipld-legacy"
	"github.com/ipfs/go-ipld-legacy/testutil"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	tree := testutil.NewTestIPLDTree()
	legacyRootNode := &ipldlegacy.LegacyNode{
		tree.RootBlock,
		tree.RootNode,
	}
	legacyListNode := &ipldlegacy.LegacyNode{
		tree.MiddleListBlock,
		tree.MiddleListNode,
	}
	legacyMapNode := &ipldlegacy.LegacyNode{
		tree.MiddleMapBlock,
		tree.MiddleMapNode,
	}
	testCases := map[string]struct {
		node              *ipldlegacy.LegacyNode
		path              []string
		expectedValue     interface{}
		expectedRemaining []string
		expectedErr       error
	}{
		"resolves plain string": {
			node:          legacyRootNode,
			path:          []string{"plain"},
			expectedValue: "olde string",
		},
		"errors on path beyond terminal type": {
			node:        legacyRootNode,
			path:        []string{"plain", "old"},
			expectedErr: fmt.Errorf("error traversing at \"plain\": tried to resolve through object that had no links"),
		},
		"stops at link with remaining": {
			node:              legacyRootNode,
			path:              []string{"linkedMap", "foo"},
			expectedValue:     &format.Link{Cid: tree.MiddleMapNodeLnk.(cidlink.Link).Cid},
			expectedRemaining: []string{"foo"},
		},
		"resolves array": {
			node:          legacyListNode,
			path:          []string{"2"},
			expectedValue: &format.Link{Cid: tree.LeafBetaLnk.(cidlink.Link).Cid},
		},
		"resolves other types": {
			node:          legacyMapNode,
			path:          []string{"foo"},
			expectedValue: true,
		},
		"resolves complex types": {
			node: legacyMapNode,
			path: []string{"nested"},
			expectedValue: map[string]interface{}{
				"alink": map[string]interface{}{
					"/": tree.LeafAlphaLnk.String(),
				},
				"nonlink": "zoo",
			},
		},
	}
	for testCase, data := range testCases {
		t.Run(testCase, func(t *testing.T) {
			value, remaining, err := data.node.Resolve(data.path)
			if data.expectedErr == nil {
				require.NoError(t, err)
				require.Equal(t, data.expectedValue, value)
				require.Equal(t, data.expectedRemaining, remaining)
			} else {
				require.EqualError(t, err, data.expectedErr.Error())
			}
		})
	}
}
func TestTree(t *testing.T) {
	tree := testutil.NewTestIPLDTree()
	legacyMapNode := &ipldlegacy.LegacyNode{
		tree.MiddleMapBlock,
		tree.MiddleMapNode,
	}
	legacyListNode := &ipldlegacy.LegacyNode{
		tree.MiddleListBlock,
		tree.MiddleListNode,
	}
	testCases := map[string]struct {
		node     *ipldlegacy.LegacyNode
		path     string
		depth    int
		expected []string
	}{
		"resolves array": {
			node:     legacyListNode,
			path:     "",
			depth:    -1,
			expected: []string{"0", "1", "2", "3"},
		},
		"resolve to depth": {
			node:     legacyMapNode,
			path:     "",
			depth:    1,
			expected: []string{"foo", "bar", "nested"},
		},
		"resolve nested": {
			node:     legacyMapNode,
			path:     "",
			depth:    -1,
			expected: []string{"foo", "bar", "nested", "nested/alink", "nested/nonlink"},
		},
		"resolves with starting path": {
			node:     legacyMapNode,
			path:     "nested",
			depth:    -1,
			expected: []string{"alink", "nonlink"},
		},
	}
	for testCase, data := range testCases {
		t.Run(testCase, func(t *testing.T) {
			paths := data.node.Tree(data.path, data.depth)
			require.Equal(t, data.expected, paths)
		})
	}
}
