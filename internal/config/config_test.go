package config

import (
	"encoding/json"
	"testing"

	"github.com/traffic-refinery/traffic-refinery/internal/utils"
)

func TestSetDefaults(t *testing.T) {
	conf := TrafficRefineryConfig{}
	conf.setDefaults()
}

func TestDefaultConfig(t *testing.T) {
	conf := TrafficRefineryConfig{}
	t.Logf("Loading config %s", utils.GetRepoPath()+"/test/config/trconfig_default.json")
	conf.ImportConfigFromFile(utils.GetRepoPath() + "/test/config/trconfig_default.json")
	out, _ := json.Marshal(conf)
	t.Logf("Loaded config: %s", out)
}

func TestVideoConfig(t *testing.T) {
	conf := TrafficRefineryConfig{}
	conf.ImportConfigFromFile(utils.GetRepoPath() + "/test/config/configs/trconfig_video.json")
	out, _ := json.Marshal(conf)
	t.Logf("Loaded config: %s", out)
}

func TestAdsConfig(t *testing.T) {
	conf := TrafficRefineryConfig{}
	conf.ImportConfigFromFile(utils.GetRepoPath() + "/test/config/configs/trconfig_ads.json")
	out, _ := json.Marshal(conf)
	t.Logf("Loaded config: %s", out)
}

func TestReplayConfig(t *testing.T) {
	conf := TrafficRefineryConfig{}
	conf.ImportConfigFromFile(utils.GetRepoPath() + "/test/config/configs/trconfig_replay.json")
	out, _ := json.Marshal(conf)
	t.Logf("Loaded config: %s", out)
}

func TestVoipConfig(t *testing.T) {
	conf := TrafficRefineryConfig{}
	conf.ImportConfigFromFile(utils.GetRepoPath() + "/test/config/configs/trconfig_voip.json")
	out, _ := json.Marshal(conf)
	t.Logf("Loaded config: %s", out)
}
