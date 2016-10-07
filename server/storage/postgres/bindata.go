// Code generated by go-bindata.
// sources:
// migrations/0001_initial_structure.sql
// migrations/0002_drop_parent_id_not_null.sql
// migrations/0003_move_owner_id_to_taggables.sql
// migrations/0004_add_on_delete.sql
// DO NOT EDIT!

package postgres

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)
type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _migrations0001_initial_structureSql = []byte(`-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE users (
  id SERIAL NOT NULL PRIMARY KEY,
  username VARCHAR(50),
  password VARCHAR(100),
  email VARCHAR(50)
);

CREATE TABLE tags (
  id SERIAL NOT NULL PRIMARY KEY,
  parent_id INTEGER NOT NULL,
  owner_id INTEGER NOT NULL,
  FOREIGN KEY (parent_id) REFERENCES tags(id),
  FOREIGN KEY (owner_id) REFERENCES users(id)
);

CREATE TABLE tag_names (
  id SERIAL NOT NULL PRIMARY KEY,
  tag_id INTEGER NOT NULL,
  name VARCHAR(30) NOT NULL,
  FOREIGN KEY (tag_id) REFERENCES tags(id)
);

CREATE TABLE taggables (
  id SERIAL NOT NULL PRIMARY KEY,
  "type" VARCHAR(30) NOT NULL
);

CREATE TABLE bookmarks (
  id SERIAL NOT NULL PRIMARY KEY,
  url TEXT NOT NULL,
  comment TEXT NOT NULL,
  owner_id INTEGER NOT NULL,
  FOREIGN KEY (id) REFERENCES taggables(id),
  FOREIGN KEY (owner_id) REFERENCES users(id)
);

CREATE TABLE taggings (
  id SERIAL NOT NULL PRIMARY KEY,
  taggable_id INTEGER NOT NULL,
  tag_id INTEGER NOT NULL,
  FOREIGN KEY (taggable_id) REFERENCES taggables(id),
  FOREIGN KEY (tag_id) REFERENCES tags(id)
);

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE users;
DROP TABLE tags;
DROP TABLE tag_names;
DROP TABLE bookmarks;
DROP TABLE taggables;
DROP TABLE taggings;
`)

func migrations0001_initial_structureSqlBytes() ([]byte, error) {
	return _migrations0001_initial_structureSql, nil
}

func migrations0001_initial_structureSql() (*asset, error) {
	bytes, err := migrations0001_initial_structureSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "migrations/0001_initial_structure.sql", size: 1329, mode: os.FileMode(420), modTime: time.Unix(1, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _migrations0002_drop_parent_id_not_nullSql = []byte(`-- NULL parent is going to be used for root tag for each user

-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE "tags" ALTER COLUMN "parent_id" DROP NOT NULL;

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE "tags" ALTER COLUMN "parent_id" SET NOT NULL;
`)

func migrations0002_drop_parent_id_not_nullSqlBytes() ([]byte, error) {
	return _migrations0002_drop_parent_id_not_nullSql, nil
}

func migrations0002_drop_parent_id_not_nullSql() (*asset, error) {
	bytes, err := migrations0002_drop_parent_id_not_nullSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "migrations/0002_drop_parent_id_not_null.sql", size: 348, mode: os.FileMode(420), modTime: time.Unix(1, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _migrations0003_move_owner_id_to_taggablesSql = []byte(`-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE "bookmarks" DROP COLUMN "owner_id" RESTRICT;

ALTER TABLE "taggables" ADD COLUMN "owner_id" INTEGER NOT NULL;

ALTER TABLE "taggables" ADD CONSTRAINT "taggables_owner_id_fkey"
  FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE;
-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE "taggables" DROP COLUMN "owner_id" RESTRICT;

ALTER TABLE "bookmarks" ADD COLUMN "owner_id" INTEGER NOT NULL;

ALTER TABLE "bookmarks" ADD CONSTRAINT "bookmarks_owner_id_fkey"
  FOREIGN KEY (owner_id) REFERENCES users(id);
`)

func migrations0003_move_owner_id_to_taggablesSqlBytes() ([]byte, error) {
	return _migrations0003_move_owner_id_to_taggablesSql, nil
}

func migrations0003_move_owner_id_to_taggablesSql() (*asset, error) {
	bytes, err := migrations0003_move_owner_id_to_taggablesSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "migrations/0003_move_owner_id_to_taggables.sql", size: 655, mode: os.FileMode(420), modTime: time.Unix(1, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _migrations0004_add_on_deleteSql = []byte(`-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE "tags" DROP CONSTRAINT "tags_parent_id_fkey";

ALTER TABLE "tags" ADD CONSTRAINT "tags_parent_id_fkey"
  FOREIGN KEY (parent_id) REFERENCES tags(id) ON DELETE CASCADE;

ALTER TABLE "tags" DROP CONSTRAINT "tags_owner_id_fkey";

ALTER TABLE "tags" ADD CONSTRAINT "tags_owner_id_fkey"
  FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE "tag_names" DROP CONSTRAINT "tag_names_tag_id_fkey";

ALTER TABLE "tag_names" ADD CONSTRAINT "tag_names_tag_id_fkey"
  FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE;

ALTER TABLE "bookmarks" DROP CONSTRAINT "bookmarks_id_fkey";

ALTER TABLE "bookmarks" ADD CONSTRAINT "bookmarks_id_fkey"
  FOREIGN KEY (id) REFERENCES taggables(id) ON DELETE CASCADE;

ALTER TABLE "taggings" DROP CONSTRAINT "taggings_taggable_id_fkey";

ALTER TABLE "taggings" ADD CONSTRAINT "taggings_taggable_id_fkey"
  FOREIGN KEY (taggable_id) REFERENCES taggables(id) ON DELETE CASCADE;

ALTER TABLE "taggings" DROP CONSTRAINT "taggings_tag_id_fkey";

ALTER TABLE "taggings" ADD CONSTRAINT "taggings_tag_id_fkey"
  FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE;

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
`)

func migrations0004_add_on_deleteSqlBytes() ([]byte, error) {
	return _migrations0004_add_on_deleteSql, nil
}

func migrations0004_add_on_deleteSql() (*asset, error) {
	bytes, err := migrations0004_add_on_deleteSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "migrations/0004_add_on_delete.sql", size: 1300, mode: os.FileMode(420), modTime: time.Unix(1, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
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
var _bindata = map[string]func() (*asset, error){
	"migrations/0001_initial_structure.sql": migrations0001_initial_structureSql,
	"migrations/0002_drop_parent_id_not_null.sql": migrations0002_drop_parent_id_not_nullSql,
	"migrations/0003_move_owner_id_to_taggables.sql": migrations0003_move_owner_id_to_taggablesSql,
	"migrations/0004_add_on_delete.sql": migrations0004_add_on_deleteSql,
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
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = &bintree{nil, map[string]*bintree{
	"migrations": &bintree{nil, map[string]*bintree{
		"0001_initial_structure.sql": &bintree{migrations0001_initial_structureSql, map[string]*bintree{}},
		"0002_drop_parent_id_not_null.sql": &bintree{migrations0002_drop_parent_id_not_nullSql, map[string]*bintree{}},
		"0003_move_owner_id_to_taggables.sql": &bintree{migrations0003_move_owner_id_to_taggablesSql, map[string]*bintree{}},
		"0004_add_on_delete.sql": &bintree{migrations0004_add_on_deleteSql, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

