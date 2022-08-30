FROM gcr.io/distroless/static@sha256:1f580b0a1922c3e54ae15b0758b5747b260bd99d39d40c2edb3e7f6e2452298b

ARG TARGETOS
ARG TARGETARCH

ADD bin/$TARGETOS\_$TARGETARCH/provider /usr/local/bin/crossplane-aws-provider

USER 65532
ENTRYPOINT ["crossplane-aws-provider"]
