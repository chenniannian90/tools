package appinfo

import (
	"bytes"
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"path"
	"path/filepath"
	"strings"
)

func (c *AppCtx) dockerize() {
	_ = writeToFile(path.Join(c.root, "./config/default.yml"), c.defaultConfig(), yaml.Marshal)
	_ = writeToFile(path.Join(c.root, "./deploy/qservice.yml"), c.qservice(), yaml.Marshal)
	_ = writeToFile(path.Join(c.root, "./Dockerfile.default"), c.dockerfile(), func(v interface{}) ([]byte, error) {
		return v.([]byte), nil
	})
}

func (c *AppCtx) defaultConfig() map[string]string {
	m := map[string]string{}
	m["GOENV"] = "DEV"

	for _, envVars := range c.defaultEnvVarList {
		for _, envVar := range envVars.Values {
			if !envVar.Optional {
				m[envVar.Key(envVars.Prefix)] = envVar.Value
			}
		}
	}

	return m
}

func (c *AppCtx) qservice() map[string]interface{} {
	values := map[string]interface{}{}

	values["apiVersion"] = "serving.octohelm.tech/v1alpha1"
	values["kind"] = "QService"

	spec := map[string]interface{}{}

	spec["image"] = "${{ PROJECT_IMAGE }}"

	annotations := map[string]interface{}{}

	ports := make([]string, 0)
	envs := map[string]string{}

	for _, envVars := range c.defaultEnvVarList {
		for _, envVar := range envVars.Values {
			v := envVar.Value
			key := envVar.Key(envVars.Prefix)

			if !envVar.Optional {
				envs[key] = fmt.Sprintf("${{ %s }}", key)
			}

			if v != "" {

				if envVar.IsCopy && v == "./openapi.json" {
					annotations["octohelm.tech/openAPISpecPath"] = "/" + strings.TrimPrefix(c.name, "srv-")
				}

				if envVar.IsExpose {
					ports = append(ports, v)
				}

				if envVar.IsHealthCheck {
					spec["livenessProbe"] = map[string]interface{}{
						"action":              "http://:80/",
						"initialDelaySeconds": 5,
						"periodSeconds":       5,
					}

					spec["readinessProbe"] = spec["livenessProbe"]
				}
			}
		}
	}

	values["metadata"] = map[string]interface{}{
		"annotations": annotations,
	}
	values["spec"] = spec

	spec["envs"] = envs

	if len(ports) > 0 {
		spec["ports"] = ports
	}

	return values
}

func (c *AppCtx) dockerfile() []byte {
	dockerfile := bytes.NewBuffer(nil)

	_, _ = fmt.Fprintln(dockerfile, `ARG DOCKER_REGISTRY=hub-dev.xxx.com
FROM ${DOCKER_REGISTRY}/docker.io/library/golang:1.17-buster AS build-env

# setup private pkg 
ARG CI_JOB_TOKEN
ARG GITLAB_HOST=git.xxx.com
ARG GOPROXY=https://goproxy.cn,direct
ENV GONOSUMDB=${GITLAB_HOST}/*
RUN git config --global url.https://gitlab-ci-token:${CI_JOB_TOKEN}@${GITLAB_HOST}/.insteadOf https://${GITLAB_HOST}/

FROM build-env AS builder

WORKDIR /go/src
COPY ./ ./

# build
ARG COMMIT_SHA
RUN make build

# runtime
FROM ${DOCKER_REGISTRY}/ghcr.io/xxx/distroless/static-debian10:latest

COPY --from=builder `+filepath.Join("/go/src/cmd", c.name, c.name)+` `+filepath.Join(`/go/bin`, c.name)+`
`)

	for _, envVars := range c.defaultEnvVarList {
		for _, envVar := range envVars.Values {
			if envVar.Value != "" {
				if envVar.IsCopy {
					_, _ = fmt.Fprintf(dockerfile, "COPY --from=builder %s %s\n", filepath.Join("/go/src/cmd", c.name, envVar.Value), filepath.Join("/go/bin", envVar.Value))
				}
				if envVar.IsExpose {
					_, _ = fmt.Fprintf(dockerfile, "EXPOSE %s\n", envVar.Value)
				}
			}
		}
	}

	fmt.Fprintf(dockerfile, `
ARG PROJECT_NAME
ARG PROJECT_VERSION
ENV GOENV=DEV PROJECT_NAME=${PROJECT_NAME} PROJECT_VERSION=${PROJECT_VERSION}

WORKDIR /go/bin
ENTRYPOINT ["`+filepath.Join(`/go/bin`, c.name)+`"]
`)

	return dockerfile.Bytes()
}
