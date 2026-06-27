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
			Level:          cfg.Logger.Level,
			RetentionDays:  int32(cfg.Logger.RetentionDays),  //nolint:gosec // retention values are small, non-negative
			RetentionFiles: int32(cfg.Logger.RetentionFiles), //nolint:gosec // retention values are small, non-negative
		},
		DoublePress: &controlv1.DoublePress{
			Enabled:      cfg.DoublePress.Enabled,
			MaximumDelay: cfg.DoublePress.MaxDelay.String(),
		},
		Reverse: &controlv1.Reverse{
			Enabled:  cfg.Reverse.Enabled,
			Modifier: cfg.Reverse.Modifier,
		},
		Hud: &controlv1.Hud{
			Enabled:        cfg.HUD.Enabled,
			Duration:       cfg.HUD.Duration.String(),
			ShowCollection: cfg.HUD.ShowCollection,
		},
		Collections: collections,
	}
}

// fromProto converts a protobuf configuration into the domain model, parsing the
// duration fields. Validation of the resulting model is left to the caller.
func fromProto(proto *controlv1.Config) (domainconfig.Config, error) {
	if proto == nil {
		return domainconfig.Config{}, fmt.Errorf("config is required")
	}

	delay, err := time.ParseDuration(proto.GetDoublePress().GetMaximumDelay())
	if err != nil {
		return domainconfig.Config{}, fmt.Errorf(
			"invalid double_press.maximum_delay %q: %w", proto.GetDoublePress().GetMaximumDelay(), err,
		)
	}

	duration, err := time.ParseDuration(proto.GetHud().GetDuration())
	if err != nil {
		return domainconfig.Config{}, fmt.Errorf("invalid hud.duration %q: %w", proto.GetHud().GetDuration(), err)
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
			Level:          proto.GetLogger().GetLevel(),
			RetentionDays:  int(proto.GetLogger().GetRetentionDays()),
			RetentionFiles: int(proto.GetLogger().GetRetentionFiles()),
		},
		DoublePress: domainconfig.DoublePress{Enabled: proto.GetDoublePress().GetEnabled(), MaxDelay: delay},
		Reverse: domainconfig.Reverse{
			Enabled:  proto.GetReverse().GetEnabled(),
			Modifier: proto.GetReverse().GetModifier(),
		},
		HUD: domainconfig.HUD{
			Enabled:        proto.GetHud().GetEnabled(),
			Duration:       duration,
			ShowCollection: proto.GetHud().GetShowCollection(),
		},
		Collections: collections,
	}, nil
}
