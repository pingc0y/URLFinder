package mode

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
	InfoFind   map[string][]string `yaml:"infoFiler"`
	Risks      []string            `yaml:"risks"`
	JsFiler    []string            `yaml:"jsFiler"`
	UrlFiler   []string            `yaml:"urlFiler"`
	JsFuzzPath []string            `yaml:"jsFuzzPath"`
}

type Link struct {
	Url             string
	Status          string
	Size            string
	Title           string
	Redirect        string
	Source          string
	ResponseHeaders map[string]string
	ResponseBody    string
}

type Info struct {
	Phone  []string
	Email  []string
	IDcard []string
	JWT    []string
	Other  []string
	Source string
}
