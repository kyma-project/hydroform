package runtimes

const Python38Dockerfile = `FROM eu.gcr.io/kyma-project/function-runtime-python38:cc7dd53f
USER root
ENV KUBELESS_INSTALL_VOLUME=/kubeless

COPY /src/requirements.txt $KUBELESS_INSTALL_VOLUME/requirements.txt
RUN pip install -r $KUBELESS_INSTALL_VOLUME/requirements.txt
COPY /src $KUBELESS_INSTALL_VOLUME
USER 1000
`

//const (
//	Python38Path          = "PYTHONPATH=$(KUBELESS_INSTALL_VOLUME)/lib.python3.8/site-packages:$(KUBELESS_INSTALL_VOLUME)"
//	Python38DebugEndpoint = `5678`
//)
