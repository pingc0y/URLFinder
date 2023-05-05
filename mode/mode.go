package mode

type Config struct {
	Headers map[string]string `yaml:"headers"`
	Proxy   string            `yaml:"proxy"`

	JsFind   []string            `yaml:"jsFind"`
	UrlFind  []string            `yaml:"urlFind"`
	InfoFind map[string][]string `yaml:"infoFiler"`

	JsFiler  []string `yaml:"jsFiler"`
	UrlFiler []string `yaml:"urlFiler"`

	JsFuzzPath []string `yaml:"jsFuzzPath"`
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
