package db

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
)

func bindata_read(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	return buf.Bytes(), nil
}

var _schema_000_init_sql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x0e\x72\x75\x0c\x71\x55\x08\x71\x74\xf2\x71\x55\x28\x2d\x4e\x2d\x2a\xd6\xe0\x52\x80\x82\xcc\x14\x05\x4f\xbf\x10\x57\x77\xd7\x20\x85\x80\x20\x4f\x5f\xc7\xa0\x48\x05\x6f\xd7\x48\x05\x7d\x2d\x45\x13\x03\x43\x03\x43\x05\xc7\xd0\x10\xff\x78\x4f\x3f\xe7\x20\x57\x5f\x57\xbf\x10\x05\x2d\x7d\x1d\xb8\xd6\xbc\xc4\xdc\x54\x85\x30\xc7\x20\x67\x0f\xc7\x20\x0d\x43\x03\x23\x13\x4d\x84\x5c\x6a\x6e\x62\x66\x0e\x2e\xc9\xe2\x82\xc4\x64\x34\x9d\x5c\x9a\xd6\x5c\x5c\x80\x00\x00\x00\xff\xff\x1c\x0f\x28\xf1\xa8\x00\x00\x00")

func schema_000_init_sql() ([]byte, error) {
	return bindata_read(
		_schema_000_init_sql,
		"schema/000_init.sql",
	)
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		return f()
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() ([]byte, error){
	"schema/000_init.sql": schema_000_init_sql,
}
// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for name := range node.Children {
		rv = append(rv, name)
	}
	return rv, nil
}

type _bintree_t struct {
	Func func() ([]byte, error)
	Children map[string]*_bintree_t
}
var _bintree = &_bintree_t{nil, map[string]*_bintree_t{
	"schema": &_bintree_t{nil, map[string]*_bintree_t{
		"000_init.sql": &_bintree_t{schema_000_init_sql, map[string]*_bintree_t{
		}},
	}},
}}
