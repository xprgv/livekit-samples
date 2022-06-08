package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

func GetConfig(path string) (Config, error) {
	viper.SetConfigFile(path)

	cfg := Config{}

	if err := viper.ReadInConfig(); err != nil {
		return cfg, err
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	if err := validate.Struct(cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

var validate = validator.New()

type Config struct {
	Inputs  []Input  `mapstructure:"inputs"`
	Outputs []Output `mapstructure:"outputs"`
	Livekit Livekit  `mapstructure:"livekit" validate:"required"`
}

// func (c Config) Validate() error {
// 	if err := validate.Struct(c); err != nil {
// 		return err
// 	}
// 	return nil
// }

type Input struct {
	Codec      string `mapstructure:"codec" validate:"required"`
	Resolution int    `mapstructure:"resolution" validate:"required"`
}

type Output struct {
	Codec      string `mapstructure:"codec" validate:"required"`
	Resolution int    `mapstructure:"resolution" validate:"required"`
}

type Livekit struct {
	Host                string `mapstructure:"host" validate:"required"`
	RoomName            string `mapstructure:"room_name" validate:"required"`
	ParticipantName     string `mapstructure:"participant_name" validate:"required"`
	ParticipantIdentity string `mapstructure:"participant_identity" validate:"required"`
	ApiKey              string `mapstructure:"api_key" validate:"required"`
	ApiSecret           string `mapstructure:"api_secret" validate:"required"`
	Token               string `mapstructure:"token" validate:"required"`
}
