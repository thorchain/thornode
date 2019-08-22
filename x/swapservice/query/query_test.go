package query

import (
	"testing"

	. "gopkg.in/check.v1"
)

func TestPackage(t *testing.T) { TestingT(t) }

type QuerySuite struct{}

var _ = Suite(&QuerySuite{})

func (s QuerySuite) TestQuery(c *C) {
	c.Check(QueryTxIn.Endpoint("foo", "bar"), Equals, "/foo/tx/{bar}")
	c.Check(QueryTxIn.Path("foo", "bar"), Equals, "custom/foo/txin/bar")

	c.Check(QueryAdminConfig.Endpoint("foo", "bar", "baz"), Equals, "/foo/admin/{bar}/{baz}")
	c.Check(QueryAdminConfig.Path("foo", "bar", "baz"), Equals, "custom/foo/adminconfig/bar/baz")
}
