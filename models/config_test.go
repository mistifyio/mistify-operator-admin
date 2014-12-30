package models_test

import (
	"strings"
	"testing"

	h "github.com/bakins/test-helpers"
	"github.com/mistifyio/mistify-operator-admin/config"
	"github.com/mistifyio/mistify-operator-admin/models"
)

var configFileName = "../cmd/mistify-operator-admin/testconfig.json"
var configJSON = `{
	"foo": {
		"bar":"baz"
	}
}`

func TestConfigValidate(t *testing.T) {
	c := &models.Config{}
	h.Equals(t, models.ErrNilData, c.Validate())

	c = models.NewConfig()
	h.Ok(t, c.Validate())
}

func TestConfigGet(t *testing.T) {
	config.Load(configFileName)
	c := models.NewConfig()
	conf := c.Get()
	h.Assert(t, conf != nil, "undefined conf")
	ns, ok := conf["foobar"]
	h.Assert(t, ok, "default ns not found")
	baz, ok := ns["baz"]
	h.Assert(t, ok, "default key not found")
	h.Equals(t, "default", baz)
}

func TestConfigGetNamespace(t *testing.T) {
	config.Load(configFileName)
	c := models.NewConfig()
	ns := c.GetNamespace("foobar")
	h.Assert(t, ns != nil, "default ns not found")
	baz, ok := ns["baz"]
	h.Equals(t, true, ok)
	h.Equals(t, "default", baz)
}

func TestConfigSetNamespace(t *testing.T) {
	config.Load(configFileName)
	c := models.NewConfig()
	c.SetNamespace("ns", map[string]string{"baz": "bang"})
	ns := c.GetNamespace("ns")
	h.Equals(t, "bang", ns["baz"])
	c.SetNamespace("ns", nil)
	ns = c.GetNamespace("ns")
	h.Equals(t, 0, len(ns))
	c.SetNamespace("foobar", map[string]string{"baz": "bang"})
	ns = c.GetNamespace("foobar")
	h.Equals(t, "bang", ns["baz"])
}

func TestConfigDeleteNamespace(t *testing.T) {
	config.Load(configFileName)
	c := models.NewConfig()
	c.SetNamespace("ns", map[string]string{"baz": "bang"})
	c.DeleteNamespace("ns")
	ns := c.GetNamespace("ns")
	h.Equals(t, 0, len(ns))
}

func TestConfigGetValue(t *testing.T) {
	config.Load(configFileName)
	c := models.NewConfig()
	val, ok := c.GetValue("foobar", "baz")
	h.Equals(t, true, ok)
	h.Equals(t, "default", val)
	val, ok = c.GetValue("foobar", "baz2")
	h.Equals(t, false, ok)
}

func TestConfigSetValue(t *testing.T) {
	config.Load(configFileName)
	c := models.NewConfig()
	c.SetValue("foobar", "baz", "bang")
	val, ok := c.GetValue("foobar", "baz")
	h.Equals(t, true, ok)
	h.Equals(t, "bang", val)
}

func TestConfigDeleteValue(t *testing.T) {
	config.Load(configFileName)
	c := models.NewConfig()
	c.SetValue("foobar", "baz2", "bang")
	c.DeleteValue("foobar", "baz2")
	_, ok := c.GetValue("foobar", "baz2")
	h.Equals(t, false, ok)
}

func TestConfigMerge(t *testing.T) {
	config.Load(configFileName)
	c := models.NewConfig()
	c.SetValue("foobar", "baz2", "bang")
	c2 := models.NewConfig()
	c2.SetValue("foobar", "baz3", "bong")
	c2.SetValue("foobar2", "baz", "bang")
	c.Merge(c2)
	val, _ := c.GetValue("foobar", "baz2")
	h.Equals(t, "bang", val)
	val, _ = c.GetValue("foobar", "baz3")
	h.Equals(t, "bong", val)
	val, _ = c.GetValue("foobar2", "baz")
	h.Equals(t, "bang", val)
}

func TestConfigSave(t *testing.T) {
	config.Load(configFileName)
	c := models.NewConfig()
	c.SetValue("foobar", "baz", "bang")
	h.Ok(t, c.Save())
}

func TestConfigLoad(t *testing.T) {
	config.Load(configFileName)
	c := models.NewConfig()
	c.Load()
	val, ok := c.GetValue("foobar", "baz")
	h.Equals(t, true, ok)
	h.Equals(t, "bang", val)
}

func TestConfigDecode(t *testing.T) {
	config.Load(configFileName)
	r := strings.NewReader(configJSON)
	c := models.NewConfig()
	h.Ok(t, c.Decode(r))
	val, _ := c.GetValue("foo", "bar")
	h.Equals(t, "baz", val)
}
