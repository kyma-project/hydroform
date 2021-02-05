package runtimes

const Nodejs10Dockerfile = `FROM eu.gcr.io/kyma-project/function-runtime-nodejs10:cc7dd53f
USER root
ENV KUBELESS_INSTALL_VOLUME=/kubeless

COPY /src/package.json $KUBELESS_INSTALL_VOLUME/package.json
RUN /kubeless-npm-install.sh
COPY /src $KUBELESS_INSTALL_VOLUME
USER 1000
`

const (
	Nodejs10Path          = "NODE_PATH=$(KUBELESS_INSTALL_VOLUME)/node_modules"
	Nodejs10DebugOption   = "--inspect=0.0.0.0"
	Nodejs10DebugEndpoint = `9229`
)