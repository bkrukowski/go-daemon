package config

type Config struct {
	Vars      map[string]string  `yaml:"vars"`
	Templates map[string]string  `yaml:"templates"`
	Processes map[string]Process `yaml:"processes"`
}

type Process struct {
	Template string
	Vars     map[string]string
	Tags     []string
	Compiled struct {
		Template string
		Vars     map[string]string
	}
}

func (p *Process) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err == nil {
		*p = Process{Template: s}
		return nil
	}

	var v struct {
		Template string            `yaml:"template"`
		Vars     map[string]string `yaml:"vars"`
		Tags     []string          `yaml:"tags"`
	}

	if err := unmarshal(&v); err != nil {
		return err
	}

	*p = Process{
		Template: v.Template,
		Vars:     v.Vars,
		Tags:     v.Tags,
	}

	return nil
}
