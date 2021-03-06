// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/juju/names.v2"
	"gopkg.in/yaml.v2"
)

type RemoteApplicationSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&RemoteApplicationSerializationSuite{})

func (s *RemoteApplicationSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "remote applications"
	s.sliceName = "remote-applications"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importRemoteApplications(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["remote-applications"] = []interface{}{}
	}
}

func minimalRemoteApplicationMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"name":              "civil-wars",
		"offer-name":        "barton-hollow",
		"url":               "http://a.url",
		"source-model-uuid": "abcd-1234",
		"registered":        true,
		"endpoints": map[interface{}]interface{}{
			"version": 1,
			"endpoints": []interface{}{map[interface{}]interface{}{
				"name":      "lana",
				"role":      "provider",
				"interface": "mysql",
				"limit":     1,
				"scope":     "global",
			}},
		},
	}
}

func minimalRemoteApplication() *remoteApplication {
	a := newRemoteApplication(RemoteApplicationArgs{
		Tag:         names.NewApplicationTag("civil-wars"),
		OfferName:   "barton-hollow",
		URL:         "http://a.url",
		SourceModel: names.NewModelTag("abcd-1234"),
		Registered:  true,
	})
	a.AddEndpoint(RemoteEndpointArgs{
		Name:      "lana",
		Role:      "provider",
		Interface: "mysql",
		Limit:     1,
		Scope:     "global",
	})
	return a
}

func (*RemoteApplicationSerializationSuite) TestNew(c *gc.C) {
	r := minimalRemoteApplication()
	c.Check(r.Tag(), gc.Equals, names.NewApplicationTag("civil-wars"))
	c.Check(r.Name(), gc.Equals, "civil-wars")
	c.Check(r.OfferName(), gc.Equals, "barton-hollow")
	c.Check(r.URL(), gc.Equals, "http://a.url")
	c.Check(r.SourceModelTag(), gc.Equals, names.NewModelTag("abcd-1234"))
	c.Check(r.Registered(), jc.IsTrue)
	ep := r.Endpoints()
	c.Assert(ep, gc.HasLen, 1)
	c.Check(ep[0].Name(), gc.Equals, "lana")
}

func (*RemoteApplicationSerializationSuite) TestBadSchema1(c *gc.C) {
	container := map[string]interface{}{
		"version":             1,
		"remote-applications": []interface{}{1234},
	}
	_, err := importRemoteApplications(container)
	c.Assert(err, gc.ErrorMatches, `remote applications version schema check failed: remote-applications\[0\]: expected map, got int\(1234\)`)
}

func (*RemoteApplicationSerializationSuite) TestBadSchema2(c *gc.C) {
	m := minimalRemoteApplicationMap()
	m["registered"] = "blah"
	container := map[string]interface{}{
		"version":             1,
		"remote-applications": []interface{}{m},
	}
	_, err := importRemoteApplications(container)
	c.Assert(err, gc.ErrorMatches, `remote application 0: remote application v1 schema check failed: registered: expected bool, got string\("blah"\)`)
}

func (s *RemoteApplicationSerializationSuite) TestBadEndpoints(c *gc.C) {
	m := minimalRemoteApplicationMap()
	m["endpoints"] = map[interface{}]interface{}{
		"version": 1,
		"bishop":  "otter-trouserpress",
	}
	container := map[string]interface{}{
		"version":             1,
		"remote-applications": []interface{}{m},
	}
	_, err := importRemoteApplications(container)
	c.Assert(err, gc.ErrorMatches, `remote application 0: remote endpoints version schema check failed: endpoints: expected list, got nothing`)
}

func (*RemoteApplicationSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalRemoteApplication())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalRemoteApplicationMap())
}

func (s *RemoteApplicationSerializationSuite) TestRoundTrip(c *gc.C) {
	rIn := minimalRemoteApplication()
	rOut := s.exportImport(c, rIn)
	c.Assert(rOut, jc.DeepEquals, rIn)
}

func (s *RemoteApplicationSerializationSuite) exportImport(c *gc.C, applicationIn *remoteApplication) *remoteApplication {
	applicationsIn := &remoteApplications{
		Version:            1,
		RemoteApplications: []*remoteApplication{applicationIn},
	}
	bytes, err := yaml.Marshal(applicationsIn)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	applicationsOut, err := importRemoteApplications(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(applicationsOut, gc.HasLen, 1)
	return applicationsOut[0]
}
