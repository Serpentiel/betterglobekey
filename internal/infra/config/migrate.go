package config

import (
	"fmt"
	"maps"
	"slices"

	"gopkg.in/yaml.v3"
)

// migration transforms a configuration document from one schema version to the next.
type migration struct {
	from    int
	to      int
	migrate func([]byte) ([]byte, error)
}

// migrations is the ordered registry of schema migrations.
var migrations = []migration{
	{from: versionV1, to: versionV2, migrate: migrateV1ToV2},
}

// detectVersion reports the schema version of a configuration document. A missing
// version field denotes the legacy, unversioned v1 format.
func detectVersion(data []byte) (int, error) {
	var header struct {
		Version int `yaml:"version"`
	}

	if err := yaml.Unmarshal(data, &header); err != nil {
		return 0, fmt.Errorf("reading config version: %w", err)
	}

	if header.Version == 0 {
		return 1, nil
	}

	return header.Version, nil
}

// migrate upgrades a configuration document to the current schema version. It
// returns the upgraded document and the version it started from.
func migrate(data []byte) ([]byte, int, error) {
	from, err := detectVersion(data)
	if err != nil {
		return nil, 0, err
	}

	if from > currentVersion {
		return nil, from, fmt.Errorf("config version %d is newer than supported version %d", from, currentVersion)
	}

	out := data

	for version := from; version < currentVersion; {
		index := slices.IndexFunc(migrations, func(m migration) bool { return m.from == version })
		if index < 0 {
			return nil, from, fmt.Errorf("no migration available from config version %d", version)
		}

		out, err = migrations[index].migrate(out)
		if err != nil {
			return nil, from, fmt.Errorf("migrating config from v%d to v%d: %w", version, migrations[index].to, err)
		}

		version = migrations[index].to
	}

	return out, from, nil
}

// migrateV1ToV2 converts the legacy unordered-map format to the ordered v2 schema.
// Collections are emitted in alphabetical order, since the v1 map has no inherent order.
func migrateV1ToV2(data []byte) ([]byte, error) {
	var v1 schemaV1
	if err := yaml.Unmarshal(data, &v1); err != nil {
		return nil, fmt.Errorf("parsing v1 config: %w", err)
	}

	var v2 schemaV2

	v2.Version = versionV2
	v2.Logger.Path = v1.Logger.Path
	v2.Logger.Retention.Days = v1.Logger.Retain.Days
	v2.Logger.Retention.Files = v1.Logger.Retain.Copies
	v2.DoublePress.MaximumDelay = fmt.Sprintf("%dms", v1.DoublePress.MaximumDelay)

	for _, name := range slices.Sorted(maps.Keys(v1.InputSources)) {
		v2.Collections = append(v2.Collections, collectionV2{Name: name, Sources: v1.InputSources[name]})
	}

	out, err := yaml.Marshal(v2)
	if err != nil {
		return nil, fmt.Errorf("encoding v2 config: %w", err)
	}

	return out, nil
}
