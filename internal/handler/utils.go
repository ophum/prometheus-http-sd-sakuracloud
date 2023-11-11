package handler

import (
	"slices"
	"strings"

	"github.com/sacloud/iaas-api-go/types"
)

func getTag(tags types.Tags, key string) (string, bool) {
	key = key + "="
	for _, t := range tags {
		if strings.HasPrefix(t, key) {
			return strings.TrimPrefix(t, key), true
		}
	}
	return "", false
}

type tagsObject interface {
	GetTags() types.Tags
}

func isScrape(obj tagsObject) bool {
	scrape, ok := getTag(obj.GetTags(), "prometheus.io/scrape")
	return ok && scrape == "true"
}

func getPort(obj tagsObject) string {
	port, ok := getTag(obj.GetTags(), "prometheus.io/port")
	if !ok {
		port = "9100"
	}
	return port
}

func getExcludes(obj tagsObject, key string) []string {
	key += "="
	excludes := []string{}
	for _, t := range obj.GetTags() {
		if strings.HasPrefix(t, key) {
			excludes = append(excludes, strings.TrimPrefix(t, key))
		}
	}
	return excludes
}

func isExclude(excludes []string, v string) bool {
	return slices.Index(excludes, v) != -1
}
