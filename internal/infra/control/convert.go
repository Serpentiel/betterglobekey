package control

import (
	"fmt"
	"time"

	domainconfig "github.com/Serpentiel/betterglobekey/internal/domain/config"
	controlv1 "github.com/Serpentiel/betterglobekey/internal/gen/betterglobekey/control/v1"
)

// toProto converts the domain configuration model into its protobuf representation.
func toProto(cfg domainconfig.Config) *controlv1.Config {
	collections := make([]*controlv1.Collection, 0, len(cfg.Collections))
	for _, collection := range cfg.Collections {
		collections = append(collections, &controlv1.Collection{
			Name:    collection.Name,
			Sources: collection.Sources,
		})
	}

	return &controlv1.Config{
		Logger: &controlv1.Logger{
			Path:           cfg.Logger.Path,
			RetentionDays:  int32(cfg.Logger.RetentionDays),  //nolint:gosec // retention values are small, non-negative
			RetentionFiles: int32(cfg.Logger.RetentionFiles), //nolint:gosec // retention values are small, non-negative
		},
		DoublePressMaxDelay: cfg.DoublePressMaxDelay.String(),
		Hud:                 cfg.HUD,
		Collections:         collections,
	}
}

// fromProto converts a protobuf configuration into the domain model, parsing the
// double-press delay. Validation of the resulting model is left to the caller.
func fromProto(proto *controlv1.Config) (domainconfig.Config, error) {
	if proto == nil {
		return domainconfig.Config{}, fmt.Errorf("config is required")
	}

	delay, err := time.ParseDuration(proto.GetDoublePressMaxDelay())
	if err != nil {
		return domainconfig.Config{}, fmt.Errorf(
			"invalid double_press_max_delay %q: %w", proto.GetDoublePressMaxDelay(), err,
		)
	}

	collections := make([]domainconfig.Collection, 0, len(proto.GetCollections()))
	for _, collection := range proto.GetCollections() {
		collections = append(collections, domainconfig.Collection{
			Name:    collection.GetName(),
			Sources: collection.GetSources(),
		})
	}

	return domainconfig.Config{
		Logger: domainconfig.Logger{
			Path:           proto.GetLogger().GetPath(),
			RetentionDays:  int(proto.GetLogger().GetRetentionDays()),
			RetentionFiles: int(proto.GetLogger().GetRetentionFiles()),
		},
		DoublePressMaxDelay: delay,
		HUD:                 proto.GetHud(),
		Collections:         collections,
	}, nil
}
