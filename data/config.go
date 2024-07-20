package data

import (
	"os"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Volume struct {
	Name       string
	Type       string
	FileSystem string
	Location   string
}

type Host struct {
	Name    string
	Volumes []Volume
}

type Configuration struct {
	PasswdFile string
	Mounts     string
	Hosts      []Host
	Dryrun     bool
	Version    bool
}

type fileConfiguration struct {
	PasswdFile string
	Mounts     string
	Hosts      []Host
}

func NewConfiguration() (*Configuration, error) {
	v := viper.New()

	v.SetConfigName("remotebackup.toml")
	v.SetConfigType("toml")
	v.AddConfigPath(".") // TODO: set up search path infra

	v.SetDefault("mounts", "/srv/remotebackup")

	cli := pflag.NewFlagSet("remote_backup", pflag.ExitOnError)
	cli.BoolP("version", "v", false, "Display the version string.")
	cli.BoolP("dry-run", "n", true, "Only show what would be done.") // TODO: switch to "false" default for prod.
	cli.Parse(os.Args[1:])
	v.BindPFlags(cli)

	v.SetEnvPrefix("REMOTEBACKUP")
	v.BindEnv("mounts")

	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	fc := fileConfiguration{}
	err = v.Unmarshal(&fc)
	if err != nil {
		return nil, err
	}

	config := &Configuration{}
	config.Hosts = fc.Hosts
	config.PasswdFile = fc.PasswdFile

	config.Mounts = v.Get("mounts").(string)
	config.Dryrun = v.Get("dry-run").(bool)
	config.Version = v.Get("version").(bool)

	return config, nil
}
