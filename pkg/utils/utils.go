package utils

import (
	"bytes"
	"math/rand"
	"strings"
	"text/template"
	"time"

	"github.com/oklog/ulid"
)

// GetULID gives us a new ulid
//
func GetULID() string {
	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}

// CloudFormationName fixes up names to be compatible with cloudformation
//
func CloudFormationName(name string) string {
	return strings.ToLower(strings.Replace(name, "-", "", -1))
}

// RenderCloudformationTemplate renders a golang template as a string
//
func RenderCloudformationTemplate(templateBody string, obj interface{}) (string, error) {
	tmpl, err := template.
		New("CloudFormationTemplate").
		Funcs(template.FuncMap{
			"CloudFormationName": CloudFormationName,
		}).
		Parse(templateBody)

	if err != nil {
		return "", err
	}
	var buff bytes.Buffer

	err = tmpl.Execute(&buff, obj)
	var templateResult = buff.String()
	return templateResult, nil
}
