package fuzz

import (
	"github.com/pingc0y/URLFinder/config"
	"github.com/pingc0y/URLFinder/mode"
	"github.com/pingc0y/URLFinder/result"
	"github.com/pingc0y/URLFinder/util"
	"regexp"
)

func JsFuzz() {

	paths := []string{}
	for i := range result.ResultJs {
		re := regexp.MustCompile("(.+/)[^/]+.js").FindAllStringSubmatch(result.ResultJs[i].Url, -1)
		if len(re) != 0 {
			paths = append(paths, re[0][1])
		}
		re2 := regexp.MustCompile("(https{0,1}://([a-z0-9\\-]+\\.)*([a-z0-9\\-]+\\.[a-z0-9\\-]+)(:[0-9]+)?/)").FindAllStringSubmatch(result.ResultJs[i].Url, -1)
		if len(re2) != 0 {
			paths = append(paths, re2[0][1])
		}
	}
	paths = util.UniqueArr(paths)
	for i := range paths {
		for i2 := range config.JsFuzzPath {
			result.ResultJs = append(result.ResultJs, mode.Link{
				Url:    paths[i] + config.JsFuzzPath[i2],
				Source: "Fuzz",
			})
		}
	}
}
