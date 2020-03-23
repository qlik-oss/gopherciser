package helpers

import (
	"fmt"
	"strconv"
	"math/rand"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/randomizer"
)

type (
	//DistributionType type of distribution
	DistributionType int

	// DistributionSettings think time settings
	DistributionSettings struct {
		// Type of timer delay
		Type DistributionType `json:"type" displayname:"Distribution type" doc-key:"thinktime.type"`
		// Delay static delay when using static timer delay
		Delay float64 `json:"delay,omitempty" displayname:"Static delay" doc-key:"thinktime.delay"`
		// Mean value
		Mean float64 `json:"mean,omitempty" displayname:"Mean value" doc-key:"thinktime.mean"`
		// Deviation value
		Deviation float64 `json:"dev,omitempty" displayname:"Deviation value" doc-key:"thinktime.dev"`
	}
)

const (
	// StaticDistribution fixed distribution
	StaticDistribution DistributionType = iota
	// UniformDistribution uniform distribution, defined by "mean" and "dev"
	UniformDistribution
)

func (value DistributionType) GetEnumMap() *enummap.EnumMap{
	enumMap, _ := enummap.NewEnumMap(map[string]int{
		"static":  int(StaticDistribution),
		"uniform": int(UniformDistribution),
	})
	return enumMap
}

// UnmarshalJSON unmarshal DistributionType
func (value *DistributionType) UnmarshalJSON(arg []byte) error {
	i, err := value.GetEnumMap().UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal DistributionType")
	}

	*value = DistributionType(i)
	return nil
}

// MarshalJSON marshal ThinkTime type
func (value DistributionType) MarshalJSON() ([]byte, error) {
	str, err := value.GetEnumMap().String(int(value))
	if err != nil {
		return nil, errors.Errorf("Unknown DistributionType<%d>", value)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

func (settings DistributionSettings) Validate() error {
	switch settings.Type {
	case StaticDistribution:
		if settings.Delay <= 0.001 {
			return errors.New("Illegal static distribution value")
		}
	case UniformDistribution:
		base := "uniform distribution requires"
		if settings.Mean <= 0.001 {
			return errors.Errorf("%s a (positive) mean value defined", base)
		}
		if settings.Deviation <= 0.001 {
			return errors.Errorf("%s a (positive) deviation defined", base)
		}
		if settings.Mean <= settings.Deviation {
			return errors.Errorf("%s a mean value<%f> greater than the deviation<%f>", base, settings.Mean, settings.Deviation)
		}
	default:
		typ, err := settings.Type.GetEnumMap().String(int(settings.Type))
		if err != nil {
			return errors.Errorf("distribution type<%d> not supported", settings.Type)
		}
		return errors.Errorf("distribution type<%s> not supported", typ)
	}
	return nil
}

func (settings DistributionSettings) GetSample(rnd *randomizer.Randomizer) (float64, error) {
	switch settings.Type {
	case StaticDistribution:
		return settings.Delay, nil
	case UniformDistribution:
		min := settings.Mean - settings.Deviation
		max := settings.Mean + settings.Deviation
		delay := (rand.Float64() * (max-min)) + min
		return delay, nil
	default:
		return 0, errors.Errorf("distribution type<%d> not yet supported", settings.Type)
	}
}

func (settings DistributionSettings) GetMax(rnd *randomizer.Randomizer) (float64, error) {
	switch settings.Type {
	case StaticDistribution:
		return settings.Delay, nil
	case UniformDistribution:
		max := settings.Mean + settings.Deviation
		return max, nil
	default:
		return 0, errors.Errorf("distribution type<%d> not yet supported", settings.Type)
	}
}

func (settings DistributionSettings) GetMin(rnd *randomizer.Randomizer) (float64, error) {
	switch settings.Type {
	case StaticDistribution:
		return settings.Delay, nil
	case UniformDistribution:
		min := settings.Mean - settings.Deviation
		return min, nil
	default:
		return 0, errors.Errorf("distribution type<%d> not yet supported", settings.Type)
	}
}

func (settings DistributionSettings) GetActionInfo() string {
	switch settings.Type {
	case StaticDistribution:
		return fmt.Sprintf("delay:%s", strconv.FormatFloat(settings.Delay, 'f', -1, 64))
	case UniformDistribution:
		return fmt.Sprintf("mean:%f;deviation:%f", settings.Mean, settings.Deviation)
	default:
		return ""
	}
}
