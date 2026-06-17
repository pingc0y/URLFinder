package mode

import "gopkg.in/yaml.v3"

type Config struct {
	Proxy      string              `yaml:"proxy"`
	Timeout    int                 `yaml:"timeout"`
	Thread     int                 `yaml:"thread"`
	UrlSteps   int                 `yaml:"urlSteps"`
	JsSteps    int                 `yaml:"jsSteps"`
	Max        int                 `yaml:"max"`
	Headers    map[string]string   `yaml:"headers"`
	JsFind     []string            `yaml:"jsFind"`
	UrlFind    []string            `yaml:"urlFind"`
	InfoFind   map[string][]string `yaml:"infoFind"`
	Risks      []string            `yaml:"risks"`
	JsFiler    []string            `yaml:"jsFiler"`
	UrlFiler   []string            `yaml:"urlFiler"`
	JsFuzzPath []string            `yaml:"jsFuzzPath"`
}

func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	type configAlias Config
	aux := struct {
		configAlias    `yaml:",inline"`
		LegacyInfoFind map[string][]string `yaml:"infoFiler"`
	}{}

	if err := value.Decode(&aux); err != nil {
		return err
	}

	*c = Config(aux.configAlias)
	if c.InfoFind == nil && aux.LegacyInfoFind != nil {
		c.InfoFind = aux.LegacyInfoFind
	}
	return nil
}

type Link struct {
	Url      string
	Status   string
	Size     string
	Title    string
	Redirect string
	Source   string
}

type Info struct {
	Phone  []string
	Email  []string
	IDcard []string
	JWT    []string
	Other  []string
	Source string
}
